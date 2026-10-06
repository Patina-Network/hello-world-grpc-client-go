package greeter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "patinanetwork.org/grpc/hello-world-grpc-service"
)

var (
	ErrInvalidArgument   = errors.New("invalid argument")
	ErrNotFound          = errors.New("not found")
	ErrAlreadyExists     = errors.New("already exists")
	ErrUnauthenticated   = errors.New("unauthenticated")
	ErrPermissionDenied  = errors.New("permission denied")
	ErrResourceExhausted = errors.New("resource exhausted")
	ErrUnavailable       = errors.New("unavailable")
	ErrTimeout           = errors.New("deadline exceeded or canceled")
)

type NewGreeting struct {
	SenderName    string
	RecipientName string
	Message       string
}

type Greeting struct {
	ID            uint32
	Message       string
	SenderName    string
	RecipientName string
	ReceivedAt    *time.Time
}

type Client struct {
	rpc     pb.GreeterServiceClient
	timeout time.Duration
}

func New(rpc pb.GreeterServiceClient, timeout time.Duration) *Client {
	return &Client{rpc: rpc, timeout: timeout}
}

func (c *Client) Echo(ctx context.Context, name string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	reply, err := c.rpc.EchoHello(ctx, &pb.EchoHelloRequest{Name: name})
	if err != nil {
		return "", classify(err)
	}
	return reply.GetResponse(), nil
}

func (c *Client) SendGreeting(ctx context.Context, g NewGreeting) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	_, err := c.rpc.SayGreeting(ctx, &pb.SayGreetingRequest{
		SenderName:    g.SenderName,
		RecipientName: g.RecipientName,
		Greeting:      g.Message,
	})
	return classify(err)
}

func (c *Client) ListGreetings(ctx context.Context, recipientName *string) ([]Greeting, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	reply, err := c.rpc.GetGreetingsByName(ctx, &pb.GetGreetingsByNameRequest{RecipientName: recipientName})
	if err != nil {
		return nil, classify(err)
	}
	greetings := make([]Greeting, 0, len(reply.GetReplies()))
	for _, r := range reply.GetReplies() {
		g := Greeting{
			ID:            r.GetId(),
			Message:       r.GetMessage(),
			SenderName:    r.GetSenderName(),
			RecipientName: r.GetRecipientName(),
		}
		if ts := r.GetReceivedAt(); ts != nil {
			at := ts.AsTime()
			g.ReceivedAt = &at
		}
		greetings = append(greetings, g)
	}
	return greetings, nil
}

func classify(err error) error {
	if err == nil {
		return nil
	}

	var kind error
	switch status.Code(err) {
	case codes.InvalidArgument:
		kind = ErrInvalidArgument
	case codes.NotFound:
		kind = ErrNotFound
	case codes.AlreadyExists:
		kind = ErrAlreadyExists
	case codes.Unauthenticated:
		kind = ErrUnauthenticated
	case codes.PermissionDenied:
		kind = ErrPermissionDenied
	case codes.ResourceExhausted:
		kind = ErrResourceExhausted
	case codes.Unavailable:
		kind = ErrUnavailable
	case codes.DeadlineExceeded, codes.Canceled:
		kind = ErrTimeout
	default:
		return err
	}

	return fmt.Errorf("%w: %w", kind, err)
}
