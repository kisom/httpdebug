package main

import (
	"flag"
	"log"
	"net/http"
	"time"

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
