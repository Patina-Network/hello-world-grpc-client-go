set shell := ["bash", "-euo", "pipefail", "-c"]

default:
    @just --list

run *args:
    HELLO_WORLD_SERVICE_GRPC_HOST=stg.hello-world-grpc-service.vpn.patinanetwork.org:50051 \
    HELLO_WORLD_SERVICE_GRPC_TLS=true \
    HTTP_PORT=8081 \
    HELLO_WORLD_CLIENT_URLS=http://localhost:8080,http://localhost:8082 \
    go run ./cmd/server {{ args }}

test *args:
    go test -race -count=1 ./... {{ args }}

lint:
    test -z "$(gofmt -l cmd internal frontend/embed.go)" || (gofmt -l cmd internal frontend/embed.go && exit 1)
    go vet ./...

fmt:
    gofmt -w cmd internal frontend/embed.go

docker-build tag="hello-world-grpc-client-go":
    docker build -t {{ tag }} .
