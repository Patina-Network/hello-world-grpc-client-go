package httpserve

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/Patina-Network/hello-world-grpc-client-go/internal/config"
	"github.com/Patina-Network/hello-world-grpc-client-go/internal/grpc/greeter"
)

type fakeGreeter struct {
	err error

	echoName  string
	sent      *greeter.NewGreeting
	listName  *string
	greetings []greeter.Greeting
}

func (f *fakeGreeter) Echo(_ context.Context, name string) (string, error) {
	f.echoName = name
	if f.err != nil {
		return "", f.err
	}
	return "hello " + name, nil
}

func (f *fakeGreeter) SendGreeting(_ context.Context, g greeter.NewGreeting) error {
	f.sent = &g
	return f.err
}

func (f *fakeGreeter) ListGreetings(_ context.Context, recipientName *string) ([]greeter.Greeting, error) {
	f.listName = recipientName
	if f.err != nil {
		return nil, f.err
	}
	return f.greetings, nil
}

func serve(svc Greeter, method, path, contentType, body string) *httptest.ResponseRecorder {
	return serveWithURLs(svc, []string{}, method, path, contentType, body)
}

func serveWithURLs(svc Greeter, urls []string, method, path, contentType, body string) *httptest.ResponseRecorder {
	static := fstest.MapFS{"index.html": {Data: []byte("<!doctype html>ui")}}
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if contentType != "" {
		r.Header.Set("Content-Type", contentType)
	}
	w := httptest.NewRecorder()
	metrics := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("metrics")) })
	cfg := &config.Config{Version: "abc1234", URLs: urls}
	routes := slices.Concat(
		SysRoutes(cfg, metrics),
		HelloRoutes(svc),
		GreetingRoutes(svc),
	)
	NewHandler(HandlerOptions{
		Static: static,
		Routes: routes,
	}).ServeHTTP(w, r)
	return w
}

func expect(t *testing.T, w *httptest.ResponseRecorder, code int, body string) {
	t.Helper()
	if w.Code != code || strings.TrimSpace(w.Body.String()) != body {
		t.Fatalf("got %d %s, want %d %s", w.Code, w.Body, code, body)
	}
}

func TestEcho(t *testing.T) {
	svc := &fakeGreeter{}
	w := serve(svc, "GET", "/api/echo?name=Ada%20L", "", "")
	expect(t, w, 200, `{"response":"hello Ada L"}`)
	if svc.echoName != "Ada L" {
		t.Fatalf("passed %q", svc.echoName)
	}
}

func TestSendGreeting(t *testing.T) {
	svc := &fakeGreeter{}
	w := serve(svc, "POST", "/api/greetings", "application/json; charset=utf-8", `{"senderName":"Ada","recipientName":"Lin","greeting":"hi"}`)
	expect(t, w, 200, `{}`)
	if svc.sent == nil || *svc.sent != (greeter.NewGreeting{SenderName: "Ada", RecipientName: "Lin", Message: "hi"}) {
		t.Fatalf("sent %+v", svc.sent)
	}
}

func TestSendGreetingAcceptsJSONMediaTypes(t *testing.T) {
	for _, contentType := range []string{"application/json", "application/json;charset=utf-8", "application/cloudevents+json"} {
		t.Run(contentType, func(t *testing.T) {
			svc := &fakeGreeter{}
			expect(t, serve(svc, "POST", "/api/greetings", contentType, `{"senderName":"","recipientName":"","greeting":""}`), 200, `{}`)
			if svc.sent == nil || *svc.sent != (greeter.NewGreeting{}) {
				t.Fatalf("sent %+v", svc.sent)
			}
		})
	}
}

func TestSendGreetingRejectsMalformedRequests(t *testing.T) {
	const valid = `"senderName":"A","recipientName":"B","greeting":"hi"`
	for _, tc := range []struct {
		name, contentType, body string
		code                    int
	}{
		{"missing content type", "", `{` + valid + `}`, 415},
		{"wrong content type", "text/json", `{` + valid + `}`, 415},
		{"truncated JSON", "application/json", `{`, 400},
		{"missing fields", "application/json", `{}`, 400},
		{"missing one field", "application/json", `{"senderName":"A","recipientName":"B"}`, 400},
		{"null field", "application/json", `{"senderName":"A","recipientName":"B","greeting":null}`, 400},
		{"wrong field type", "application/json", `{"senderName":"A","recipientName":"B","greeting":1}`, 400},
		{"duplicate field", "application/json", `{` + valid + `,"greeting":"again"}`, 400},
		{"unknown field", "application/json", `{` + valid + `,"unknown":1}`, 400},
		{"trailing value", "application/json", `{` + valid + `} {}`, 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := &fakeGreeter{}
			expect(t, serve(svc, "POST", "/api/greetings", tc.contentType, tc.body), tc.code, `{"error":"invalid JSON request"}`)
			if svc.sent != nil {
				t.Fatal("malformed request reached the service")
			}
		})
	}
}

