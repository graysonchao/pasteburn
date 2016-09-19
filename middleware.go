package pasteburn

import "net/http"

// CorsMiddleware sets CORS headers on responses.
func CorsMiddleware(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
