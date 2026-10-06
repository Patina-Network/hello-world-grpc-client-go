package main

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	helloworld "patinanetwork.org/grpc/hello-world-grpc-service"

	"github.com/Patina-Network/hello-world-grpc-client-go/frontend"
	"github.com/Patina-Network/hello-world-grpc-client-go/internal/config"
	"github.com/Patina-Network/hello-world-grpc-client-go/internal/grpc/greeter"
	"github.com/Patina-Network/hello-world-grpc-client-go/internal/httpserve"
)

func main() {
	slog.SetDefault(slog.New(logHandler(os.Getenv("ENVIRONMENT"))))
	if err := config.Command(serve).Run(context.Background(), os.Args); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func logHandler(environment string) slog.Handler {
	switch environment {
	case "production", "staging":
		return slog.NewJSONHandler(os.Stdout, nil)
	default:
		return slog.NewTextHandler(os.Stdout, nil)
	}
}

func serve(ctx context.Context, cfg config.Config) error {
	creds := insecure.NewCredentials()
	if cfg.TLS {
		creds = credentials.NewTLS(&tls.Config{})
	}
	m := httpserve.NewMetrics()
	conn, err := grpc.NewClient(cfg.Host,
		grpc.WithTransportCredentials(creds),
		grpc.WithUnaryInterceptor(m.UnaryClientInterceptor()),
	)
	if err != nil {
		return err
	}
	defer conn.Close()

	greeterService := greeter.New(helloworld.NewGreeterServiceClient(conn), cfg.Timeout)

	routes := slices.Concat(
		httpserve.SysRoutes(&cfg, m.Handler()),
		httpserve.HelloRoutes(greeterService),
		httpserve.GreetingRoutes(greeterService),
	)

	srv := &http.Server{
		Addr: net.JoinHostPort("0.0.0.0", strconv.Itoa(cfg.HTTPPort)),
		Handler: httpserve.NewHandler(httpserve.HandlerOptions{
			Static: frontend.FS(),
			Routes: routes,
		}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      35 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errc := make(chan error, 1)
	go func() {
		errc <- srv.ListenAndServe()
	}()

	slog.Info("HTTP server started", "address", srv.Addr)

	select {
	case err := <-errc:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