func TestListGreetingsRecipientPresence(t *testing.T) {
	for _, tc := range []struct {
		path string
		want *string
	}{
		{"/api/greetings", nil},
		{"/api/greetings?recipientName=", new(string)},
		{"/api/greetings?recipientName=Lin", func() *string { s := "Lin"; return &s }()},
	} {
		t.Run(tc.path, func(t *testing.T) {
			svc := &fakeGreeter{}
			serve(svc, "GET", tc.path, "", "")
			if (svc.listName == nil) != (tc.want == nil) || (tc.want != nil && *svc.listName != *tc.want) {
				t.Fatalf("passed %v, want %v", svc.listName, tc.want)
			}
		})
	}
}

func TestListGreetingsJSON(t *testing.T) {
	at := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	svc := &fakeGreeter{greetings: []greeter.Greeting{
		{ID: 4294967295, Message: "Ada says hi", SenderName: "Ada", RecipientName: "Lin", ReceivedAt: &at},
		{ID: 2, Message: "m", SenderName: "S", RecipientName: "R"},
	}}
	expect(t, serve(svc, "GET", "/api/greetings", "", ""), 200,
		`{"replies":[{"id":4294967295,"message":"Ada says hi","senderName":"Ada","recipientName":"Lin","receivedAt":"2026-10-05T12:00:00Z"},`+
			`{"id":2,"message":"m","senderName":"S","recipientName":"R","receivedAt":null}]}`)
}

func TestListGreetingsEmptyIsArray(t *testing.T) {
	expect(t, serve(&fakeGreeter{}, "GET", "/api/greetings", "", ""), 200, `{"replies":[]}`)
}

func TestServiceErrorsMapToHTTPWithoutLeakingDetail(t *testing.T) {
	for _, tc := range []struct {
		err  error
		code int
	}{
		{greeter.ErrInvalidArgument, 400},
		{greeter.ErrNotFound, 404},
		{greeter.ErrAlreadyExists, 409},
		{greeter.ErrUnauthenticated, 401},
		{greeter.ErrPermissionDenied, 403},
		{greeter.ErrResourceExhausted, 429},
		{greeter.ErrUnavailable, 503},
		{greeter.ErrTimeout, 504},
		{errors.New("something else"), 502},
	} {
		t.Run(tc.err.Error(), func(t *testing.T) {
			svc := &fakeGreeter{err: fmt.Errorf("%w: private infrastructure detail", tc.err)}
			for _, w := range []*httptest.ResponseRecorder{
				serve(svc, "GET", "/api/echo?name=x", "", ""),
				serve(svc, "POST", "/api/greetings", "application/json", `{"senderName":"A","recipientName":"B","greeting":"hi"}`),
				serve(svc, "GET", "/api/greetings", "", ""),
			} {
				if w.Code != tc.code {
					t.Fatalf("got %d, want %d: %s", w.Code, tc.code, w.Body)
				}
				if strings.Contains(w.Body.String(), "private infrastructure") {
					t.Fatalf("leaked upstream detail: %s", w.Body)
				}
			}
		})
	}
}

func TestRouting(t *testing.T) {
	svc := &fakeGreeter{}
	expect(t, serve(svc, "GET", "/metrics", "", ""), 200, `metrics`)
	expect(t, serve(svc, "GET", "/version", "", ""), 200, `abc1234`)
	expect(t, serve(svc, "GET", "/urls", "", ""), 200, `[]`)
	expect(t, serveWithURLs(svc, []string{"https://a.example", "https://b.example"}, "GET", "/urls", "", ""), 200, `["https://a.example","https://b.example"]`)
	expect(t, serve(svc, "GET", "/api/nope", "", ""), 404, `{"error":"not found"}`)
	expect(t, serve(svc, "GET", "/", "", ""), 200, `<!doctype html>ui`)
}
