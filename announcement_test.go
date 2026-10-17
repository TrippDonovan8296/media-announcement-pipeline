package announcement

import (
	"errors"
	"testing"
)

func TestPlanDelivery(t *testing.T) {
	tests := []struct {
		name      string
		jobState  string
		wantError error
	}{
		{name: "completed processing job", jobState: "ready"},
		{name: "queued processing job", jobState: "queued", wantError: ErrAssetNotReady},
		{name: "running processing job", jobState: "processing", wantError: ErrAssetNotReady},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PlanDelivery(DeliveryRequest{
				DeliveryID: "delivery-42",
				AssetID:    "asset-17",
				JobID:      "job-9",
				JobState:   tt.jobState,
				CreatorID:  "creator-3",
				Title:      "Studio update",
				StreamURL:  "https://media.example/asset-17/master.m3u8",
			})
			if !errors.Is(err, tt.wantError) {
				t.Fatalf("PlanDelivery() error = %v, want %v", err, tt.wantError)
			}
			if tt.wantError == nil && got.AssetID != "asset-17" {
				t.Fatalf("PlanDelivery() asset = %q, want asset-17", got.AssetID)
			}
		})
	}
}
