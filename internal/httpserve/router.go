package httpserve

import (
	"context"
	"io/fs"
	"net/http"

	"github.com/Patina-Network/hello-world-grpc-client-go/internal/config"
	"github.com/Patina-Network/hello-world-grpc-client-go/internal/grpc/greeter"
)

type Greeter interface {
	Echo(ctx context.Context, name string) (string, error)
	SendGreeting(ctx context.Context, g greeter.NewGreeting) error
	ListGreetings(ctx context.Context, recipientName *string) ([]greeter.Greeting, error)
}

type Route struct {
	Pattern string
	Handler http.Handler
}

type HandlerOptions struct {
	Static fs.FS
	Routes []Route
}

func NewHandler(opts HandlerOptions) http.Handler {
	mux := http.NewServeMux()
	for _, route := range opts.Routes {
		mux.Handle(route.Pattern, route.Handler)
	}
	mux.HandleFunc("/api/", func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusNotFound, "not found")
	})
	mux.Handle("/", http.FileServerFS(opts.Static))
	return mux
}

func SysRoutes(cfg *config.Config, metrics http.Handler) []Route {
	return []Route{
		{Pattern: "GET /readyz", Handler: http.HandlerFunc(ready)},
		{Pattern: "GET /livez", Handler: http.HandlerFunc(live)},
		{Pattern: "GET /metrics", Handler: metrics},
		{
			Pattern: "GET /version",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(cfg.Version))
			}),
		},
		{
			Pattern: "GET /urls",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				writeJSON(w, http.StatusOK, cfg.URLs)
			}),
		},
	}
}

func HelloRoutes(svc Greeter) []Route {
	h := &handlers{svc: svc}
	return []Route{
		{Pattern: "GET /api/echo", Handler: http.HandlerFunc(h.echo)},
	}
}

func GreetingRoutes(svc Greeter) []Route {
	h := &handlers{svc: svc}
	return []Route{
		{Pattern: "POST /api/greetings", Handler: http.HandlerFunc(h.sendGreeting)},
		{Pattern: "GET /api/greetings", Handler: http.HandlerFunc(h.listGreetings)},
	}
}
