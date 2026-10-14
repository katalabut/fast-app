package service

import (
	"net"
	"net/http"
	"net/http/pprof"
	"strconv"
	"strings"
)

func httpAddress(host string, port int) string {
	return net.JoinHostPort(host, strconv.Itoa(port))
}

// Register only the intended pprof handlers, never arbitrary DefaultServeMux routes.
func registerPprof(mux *http.ServeMux, prefix string) {
	prefix = strings.TrimSuffix(prefix, "/") + "/pprof/"
	mux.HandleFunc(prefix, func(w http.ResponseWriter, r *http.Request) {
		request := r.Clone(r.Context())
		request.URL.Path = "/debug/pprof/" + strings.TrimPrefix(r.URL.Path, prefix)
		request.URL.RawPath = ""
		pprof.Index(w, request)
	})
	mux.HandleFunc(prefix+"cmdline", pprof.Cmdline)
	mux.HandleFunc(prefix+"profile", pprof.Profile)
	mux.HandleFunc(prefix+"symbol", pprof.Symbol)
	mux.HandleFunc(prefix+"trace", pprof.Trace)
}
