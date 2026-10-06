package httpserve

import (
	"encoding/json/v2"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Patina-Network/hello-world-grpc-client-go/internal/grpc/greeter"
)

func writeServiceError(w http.ResponseWriter, svcName string, err error) {
	slog.Warn(svcName+" call failed", "error", err)
	code, message := httpError(err)
	writeError(w, code, message)
}

type errorBody struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, errorBody{Error: message})
}

func writeJSON(w http.ResponseWriter, code int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.MarshalWrite(w, value)
}

func httpError(err error) (int, string) {
	switch {
	case errors.Is(err, greeter.ErrInvalidArgument):
		return http.StatusBadRequest, "invalid request"
	case errors.Is(err, greeter.ErrNotFound):
		return http.StatusNotFound, "not found"
	case errors.Is(err, greeter.ErrAlreadyExists):
		return http.StatusConflict, "already exists"
	case errors.Is(err, greeter.ErrUnauthenticated):
		return http.StatusUnauthorized, "authentication required"
	case errors.Is(err, greeter.ErrPermissionDenied):
		return http.StatusForbidden, "permission denied"
	case errors.Is(err, greeter.ErrResourceExhausted):
		return http.StatusTooManyRequests, "resource exhausted"
	case errors.Is(err, greeter.ErrUnavailable):
		return http.StatusServiceUnavailable, "service unavailable"
	case errors.Is(err, greeter.ErrTimeout):
		return http.StatusGatewayTimeout, "upstream timeout"
	default:
		return http.StatusBadGateway, "upstream request failed"
	}
}
