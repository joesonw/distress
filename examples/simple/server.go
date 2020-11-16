package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	log.Fatal(http.ListenAndServe(":10080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		delay, _ := time.ParseDuration(r.URL.Query().Get("delay"))
		if delay == 0 {
			delay = time.Duration(rand.Float64() * 3 * float64(time.Millisecond))
		}
		log.Println("path", r.URL.Path, "delay", delay.String(), "status", r.URL.Query().Get("status"))
		time.Sleep(delay)

		if r.URL.Query().Get("status") == "internal_error" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Error"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})))
}
