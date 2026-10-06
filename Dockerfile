FROM golang:1.27-trixie AS go-build

WORKDIR /src/app

ENV GOPROXY=https://pkg.vpn.patinanetwork.org/go/go,https://proxy.golang.org,direct \
  GONOSUMDB=patinanetwork.org

COPY go.mod go.sum ./
COPY cmd/ cmd/
COPY internal/ internal/
COPY frontend/ frontend/

RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server

FROM gcr.io/distroless/static-debian13:nonroot AS go

WORKDIR /app

COPY --from=go-build /out/server /app/server

ENV HTTP_PORT=8080
EXPOSE 8080
USER 65532:65532

ENTRYPOINT ["/app/server"]
