package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	announcement "example.com/media-announcement-pipeline"
)

func main() {
	apiKey := os.Getenv("INFRAI_API_KEY")
	if apiKey == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	service := announcement.NewService(announcement.NewInfraiClient(apiKey, nil))

	mux := http.NewServeMux()
	mux.HandleFunc("POST /announcements", func(w http.ResponseWriter, r *http.Request) {
		var input announcement.DeliveryRequest
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}
		if input.DeliveryID == "" || input.Channel == "" || input.AssetID == "" || input.JobID == "" || input.CreatorID == "" || input.Title == "" || input.StreamURL == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "delivery_id, channel, asset_id, job_id, creator_id, title, and stream_url are required"})
			return
		}

		result, err := service.Deliver(r.Context(), input)
		if errors.Is(err, announcement.ErrAssetNotReady) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		var apiErr *announcement.InfraiError
		if errors.As(err, &apiErr) {
			status := apiErr.HTTPStatus
			if status < 400 || status >= 500 {
				status = http.StatusBadGateway
			}
			writeJSON(w, status, map[string]string{"error": apiErr.Message})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "announcement delivery failed"})
			return
		}
		writeJSON(w, http.StatusAccepted, result)
	})

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Printf("announcement service listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("encode response: %v", err)
	}
}
