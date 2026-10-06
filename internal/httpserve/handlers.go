package httpserve

import (
	"encoding/json/v2"
	"net/http"

	"github.com/Patina-Network/hello-world-grpc-client-go/internal/grpc/greeter"
)

const invalidJSONRequest = "invalid JSON request"

type sendGreetingRequest struct {
	SenderName    *string `json:"senderName"`
	RecipientName *string `json:"recipientName"`
	Greeting      *string `json:"greeting"`
}

type sendGreetingResponse struct{}

type echoResponse struct {
	Response string `json:"response"`
}

type greetingResponse struct {
	ID            uint32  `json:"id"`
	Message       string  `json:"message"`
	SenderName    string  `json:"senderName"`
	RecipientName string  `json:"recipientName"`
	ReceivedAt    *string `json:"receivedAt"`
}

type listGreetingsResponse struct {
	Replies []greetingResponse `json:"replies"`
}

type handlers struct {
	svc Greeter
}

func (h *handlers) echo(w http.ResponseWriter, r *http.Request) {
	reply, err := h.svc.Echo(r.Context(), r.URL.Query().Get("name"))
	if err != nil {
		writeServiceError(w, "greeter", err)
		return
	}
	writeJSON(w, http.StatusOK, echoResponse{Response: reply})
}

func (h *handlers) sendGreeting(w http.ResponseWriter, r *http.Request) {
	if !isJSONContentType(r.Header.Get("Content-Type")) {
		writeError(w, http.StatusUnsupportedMediaType, invalidJSONRequest)
		return
	}

	var in sendGreetingRequest
	err := json.UnmarshalRead(
		r.Body,
		&in,
		json.RejectUnknownMembers(true),
	)

	if err != nil || in.SenderName == nil || in.RecipientName == nil || in.Greeting == nil {
		writeError(w, http.StatusBadRequest, invalidJSONRequest)
		return
	}

	err = h.svc.SendGreeting(r.Context(), greeter.NewGreeting{
		SenderName:    *in.SenderName,
		RecipientName: *in.RecipientName,
		Message:       *in.Greeting,
	})
	if err != nil {
		writeServiceError(w, "greeter", err)
		return
	}

	writeJSON(w, http.StatusOK, sendGreetingResponse{})
}

func (h *handlers) listGreetings(w http.ResponseWriter, r *http.Request) {
	var name *string
	if q := r.URL.Query(); q.Has("recipientName") {
		n := q.Get("recipientName")
		name = &n
	}

	greetings, err := h.svc.ListGreetings(r.Context(), name)
	if err != nil {
		writeServiceError(w, "greeter", err)
		return
	}

	replies := make([]greetingResponse, 0, len(greetings))
	for _, g := range greetings {
		reply := greetingResponse{
			ID:            g.ID,
			Message:       g.Message,
			SenderName:    g.SenderName,
			RecipientName: g.RecipientName,
		}
		if g.ReceivedAt != nil {
			ts := formatTimestamp(*g.ReceivedAt)
			reply.ReceivedAt = &ts
		}
		replies = append(replies, reply)
	}

	writeJSON(w, http.StatusOK, listGreetingsResponse{Replies: replies})
}
