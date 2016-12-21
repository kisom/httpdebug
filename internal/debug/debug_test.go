package debug

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"runtime/pprof"
	"testing"
	"time"

	// The use of the other whitelist package is intended: it
	// verifies compatibility with the other package.
	"github.com/kisom/whitelist"
)

func testEndpoint(url string, expected int) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode != expected {
		return fmt.Errorf("debug: expected %s to return %d ('%s'), but got %d ('%s')",
			resp.Request.URL.Path, expected, http.StatusText(expected),
			resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return nil
}

func okDebugResponse(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok\n"))
}

// TestDoubleRegistration verifies that calling Register twice doesn't
// cause problems. This was more of an issue when the Register
// function return an error, but it's a good regression test.
func TestDoubleRegistration(t *testing.T) {
	debug := NewLocalhost(0, true, true)
	debug.Register()
	debug.Register()
}

// TestNewLocalhost runs sanity checks against a new localhost-based
// debug.
func TestNewLocalhost(t *testing.T) {
	debug := NewLocalhost(0, false, false)

	srv := httptest.NewServer(debug)
	err := testEndpoint(srv.URL+"/debug/pprof", http.StatusInternalServerError)
	srv.Close()
	if err != nil {
		t.Fatalf("%s", err)
	}

	debug.Register()
	debug.HandleFunc("/debug/ok1", okDebugResponse)
	srv = httptest.NewServer(debug)
	defer srv.Close()

	err = testEndpoint(srv.URL+"/debug/ok1", http.StatusOK)
	if err != nil {
		t.Fatalf("%s", err)
	}
}

// TestNilHandlerACL ensures that we'll catch the case where something
// passes a nil http.Handler to aclHandler.
func TestNilHandlerACL(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("debug: ACL handler setup should have panicked")
		}
	}()

	debug := NewLocalhost(0, false, false)

	// This should panic.
	_ = debug.aclHandler(nil)
}

func testLoggingAuth(req *http.Request) bool {
	log.Println("received request from", req.RemoteAddr)
	return false
}

// TestNewOpen runs basic checks on an open debug.
func TestNewOpen(t *testing.T) {
	debug := New(nil, DefaultAdminAuth, time.Second, false, false)
	debug.Register()

	debug.Handle("/debug/ok1", http.HandlerFunc(okDebugResponse))
	debug.SetAdmin(testLoggingAuth)
	debug.LocalAdmin()
	srv := httptest.NewServer(debug)
	defer srv.Close()

	err := testEndpoint(srv.URL+"/debug/ok1", http.StatusOK)
	if err != nil {
		t.Fatalf("%s", err)
	}
}

// TestWhitelisting makes sure that only whitelisted requests pass.
func TestWhitelisting(t *testing.T) {
	acl := whitelist.NewBasic()
	acl.Add(net.ParseIP("1.2.3.4"))

	debug := New(acl, DefaultAdminAuth, time.Second, false, false)
	debug.Register()

	debug.Handle("/debug/ok1", http.HandlerFunc(okDebugResponse))
	debug.SetAdmin(testLoggingAuth)
	debug.LocalAdmin()
	srv := httptest.NewServer(debug)
	defer srv.Close()

	err := testEndpoint(srv.URL+"/debug/ok1", http.StatusForbidden)
	if err != nil {
		t.Fatalf("%s", err)
	}
}

func adminDebugRequest(debug *Debug) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if debug.admin(r) {
			w.Write([]byte("ok"))
			return
		}
		forbidden(w, r)
	}
}

func TestAdminAuth(t *testing.T) {
	debug := NewLocalhost(0, false, false)
	debug.SetAdmin(func(r *http.Request) bool {
		return false
	})
	debug.Register()

	debug.HandleFunc("/debug/admin", adminDebugRequest(debug))

	srv := httptest.NewServer(debug)
	defer srv.Close()
	err := testEndpoint(srv.URL+"/debug/admin", http.StatusForbidden)
	if err != nil {
		t.Fatalf("%s", err)
	}

	debug.SetAdmin(func(r *http.Request) bool {
		return true
	})
	err = testEndpoint(srv.URL+"/debug/admin", http.StatusOK)
	if err != nil {
		t.Fatalf("%s", err)
	}

	acl := whitelist.NewBasic()
	acl.Add(net.ParseIP("1.2.3.4"))
	debug.SetAdminACL(acl)
	err = testEndpoint(srv.URL+"/debug/admin", http.StatusForbidden)
	if err != nil {
		t.Fatalf("%s", err)
	}

	acl = whitelist.NewBasic()
	acl.Add(net.ParseIP("127.0.0.1"))
	acl.Add(net.ParseIP("::1"))
	debug.SetAdminACL(acl)
	err = testEndpoint(srv.URL+"/debug/admin", http.StatusOK)
	if err != nil {
		t.Fatalf("%s", err)
	}

	currentAdmin := debug.admin
	// Gratuitous failure mode testing.
	debug.SetAdmin(func(r *http.Request) bool {
		return currentAdmin(nil)
	})
	err = testEndpoint(srv.URL+"/debug/admin", http.StatusForbidden)
	if err != nil {
		t.Fatalf("%s", err)
	}
}

func TestAddProfile(t *testing.T) {
	debug := NewLocalhost(0, true, false)
	debug.Register()

	p := pprof.NewProfile("pkg/debug")
	debug.AddProfile("pkg/debug")
	p.Add("first dump", 0)

	srv := httptest.NewServer(debug)
	defer srv.Close()

	err := testEndpoint(srv.URL+"/debug/pprof/pkg/debug", http.StatusOK)
	if err != nil {
		t.Fatalf("%s", err)
	}
}
