package callbacks

import (
	"context"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	xds "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/sifer169966/go-xds/metrics"
	otelmetric "go.opentelemetry.io/otel/metric"
	"k8s.io/klog/v2"
)

func New() xds.CallbackFuncs {
	meter := metrics.GetGlobalMeter()
	streamConnsGauge, _ := meter.Int64UpDownCounter("xds_server_stream_conns")
	deltaConnsGauge, _ := meter.Int64UpDownCounter("xds_server_delta_stream_conns")
	requestCounter, _ := meter.Int64Counter("xds_server_stream_requests")
	responseCounter, _ := meter.Int64Counter("xds_server_stream_responses")
	return xds.CallbackFuncs{
		StreamOpenFunc: func(ctx context.Context, streamID int64, typeURL string) error {
			streamConnsGauge.Add(ctx, 1)
			klog.Info("StreamOpen", "streamID", streamID, "typeURL", typeURL)
			return nil
		},
		StreamClosedFunc: func(streamID int64, node *corev3.Node) {
			streamConnsGauge.Add(context.Background(), -1)
			klog.Info("StreamClosed", "streamID", streamID)
		},
		DeltaStreamOpenFunc: func(ctx context.Context, streamID int64, typeURL string) error {
			deltaConnsGauge.Add(ctx, 1)
			klog.Info("DeltaStreamOpen", "streamID", streamID, "typeURL", typeURL)
			return nil
		},
		DeltaStreamClosedFunc: func(streamID int64, node *corev3.Node) {
			deltaConnsGauge.Add(context.Background(), -1)
			klog.Info("DeltaStreamClosed", "streamID", streamID)
		},
		StreamRequestFunc: func(streamID int64, request *discoverygrpc.DiscoveryRequest) error {
			requestCounter.Add(context.Background(), 1, otelmetric.WithAttributes(metrics.TypeURLAttrKey.String(request.GetTypeUrl())))
			klog.Info("StreamRequest", "streamID", streamID, "request", request)
			return nil
		},
		StreamResponseFunc: func(ctx context.Context, streamID int64, request *discoverygrpc.DiscoveryRequest, response *discoverygrpc.DiscoveryResponse) {
			responseCounter.Add(context.Background(), 1, otelmetric.WithAttributes(metrics.TypeURLAttrKey.String(request.GetTypeUrl())))
			klog.Info("StreamResponse", "streamID", streamID, "resourceNames", request.ResourceNames, "response", response)
		},
	}
}
