// Package debug implements the internal Debug type used in httpdebug.
package debug

import (
	"net"
	"net/http"
	"time"

	"github.com/kisom/httpdebug/whitelist"
)

// forbidden returns a standard 403 - Forbidden response.
func forbidden(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
}

// DefaultAdminAuth is a default admin authenticator that is applied
// to new Debug values that returns the value of AllowSensitiveTrace.
func DefaultAdminAuth(req *http.Request) bool {
	return AllowSensitiveTrace
}

// Debug is an http.Handler providing debugging endpoints under /debug.
type Debug struct {
	acl       whitelist.ACL            // A whitelist for any requests.
	admin     func(*http.Request) bool // Authenticator for sensitive trace requests.
	timeo     time.Duration            // A timeout that should be setup for any requests.
	enpprof   bool                     // Enable pprof endpoints.
	entrace   bool                     // Enable trace endpoints.
	setup     bool                     // Has the Debug been setup?
	mux       *http.ServeMux           // mux provides the underlying handler.
	endpoints map[string]http.Handler  // Endpoints that will be registered.
}

// NewLocalhost returns a new Debug restricted to localhost. If
// timeout is 0, no timeouts will be applied. If pprofDisable is true,
// the pprof endpoints will not be enabled. If traceDisable is true,
// the trace endpoints will not be enabled.
func NewLocalhost(timeout time.Duration, pprofDisable, traceDisable bool) *Debug {
	localhost := whitelist.NewBasic()
	localhost.Add(net.ParseIP("127.0.0.1"))
	localhost.Add(net.ParseIP("::1"))
	return &Debug{
		acl:       localhost,
		admin:     DefaultAdminAuth,
		timeo:     timeout,
		enpprof:   !pprofDisable,
		entrace:   !traceDisable,
		mux:       http.NewServeMux(),
		endpoints: map[string]http.Handler{},
	}
}

// New returns a new Debug restricted with to the given whitelist. If
// timeout is 0, no timeouts will be applied. If pprofDisable is true,
// the pprof endpoints will not be enabled. If traceDisable is true,
// the trace endpoints will not be enabled.
func New(acl whitelist.ACL, admin func(*http.Request) bool, timeout time.Duration, pprofDisable, traceDisable bool) *Debug {
	return &Debug{
		acl:       acl,
		admin:     DefaultAdminAuth,
		timeo:     timeout,
		enpprof:   !pprofDisable,
		entrace:   !traceDisable,
		mux:       http.NewServeMux(),
		endpoints: map[string]http.Handler{},
	}
}

// aclHandler will apply the ACL to the endpoint.
func (d *Debug) aclHandler(h http.Handler) http.Handler {
	if d.acl == nil {
		return h
	}

	var err error
	h, err = whitelist.NewHandler(h, http.HandlerFunc(forbidden), d.acl)
	if err != nil {
		// whitelist.NewHandler only returns an error if
		// either the first or third arguments are nil.
		panic("debug: whitelist.NewHandler should never error")
	}
	return h
}

// aclEndpoint will apply the ACL to the endpoint.
func (d *Debug) aclHandlerFunc(f func(http.ResponseWriter, *http.Request)) http.Handler {
	return d.aclHandler(http.HandlerFunc(f))
}

// timeout applies a timeout handler to the handler.
func (d *Debug) timeout(h http.Handler) http.Handler {
	if d.timeo == 0 {
		return h
	}

	return http.TimeoutHandler(h, d.timeo, http.StatusText(http.StatusRequestTimeout))
}

// setupHandler applies any ACL and timeout handlers to the given http.Handler.
func (d *Debug) setupHandler(f func(http.ResponseWriter, *http.Request)) http.Handler {
	handler := d.aclHandlerFunc(f)
	return d.timeout(handler)
}

// Register sets up the internal multiplexer.
func (d *Debug) Register() {
	if d.setup {
		return
	}

	if d.enpprof {
		d.pprofSetup()
	}

	if d.entrace {
		d.traceSetup()
	}

	for pat, h := range d.endpoints {
		d.mux.Handle(pat, h)
	}

	d.setup = true
}

// ServeHTTP multiplexes the registered debug endpoints.
func (d *Debug) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !d.setup {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	d.mux.ServeHTTP(w, r)
}

// SetAdminACL allows an ACL to be applied to the Debug.
func (d *Debug) SetAdminACL(acl whitelist.ACL) {
	d.admin = func(req *http.Request) bool {
		reqIP, err := whitelist.HTTPRequestLookup(req)
		if err != nil {
			return false
		}
		return acl.Permitted(reqIP)
	}
}

// SetAdmin sets the admin authenticator.
func (d *Debug) SetAdmin(auth func(req *http.Request) bool) {
	d.admin = auth
}

// LocalAdmin permits viewing sensitive traces from localhost.
func (d *Debug) LocalAdmin() {
	localhost := whitelist.NewBasic()
	localhost.Add(net.ParseIP("127.0.0.1"))
	localhost.Add(net.ParseIP("::1"))
	d.SetAdminACL(localhost)
}

// Handle registers a new handler.
func (d *Debug) Handle(pat string, h http.Handler) {
	handler := d.aclHandler(h)
	d.mux.Handle(pat, d.timeout(handler))
}

// HandleFunc registers a new handler function.
func (d *Debug) HandleFunc(pat string, f func(http.ResponseWriter, *http.Request)) {
	h := d.setupHandler(f)
	d.mux.Handle(pat, h)
}
