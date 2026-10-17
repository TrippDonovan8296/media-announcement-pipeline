package announcement

import (
	"context"
	"errors"
)

type DeliveryRequest struct {
	DeliveryID string `json:"delivery_id"`
	Channel    string `json:"channel"`
	AssetID    string `json:"asset_id"`
	JobID      string `json:"job_id"`
	JobState   string `json:"job_state"`
	CreatorID  string `json:"creator_id"`
	Title      string `json:"title"`
	StreamURL  string `json:"stream_url"`
}

type Announcement struct {
	DeliveryID string `json:"delivery_id"`
	AssetID    string `json:"asset_id"`
	JobID      string `json:"job_id"`
	CreatorID  string `json:"creator_id"`
	Title      string `json:"title"`
	StreamURL  string `json:"stream_url"`
}

type DeliveryResult struct {
	DeliveryID string `json:"delivery_id"`
	State      string `json:"state"`
	Channel    string `json:"channel"`
}

var ErrAssetNotReady = errors.New("asset processing job is not ready")

type Publisher interface {
	Publish(context.Context, string, Announcement, string) error
}

type Service struct {
	publisher Publisher
}

func NewService(publisher Publisher) *Service {
	return &Service{publisher: publisher}
}

func PlanDelivery(input DeliveryRequest) (Announcement, error) {
	if input.JobState != "ready" {
		return Announcement{}, ErrAssetNotReady
	}
	return Announcement{
		DeliveryID: input.DeliveryID,
		AssetID:    input.AssetID,
		JobID:      input.JobID,
		CreatorID:  input.CreatorID,
		Title:      input.Title,
		StreamURL:  input.StreamURL,
	}, nil
}

func (s *Service) Deliver(ctx context.Context, input DeliveryRequest) (DeliveryResult, error) {
	announcement, err := PlanDelivery(input)
	if err != nil {
		return DeliveryResult{}, err
	}
	if err := s.publisher.Publish(ctx, input.Channel, announcement, input.DeliveryID); err != nil {
		return DeliveryResult{}, err
	}
	return DeliveryResult{DeliveryID: input.DeliveryID, State: "broadcast", Channel: input.Channel}, nil
}
