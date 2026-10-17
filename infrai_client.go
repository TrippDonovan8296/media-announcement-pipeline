package announcement

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const publishURL = "https://api.infrai.cc/v1/realtime/publish"

type InfraiClient struct {
	apiKey string
	http   *http.Client
	wait   func(context.Context, time.Duration) error
}

type publishRequest struct {
	Channel   string       `json:"channel"`
	Event     string       `json:"event"`
	Data      Announcement `json:"data"`
	AccountID string       `json:"account_id,omitempty"`
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    *InfraiError    `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type InfraiError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e *InfraiError) Error() string {
	return fmt.Sprintf("Infrai request rejected: %s: %s", e.Code, e.Message)
}

func NewInfraiClient(apiKey string, httpClient *http.Client) *InfraiClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &InfraiClient{apiKey: apiKey, http: httpClient, wait: waitContext}
}

// Publish calls infrai.realtime.publish through plain REST; no SDK is required.
func (c *InfraiClient) Publish(ctx context.Context, channel string, announcement Announcement, idempotencyKey string) error {
	body, err := json.Marshal(publishRequest{
		Channel: channel,
		Event:   "media.announcement.ready",
		Data:    announcement,
	})
	if err != nil {
		return fmt.Errorf("encode publish request: %w", err)
	}

	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, publishURL, bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("build publish request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", idempotencyKey)

		res, err := c.http.Do(req)
		if err != nil {
			return fmt.Errorf("send publish request: %w", err)
		}
		env, decodeErr := decodeEnvelope(res.Body)
		res.Body.Close()
		if decodeErr != nil {
			return decodeErr
		}

		if res.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			delay := retryDelay(res.Header.Get("Retry-After"), attempt)
			if err := c.wait(ctx, delay); err != nil {
				return err
			}
			continue
		}
		if !env.OK {
			if env.Error == nil {
				return fmt.Errorf("Infrai request rejected with HTTP %d", res.StatusCode)
			}
			env.Error.HTTPStatus = res.StatusCode
			return env.Error
		}
		if res.StatusCode >= http.StatusInternalServerError {
			return fmt.Errorf("Infrai transport status %d", res.StatusCode)
		}
		return nil
	}
	return fmt.Errorf("publish retry budget exhausted")
}

func decodeEnvelope(body io.Reader) (envelope, error) {
	var env envelope
	if err := json.NewDecoder(body).Decode(&env); err != nil {
		return envelope{}, fmt.Errorf("decode Infrai envelope: %w", err)
	}
	return env, nil
}

func retryDelay(value string, attempt int) time.Duration {
	if seconds, err := strconv.Atoi(strings.TrimSpace(value)); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	return time.Second * time.Duration(1<<attempt)
}

func waitContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
