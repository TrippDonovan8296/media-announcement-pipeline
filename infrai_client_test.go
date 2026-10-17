package announcement

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return fn(req) }

func TestPublishRetriesRateLimitWithSameDeliveryKey(t *testing.T) {
	calls := 0
	client := NewInfraiClient("test-key", &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		if req.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", req.Method)
		}
		if got := req.Header.Get("Idempotency-Key"); got != "delivery-42" {
			t.Fatalf("idempotency key = %q", got)
		}
		if calls == 1 {
			return response(http.StatusTooManyRequests, `{"ok":false,"data":null,"error":{"code":"RATE_LIMITED","message":"retry later"},"metadata":{}}`, "0"), nil
		}
		return response(http.StatusOK, `{"ok":true,"data":{"published":true},"error":null,"metadata":{}}`, ""), nil
	})})
	client.wait = func(context.Context, time.Duration) error { return nil }

	err := client.Publish(context.Background(), "media-members", Announcement{DeliveryID: "delivery-42"}, "delivery-42")
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

func response(status int, body, retryAfter string) *http.Response {
	header := make(http.Header)
	header.Set("Retry-After", retryAfter)
	return &http.Response{StatusCode: status, Header: header, Body: io.NopCloser(strings.NewReader(body))}
}
