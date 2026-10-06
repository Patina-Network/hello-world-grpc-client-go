package config

import (
	"context"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
)

type Config struct {
	Host     string
	TLS      bool
	Timeout  time.Duration
	HTTPPort int
	Version  string
	URLs     []string
}

func Command(run func(context.Context, Config) error) *cli.Command {
	return &cli.Command{
		Name: "hello-world-grpc-client-go",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "hello-world-service-grpc-host", Value: "hello-world-grpc-service:50051", Sources: cli.EnvVars("HELLO_WORLD_SERVICE_GRPC_HOST")},
			&cli.BoolFlag{Name: "hello-world-service-grpc-tls", Sources: cli.EnvVars("HELLO_WORLD_SERVICE_GRPC_TLS")},
			&cli.IntFlag{Name: "hello-world-service-grpc-timeout-ms", Value: 3000, Sources: cli.EnvVars("HELLO_WORLD_SERVICE_GRPC_TIMEOUT_MS")},
			&cli.IntFlag{Name: "http-port", Value: 8080, Sources: cli.EnvVars("HTTP_PORT")},
			&cli.StringFlag{Name: "version", Value: "N/A", Sources: cli.EnvVars("VERSION")},
			&cli.StringFlag{Name: "hello-world-client-urls", Sources: cli.EnvVars("HELLO_WORLD_CLIENT_URLS")},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return run(ctx, Config{
				Host:     cmd.String("hello-world-service-grpc-host"),
				TLS:      cmd.Bool("hello-world-service-grpc-tls"),
				Timeout:  time.Duration(cmd.Int("hello-world-service-grpc-timeout-ms")) * time.Millisecond,
				HTTPPort: cmd.Int("http-port"),
				Version:  cmd.String("version"),
				URLs:     splitURLs(cmd.String("hello-world-client-urls")),
			})
		},
	}
}

func splitURLs(list string) []string {
	urls := []string{}
	for u := range strings.SplitSeq(list, ",") {
		if u = strings.TrimSpace(u); u != "" {
			urls = append(urls, u)
		}
	}
	return urls
}
