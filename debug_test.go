package httpdebug

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kisom/httpdebug/internal/debug"
	"github.com/kisom/whitelist"
)

func testEndpoint(url string, expected int) error {
	resp, err := http.Get(url)
	resp.Body.Close()
	if err != nil {
		return err
	}

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

// TestUninitErrors verifies failures that should occur before calling
// New.
func TestUninitErrors(t *testing.T) {
	err := setup()
	if err == nil {
		t.Fatal("httpdebug: setup should have failed")
	}

	_, err = Handler()
	if err == nil {
		t.Fatal("httpdebug: Handler should have failed")
	}

	_, err = HandlerFunc()
	if err == nil {
		t.Fatal("httpdebug: HandlerFunc should have failed")
	}

	mux := http.NewServeMux()
	err = Register(mux)
	if err == nil {
		t.Fatal("httpdebug: Register should have failed")
	}
}

// TestDoubleRegistration verifies that calling Register will not
// return an error. This also verifies that hitting the endpoints
// before setup will fail.
func TestDoubleSetup(t *testing.T) {
	err := NewLocalhost(0, false, false)
	if err != nil {
		t.Fatalf("%s", err)
	}

	err = Setup()
	if err != nil {
		t.Fatalf("%s", err)
	}

	err = Setup()
	if err != nil {
		t.Fatalf("%s", err)
	}

}

// TestDoubleNew verifies that calling New twice should cause an error.
func TestDoubleNew(t *testing.T) {
	err := NewLocalhost(0, true, true)
	if err != errAlreadyInit {
		t.Fatalf("httpdebug: NewLocalhost should have returned '%s' when called twice, but got err=%s", errAlreadyInit, err)
	}

	err = New(nil, nil, 0, false, false)
	if err != errAlreadyInit {
		t.Fatalf("httpdebug: NewLocalhost should have returned '%s' when called twice, but got err=%s", errAlreadyInit, err)
	}
}

// TestNewLocalhost runs sanity checks against a new localhost-based
// debug.
func TestNewLocalhost(t *testing.T) {
	HandleFunc("/debug/ok1", okDebugResponse)

	handler, err := Handler()
	if err != nil {
		t.Fatalf("%s", err)
	}

	srv := httptest.NewServer(handler)
	defer srv.Close()

	err = testEndpoint(srv.URL+"/debug/ok1", http.StatusOK)
	if err != nil {
		t.Fatalf("%s", err)
	}
}

func testLoggingAuth(req *http.Request) bool {
	log.Println("received request from", req.RemoteAddr)
	return false
}

// TestNewOpen runs basic checks on an open debug.
func TestNewOpen(t *testing.T) {
	Handle("/debug/ok2", http.HandlerFunc(okDebugResponse))

	LocalAdmin()

	handler, err := Handler()
	if err != nil {
		t.Fatalf("%s", err)
	}

	srv := httptest.NewServer(handler)
	defer srv.Close()

	err = testEndpoint(srv.URL+"/debug/ok1", http.StatusOK)
	if err != nil {
		t.Fatalf("%s", err)
	}
}

// TestWhitelisting makes sure that only whitelisted requests pass.
func TestWhitelisting(t *testing.T) {
	acl := whitelist.NewBasic()
	acl.Add(net.ParseIP("1.2.3.4"))

	lock.Lock()
	debugger = nil
	lock.Unlock()

	err := New(acl, debug.DefaultAdminAuth, 0, false, false)
	if err != nil {
		t.Fatalf("%s", err)
	}

	Handle("/debug/ok1", http.HandlerFunc(okDebugResponse))
	SetAdmin(debug.DefaultAdminAuth)
	LocalAdmin()
	SetAdminACL(acl)

	handlerFunc, err := HandlerFunc()
	if err != nil {
		t.Fatalf("%s", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(handlerFunc))
	defer srv.Close()

	err = testEndpoint(srv.URL+"/debug/ok1", http.StatusForbidden)
	if err != nil {
		t.Fatalf("%s", err)
	}
}

func TestRegister(t *testing.T) {
	mux := http.NewServeMux()
	err := Register(mux)
	if err != nil {
		t.Fatalf("%s", err)
	}

	err = Register(nil)
	if err != nil {
		t.Fatalf("%s", err)
	}
}
