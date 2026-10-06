package greeter

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	pb "patinanetwork.org/grpc/hello-world-grpc-service"
)

type fakeRPC struct {
	err error

	echoReq *pb.EchoHelloRequest
	sayReq  *pb.SayGreetingRequest
	listReq *pb.GetGreetingsByNameRequest
	list    *pb.GreetingsResponse

	deadline    time.Time
	hasDeadline bool
}

func (f *fakeRPC) record(ctx context.Context) { f.deadline, f.hasDeadline = ctx.Deadline() }

func (f *fakeRPC) EchoHello(ctx context.Context, in *pb.EchoHelloRequest, _ ...grpc.CallOption) (*pb.EchoHelloResponse, error) {
	f.record(ctx)
	f.echoReq = in
	if f.err != nil {
		return nil, f.err
	}
	return &pb.EchoHelloResponse{Response: "hello " + in.GetName()}, nil
}

func (f *fakeRPC) SayGreeting(ctx context.Context, in *pb.SayGreetingRequest, _ ...grpc.CallOption) (*pb.SayGreetingResponse, error) {
	f.record(ctx)
	f.sayReq = in
	if f.err != nil {
		return nil, f.err
	}
	return &pb.SayGreetingResponse{}, nil
}

func (f *fakeRPC) GetGreetingsByName(ctx context.Context, in *pb.GetGreetingsByNameRequest, _ ...grpc.CallOption) (*pb.GreetingsResponse, error) {
	f.record(ctx)
	f.listReq = in
	if f.err != nil {
		return nil, f.err
	}
	return f.list, nil
}

func TestEcho(t *testing.T) {
	rpc := &fakeRPC{}
	got, err := New(rpc, time.Second).Echo(context.Background(), "Ada")
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello Ada" || rpc.echoReq.GetName() != "Ada" {
		t.Fatalf("got %q, sent %v", got, rpc.echoReq)
	}
}

func TestSendGreetingMapsFields(t *testing.T) {
	rpc := &fakeRPC{}
	err := New(rpc, time.Second).SendGreeting(context.Background(), NewGreeting{SenderName: "Ada", RecipientName: "Lin", Message: "hi"})
	if err != nil {
		t.Fatal(err)
	}
	if rpc.sayReq.GetSenderName() != "Ada" || rpc.sayReq.GetRecipientName() != "Lin" || rpc.sayReq.GetGreeting() != "hi" {
		t.Fatalf("sent %v", rpc.sayReq)
	}
}

func TestListGreetingsForwardsRecipientPresence(t *testing.T) {
	empty := ""
	for _, tc := range []struct {
		name string
		in   *string
	}{{"unset", nil}, {"empty", &empty}} {
		t.Run(tc.name, func(t *testing.T) {
			rpc := &fakeRPC{list: &pb.GreetingsResponse{}}
			if _, err := New(rpc, time.Second).ListGreetings(context.Background(), tc.in); err != nil {
				t.Fatal(err)
			}
			if (rpc.listReq.RecipientName != nil) != (tc.in != nil) {
				t.Fatalf("recipientName presence: sent %v, want set=%v", rpc.listReq.RecipientName, tc.in != nil)
			}
		})
	}
}

func TestListGreetingsConvertsReplies(t *testing.T) {
	at := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	epoch := time.Unix(0, 0).UTC()
	rpc := &fakeRPC{list: &pb.GreetingsResponse{Replies: []*pb.GreetingResponse{
		{Id: 4294967295, Message: "Ada says hi", SenderName: "Ada", RecipientName: "Lin", ReceivedAt: timestamppb.New(at)},
		{Id: 2, Message: "no timestamp"},
		{Id: 3, Message: "epoch", ReceivedAt: &timestamppb.Timestamp{}},
	}}}
	got, err := New(rpc, time.Second).ListGreetings(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []Greeting{
		{ID: 4294967295, Message: "Ada says hi", SenderName: "Ada", RecipientName: "Lin", ReceivedAt: &at},
		{ID: 2, Message: "no timestamp"},
		{ID: 3, Message: "epoch", ReceivedAt: &epoch},
	}
	if len(got) != len(want) {
		t.Fatalf("got %+v", got)
	}
	for i := range want {
		sameTime := (got[i].ReceivedAt == nil) == (want[i].ReceivedAt == nil) &&
			(want[i].ReceivedAt == nil || got[i].ReceivedAt.Equal(*want[i].ReceivedAt))
		if !sameTime || got[i].ID != want[i].ID || got[i].Message != want[i].Message ||
			got[i].SenderName != want[i].SenderName || got[i].RecipientName != want[i].RecipientName {
			t.Fatalf("greeting %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestCallsAreBoundedByTimeout(t *testing.T) {
	const timeout = 250 * time.Millisecond
	rpc := &fakeRPC{list: &pb.GreetingsResponse{}}
	c := New(rpc, timeout)
	calls := map[string]func() error{
		"echo": func() error { _, err := c.Echo(context.Background(), "x"); return err },
		"send": func() error { return c.SendGreeting(context.Background(), NewGreeting{}) },
		"list": func() error { _, err := c.ListGreetings(context.Background(), nil); return err },
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			rpc.hasDeadline = false
			start := time.Now()
			if err := call(); err != nil {
				t.Fatal(err)
			}
			end := time.Now()
			if !rpc.hasDeadline || rpc.deadline.Before(start.Add(timeout)) || rpc.deadline.After(end.Add(timeout)) {
				t.Fatalf("deadline set=%v at %v after start, want %v", rpc.hasDeadline, rpc.deadline.Sub(start), timeout)
			}
		})
	}
}

func TestErrorsAreClassifiedByCode(t *testing.T) {
	calls := map[string]func(*testing.T, *Client) error{
		"Echo": func(_ *testing.T, c *Client) error {
			_, err := c.Echo(context.Background(), "x")
			return err
		},
		"SendGreeting": func(_ *testing.T, c *Client) error {
			return c.SendGreeting(context.Background(), NewGreeting{})
		},
		"ListGreetings": func(t *testing.T, c *Client) error {
			got, err := c.ListGreetings(context.Background(), nil)
			if got != nil {
				t.Fatalf("returned %v alongside error", got)
			}
			return err
		},
	}
	for _, tc := range []struct {
		code codes.Code
		want error
	}{
		{codes.InvalidArgument, ErrInvalidArgument},
		{codes.NotFound, ErrNotFound},
		{codes.AlreadyExists, ErrAlreadyExists},
		{codes.Unauthenticated, ErrUnauthenticated},
		{codes.PermissionDenied, ErrPermissionDenied},
		{codes.ResourceExhausted, ErrResourceExhausted},
		{codes.Unavailable, ErrUnavailable},
		{codes.DeadlineExceeded, ErrTimeout},
		{codes.Canceled, ErrTimeout},
		{codes.Internal, nil},
	} {
		for name, call := range calls {
			t.Run(name+"/"+tc.code.String(), func(t *testing.T) {
				upstream := status.Error(tc.code, "upstream detail")
				err := call(t, New(&fakeRPC{err: upstream}, time.Second))
				if tc.want != nil && !errors.Is(err, tc.want) {
					t.Fatalf("got %v, want errors.Is %v", err, tc.want)
				}
				if status.Code(err) != tc.code {
					t.Fatalf("original gRPC status lost: %v", err)
				}
			})
		}
	}
}
