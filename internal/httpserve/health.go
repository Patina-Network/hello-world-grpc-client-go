package httpserve

import "net/http"

// k8s will use this to determine instability.
// If it returns a non-200, it will stop sending requests
// to the pod until it has recovered.
func ready(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("ok"))
}

// k8s will use this to determine non-recoverability.
// If it returns a non-200, it will kill the pod & restart it.
//
// Avoid making any external service calls here. Those should go in `/livez`.
func live(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("ok"))
}
