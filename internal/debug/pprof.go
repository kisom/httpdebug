package debug

// pprof.go contains support for the pprof endpoints.

import (
	"net/http"

	"github.com/kisom/httpdebug/pprof"
)

var pprofEndpoints = map[string]func(http.ResponseWriter, *http.Request){
	"/debug/pprof":         pprof.Index,
	"/debug/pprof/":        pprof.Index,
	"/debug/pprof/*":       pprof.Index,
	"/debug/pprof/cmdline": pprof.Cmdline,
	"/debug/pprof/profile": pprof.Profile,
	"/debug/pprof/symbol":  pprof.Symbol,
	"/debug/pprof/trace":   pprof.Trace,
}

// pprofSetup applies any ACL and timeout constraints on the pprof
// endpoints, and adds them to list of endpoints to be registered.
func (d *Debug) pprofSetup() {
	if d.setup {
		return
	}

	for pat, h := range pprofEndpoints {
		d.endpoints[pat] = d.setupHandler(h)
	}
}
