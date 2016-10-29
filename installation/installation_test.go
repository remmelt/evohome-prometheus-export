package installation

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

const (
	accessToken = "bearer test-access-token"
)

func checkAuth(r *http.Request) bool {
	if r.Header.Get("Authorization") != accessToken {
		return true
	}
	return false
}

func checkQueryData(r *http.Request) bool {
	v := r.URL.Query()
	if v.Get("includeTemperatureControlSystems") == "True" && v.Get("userId") != "" {
		return true
	}
	return false
}

func testServer() *httptest.Server {
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if checkAuth(r) {
			if checkQueryData(r) {
				w.Header().Set("Content-Type", "application/json;charset=UTF-8")
				w.Header().Set("Cache-Control", "no-cache")
				w.Header().Set("Expires", "-1")
				w.Header().Set("Pragma", "no-cache")
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, responseData)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	return s
}
