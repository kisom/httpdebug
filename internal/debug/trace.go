package debug

// trace.go contains support for the tracing endpoints.

import (
	"net/http"

	"github.com/kisom/httpdebug/trace"
	"github.com/kisom/httpdebug/whitelist"
)

// AllowSensitiveTrace controls whether sensitive traces are permitted
// when no ACL is provided. If false, and if the Debug struct is set
// up with no ACLs, no sensitive traces are permitted.
var AllowSensitiveTrace bool

// aclAuthRequest sets up the trace package to use the Debug's ACL.
func (d *Debug) aclAuthRequest() {
	if d.acl == nil {
		trace.AuthRequest = func(req *http.Request) (any, sensitive bool) {
			return true, d.admin(req)
		}
	}

	trace.AuthRequest = func(req *http.Request) (any, sensitive bool) {
		reqIP, err := whitelist.HTTPRequestLookup(req)
		if err != nil {
			return false, false
		}

		return d.acl.Permitted(reqIP), d.admin(req)
	}
}

var traceEndpoints = map[string]func(http.ResponseWriter, *http.Request){
	"/debug/requests": trace.TraceRequest,
	"/debug/events":   trace.EventRequest,
}

// traceSetup applies any ACL and timeout constraints to the trace
// handlers, and adds them to list of endpoints to be registered.
func (d *Debug) traceSetup() {
	if d.setup {
		return
	}

	d.aclAuthRequest()

	for pat, h := range traceEndpoints {
		d.endpoints[pat] = d.setupHandler(h)
	}
}
