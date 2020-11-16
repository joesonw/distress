package app

import (
	"encoding/json"
	"net"
	"net/http"
)

type Stats struct {
	StartedAt          int64   `json:"started_at,omitempty"`
	Concurrency        int64   `json:"concurrency,omitempty"`
	TotalAmount        int64   `json:"total_amount,omitempty"`
	FinishedAmount     int64   `json:"finished_amount,omitempty"`
	DurationMicro      int64   `json:"duration_micro,omitempty"`
	TotalDurationMicro int64   `json:"total_duration_micro,omitempty"`
	AverageCostMicro   float64 `json:"average_cost_micro,omitempty"`
}

func startStatsServer(addr string, job *Job) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	go func() {
		_ = http.Serve(lis, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && r.URL.Path == "/stats" {
				stats, err := job.Stats()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				b, err := json.Marshal(stats)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(b)
			}
		}))
	}()
	return nil
}
