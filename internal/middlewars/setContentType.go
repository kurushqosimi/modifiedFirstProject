package middlewars

import "net/http"

func SetContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		next.ServeHTTP(response, request)
		response.Header().Set("Custom Header", "Kurush")
	})
}
