// Package httpdebug implements debugging endpoints, incorporating
// net/http/pprof and x/net/trace. It supports whitelisting endpoints,
// and presents itself as an http.Handler. It was motivated by a
// frustration with the default imports from the two named packages,
// which make it difficult to limit requests only to localhost.
//
// One of the New functions must be called before any of the other
// functions; they may be called only once. This is due to a
// limitation in the trace package, which operates on a global level.
//
// Note that using this package should only be done *instead* of using
// the previously mentioned packages.
package httpdebug

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/kisom/httpdebug/internal/debug"
	"github.com/kisom/httpdebug/whitelist"
)

var (
	// debugger contains the internal, global debugger.
	debugger *debug.Debug

	// lock is used to synchronise multiple calls to certain functions.
	lock = &sync.Mutex{}

	// errAlreadyInit is used for testing.
	errAlreadyInit = errors.New("httpdebug: already initialised")
)

// NewLocalhost sets up the debug handler restricted to localhost. If
// timeout is 0, no timeouts will be applied. If pprofDisable is true,
// the pprof endpoints will not be enabled. If traceDisable is true,
// the trace endpoints will not be enabled.
func NewLocalhost(timeout time.Duration, pprofDisable, traceDisable bool) error {
	lock.Lock()
	defer lock.Unlock()

	if debugger != nil {
		return errAlreadyInit
	}

	debugger = debug.NewLocalhost(timeout, pprofDisable, traceDisable)
	return nil
}

// New sets up the debug handler restricted with to the given whitelist. If
// timeout is 0, no timeouts will be applied. If pprofDisable is true,
// the pprof endpoints will not be enabled. If traceDisable is true,
// the trace endpoints will not be enabled.
func New(acl whitelist.ACL, admin func(*http.Request) bool, timeout time.Duration, pprofDisable, traceDisable bool) error {
	lock.Lock()
	defer lock.Unlock()

	if debugger != nil {
		return errAlreadyInit
	}

	debugger = debug.New(acl, admin, timeout, pprofDisable, traceDisable)
	return nil
}

var errNotSetup = errors.New("httpdebug: the debug handler has not been set up (use one of the New functions)")

// setup verifies that the debug handler has been instantiated
// (e.g. via NewLocalhost) and sets up the handler.
func setup() (err error) {
	if debugger == nil {
		return errNotSetup
	}

	debugger.Register()
	return nil
}

// Setup sets up the debug handler. Register may be called multiple
// times; after the first call, it will have no effect.
func Setup() (err error) {
	lock.Lock()
	defer lock.Unlock()

	return setup()
}

// Handler does any initial setup on the debug handler and returns an
// HTTP handler for the debug endpoints.
func Handler() (http.Handler, error) {
	lock.Lock()
	defer lock.Unlock()

	err := setup()
	if err != nil {
		return nil, err
	}

	return debugger, nil
}

// HandlerFunc does any initial setup on the debug handler and returns an
// HTTP handler func for the debug endpoints.
func HandlerFunc() (func(http.ResponseWriter, *http.Request), error) {
	lock.Lock()
	defer lock.Unlock()

	err := setup()
	if err != nil {
		return nil, err
	}

	return debugger.ServeHTTP, nil
}

// Register provides a convenience function for registering the debug
// endpoints with an existing *http.ServeMux. If mux is nil, the
// http.Handle functions are used.
func Register(mux *http.ServeMux) error {
	lock.Lock()
	defer lock.Unlock()

	err := setup()
	if err != nil {
		return err
	}

	if mux == nil {
		http.Handle("/debug/", debugger)
	} else {
		mux.Handle("/debug/", debugger)
	}

	return nil
}

// SetAdminACL allows an ACL to be applied to the debug handler.
func SetAdminACL(acl whitelist.ACL) {
	lock.Lock()
	defer lock.Unlock()

	debugger.SetAdminACL(acl)
}

// SetAdmin sets the admin authenticator.
func SetAdmin(auth func(req *http.Request) bool) {
	lock.Lock()
	defer lock.Unlock()

	debugger.SetAdmin(auth)
}

// LocalAdmin permits viewing sensitive traces from localhost.
func LocalAdmin() {
	lock.Lock()
	defer lock.Unlock()

	debugger.LocalAdmin()
}

// Handle registers a new handler. It is intended to allow additional
// debugging tools to be enabled. One of the New functions must have
// been called already.
func Handle(pat string, h http.Handler) {
	lock.Lock()
	defer lock.Unlock()

	debugger.Handle(pat, h)
}

// HandleFunc registers a new handler function. It is intended to
// allow additional debugging tools to be enabled. One of the New
// functions must have been called already.
func HandleFunc(pat string, f func(http.ResponseWriter, *http.Request)) {
	lock.Lock()
	defer lock.Unlock()

	debugger.HandleFunc(pat, f)
}

// AddProfile registers a new profile endpoint for pprof under
// /debug/profile/name. The profile must already have been created
// using the runtime/pprof package.
func AddProfile(name string) {
	lock.Lock()
	defer lock.Unlock()

	debugger.AddProfile(name)
}
