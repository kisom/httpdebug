package trace

import "net/http"

func TraceRequest(w http.ResponseWriter, req *http.Request) {
	any, sensitive := AuthRequest(req)
	if !any {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	Render(w, req, sensitive)
}

func EventRequest(w http.ResponseWriter, req *http.Request) {
	any, sensitive := AuthRequest(req)
	if !any {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	RenderEvents(w, req, sensitive)
}
