httpdebug
=========

.. image:: https://travis-ci.org/kisom/httpdebug.svg?branch=master

.. image:: https://godoc.org/github.com/kisom/httpdebug?status.svg
   :target: https://godoc.org/github.com/kisom/httpdebug

This package provides debugging endpoints to which IP whitelisting and
timeout constrains can be added. It is designed to return an HTTP
handler that can be integrated into whatever HTTP server setup is
being used.

Motivation
----------

While I was working on adding debugging support to an existing project
(CFSSL_), I wanted to integrate the ``net/http/pprof`` package into
our existing server. However, an ACL needed to be added to restrict
viewing these endpoints to localhost. The ``pprof`` package registers
a number of handlers in its ``init`` function, and the ``net/http``
package doesn't allow overriding handled functions.

.. _CFSSL: https://github.com/cloudflare/cfssl

What I wanted was something that would enable us to access control
these endpoints while being able to provide a single ``http.Handler``
to present to my programs.

Versions
--------

The ``net/http/pprof`` package was copied over from the Go 1.7.4
branch, and modified to remove the ``init`` function.

The ``golang.org/x/net/trace{,/internal/timeseries}`` packages were
copied from the golang.org/x/net tree at Git commit
45e771701b814666a7eb299e6c7a57d0b1799e91.

Endpoints
---------

The pprof endpoints are

+ /debug/pprof
+ /debug/pprof/block
+ /debug/pprof/cmdline
+ /debug/pprof/goroutine
+ /debug/pprof/heap
+ /debug/pprof/profile
+ /debug/pprof/symbol
+ /debug/pprof/threadcreate
+ /debug/pprof/trace

The trace endpoints are

+ /debug/requests
+ /debug/events

Additional debugging endpoints can be added with the `Handle` and
`HandleFunc` packages. New handlers added here are wrapped in the
same ACL and timeout applied to all the other endpoints.

Usage
-----

The ``example/example.go`` package provides a good starting point.::

  // package main provides a quick demo of the httpdebug package.
  package main

  import (
          "flag"
          "log"
          "net/http"

          "github.com/kisom/httpdebug"
  )

  func index(w http.ResponseWriter, r *http.Request) {
          w.Write([]byte("Hello, world.\r\n"))
  }

  func main() {
          var addr string
          var pprofEnable, traceEnable bool
          var timeout time.Duration

          flag.StringVar(&addr, "a", "127.0.0.1:8080", "`address` to listen on")
          flag.BoolVar(&pprofEnable, "p", false, "enable pprof endpoints")
          flag.BoolVar(&traceEnable, "r", false, "enable request tracing")
          flag.DurationVar(&timeout, "t", 0, "`timeout` period for requests; 0 disables")
          flag.Parse()

          err := httpdebug.NewLocalhost(timeout, !pprofEnable, !traceEnable)
          if err != nil {
                  log.Fatal(err.Error())
          }

          http.HandleFunc("/", index)
          err = httpdebug.Register(nil)
          if err != nil {
                  log.Fatal(err.Error())
          }

          log.Println("listening on", addr)
          log.Fatal(http.ListenAndServe(addr, nil))
  }

Common Functions
----------------

The `godocs <https://godoc.org/github.com/kisom/httpdebug>`_ have a
complete list, but the following are the intended common functions.

- ``New``: set up the debug handler, providing finer control over the access control mechanism.
- ``NewLocalhost``: set up the debug handler with access restricted to
  localhost, and sensitive tracing access disabled.
- ``Handler`` returns an ``http.Handler`` for ``/debug``.
- ``HandlerFunc`` returns an ``http.HandlerFunc`` for ``/debug``.
- ``Register`` registers the ``http.Handler`` with an existing ``*http.ServeMux`` or the ``http.DefaultServeMux``.
