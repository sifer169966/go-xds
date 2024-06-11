package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	xds "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/sifer169966/go-xds/callbacks"
	"github.com/sifer169966/go-xds/configs"
	"github.com/sifer169966/go-xds/k8sreflector"
	"github.com/sifer169966/go-xds/metrics"
	"github.com/sifer169966/go-xds/monitor"
	"github.com/sifer169966/go-xds/reflector"
	"github.com/sifer169966/go-xds/snapshots"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func main() {
	var cfg configs.Config
	configs.ReadENV(&cfg)
	klog.InitFlags(nil)
	defer klog.Flush()
	flag.Parse()

	err := metrics.SetPromGlobalMeter()
	if err != nil {
		klog.Fatal("could not set prom global meter", "err", err)
	}

	k8sClientConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(clientcmd.NewDefaultClientConfigLoadingRules(), nil).ClientConfig()
	if err != nil {
		klog.Fatal("could not create k8s client configuration", "err", err)
	}
	k8sClient, err := kubernetes.NewForConfig(k8sClientConfig)
	if err != nil {
		klog.Fatal("could not create k8s client", "err", err)
	}

	snap := snapshots.New()
	endpointReflector := k8sreflector.NewEndpointReflector(k8sClient, snap, k8sreflector.ReflectorConfig{})
	serviceReflector := k8sreflector.NewServiceReflector(k8sClient, snap, k8sreflector.ReflectorConfig{})

	stopCtx, stop := context.WithCancel(context.Background())

	stopCh := make(chan bool, 1)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := reflector.Start(stopCtx, endpointReflector, serviceReflector)
		if err != nil {
			klog.Error("error while running the reflector", "err", err)
		}
		stopCh <- true
	}()

	grpcServer := grpc.NewServer()
	healthServer := health.NewServer()
	monitorServer := monitor.NewREST(snap.MuxCache(), cfg.MonitorServer)
	xdsServer := xds.NewServer(stopCtx, snap.MuxCache(), callbacks.New())
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	discoverygrpc.RegisterAggregatedDiscoveryServiceServer(grpcServer, xdsServer)

	listenPort := fmt.Sprintf(":%s", cfg.App.GRPCPort)
	lis, err := net.Listen("tcp4", listenPort)
	if err != nil {
		klog.Fatal("could not listers on the server", "err", err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-stopCh:
			klog.Warning("got stop signal")
		case sig := <-sigCh:
			klog.Warning(fmt.Sprintln("got os signal: ", sig.String()))
		}
		klog.Info("server is shuting down...")
		stop()
		healthServer.Shutdown()
		grpcServer.GracefulStop()
		err := lis.Close()
		if err != nil {
			klog.Error("error while closing the listerner", "err", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := grpcServer.Serve(lis)
		if err != nil {
			klog.Error("error while closing the gRPC server", "err", err)
		}
	}()

	go monitorServer.ListenAndServe()
	klog.Info(fmt.Sprintln("starting server at port", listenPort))
	wg.Wait()
	klog.Info("application was closed")
}
