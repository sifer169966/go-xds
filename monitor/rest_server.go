package monitor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"

	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sifer169966/go-xds/configs"
)

type RESTServer struct {
	http.Server
	mux      *http.ServeMux
	muxCache *cachev3.MuxCache
}

func NewREST(muxCache *cachev3.MuxCache, cfg configs.MonitorServer) *RESTServer {
	mux := http.NewServeMux()
	out := &RESTServer{
		mux: mux,
		Server: http.Server{
			Addr:              fmt.Sprintf(":%s", cfg.Port),
			Handler:           mux,
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		},
		muxCache: muxCache,
	}

	out.resgisterRoutes()
	return out
}

func (s *RESTServer) resgisterRoutes() {
	s.mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	s.mux.HandleFunc("/", s.retrieveSnapshotInfo)

	s.mux.Handle("/metrics", promhttp.Handler())

	s.mux.HandleFunc("/debug/pprof/", pprof.Index)
	s.mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	s.mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	s.mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	s.mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

}

func (s *RESTServer) retrieveSnapshotInfo(w http.ResponseWriter, _ *http.Request) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	out := map[string]cacheMarshaler{}
	for k, v := range s.muxCache.Caches {
		out[k] = cacheMarshaler{Cache: v}
	}
	w.Header().Set("Content-Type", "application/json")
	enc.Encode(out)
}
