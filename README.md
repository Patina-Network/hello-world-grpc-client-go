# hello-world-grpc-client-go

Hello World gRPC Go client for Patina Network

The UI is a single `frontend/index.html` (Tailwind from a CDN, no build step), compiled into the server binary with `go:embed` (`frontend/embed.go`).

## Prerequisites

- Go 1.27 or newer [(brew.sh)](https://formulae.brew.sh/formula/go)
- just [(brew.sh)](https://formulae.brew.sh/formula/just)
- direnv, then run `direnv allow` to load `.envrc` [(brew.sh)](https://formulae.brew.sh/formula/direnv)
- Tailscale, connected to the Patina VPN [(brew.sh)](https://formulae.brew.sh/cask/tailscale-app)
- Docker, only for `just docker-build` [(brew.sh)](https://formulae.brew.sh/cask/docker-desktop)

The Patina VPN is required to download the `hello-world-grpc-service` module from the private Go proxy (`GOPROXY` in `.envrc`) and to reach the staging gRPC service.

Docker builds download modules directly using the same proxy settings, configured in the build stage. The Docker builder must also be able to reach the private proxy over the VPN; vendoring is not required.

> [!NOTE]
> You must be connected to the VPN to connect locally. You can find the instructions to connect at <https://docs.patinanetwork.org/infra/how-to-connect-to-vpn/>

## Development

```sh
just run            # run the backend + UI on :8081 against the staging gRPC service
just test           # unit tests with the race detector
just lint           # gofmt check + go vet
just fmt            # apply formatting
just docker-build
```

## Environment variables

| Variable                              | Default                          | Description                                                                        |
| ------------------------------------- | -------------------------------- | ---------------------------------------------------------------------------------- |
| `HELLO_WORLD_SERVICE_GRPC_HOST`       | `hello-world-grpc-service:50051` | `host:port` of the hello-world gRPC service                                        |
| `HELLO_WORLD_SERVICE_GRPC_TLS`        | `false`                          | Connect to the gRPC service over TLS                                               |
| `HELLO_WORLD_SERVICE_GRPC_TIMEOUT_MS` | `3000`                           | Deadline for each gRPC call, in milliseconds                                       |
| `HTTP_PORT`                           | `8080`                           | Port the HTTP server listens on (always binds `0.0.0.0`)                           |
| `VERSION`                             | `N/A`                            | Version string returned by `GET /version`                                          |
| `HELLO_WORLD_CLIENT_URLS`             | none                             | Comma-separated URLs returned by `GET /urls` and listed in the UI                  |
| `ENVIRONMENT`                         | none                             | `production` or `staging` switches logs to JSON; anything else logs human-readable |

`just run` sets `HELLO_WORLD_SERVICE_GRPC_HOST` to the staging service, `HELLO_WORLD_SERVICE_GRPC_TLS=true`, `HTTP_PORT=8081`, and `HELLO_WORLD_CLIENT_URLS` to the other local clients.

## Releases

On `main`, CI creates an unprefixed semantic-version tag after both image builds succeed, using GitHub App credentials. Tags start at `1.0.0` and increment the patch version; source package versions are unchanged.

The tag triggers CD. Because the tagging step creates a new commit, CD promotes images built from its parent commit to the release tag and `latest`, then deploys production using the release tag. CI deploys only staging; production deployment runs only in CD.

Staging and production Kubernetes deployments use `patinanetwork/hello-world-client-go-arm`. Manifest directories remain `base/<environment>/hello-world-client-go`.
