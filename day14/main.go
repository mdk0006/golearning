package main

import (
	"fmt"
	"log"
	"net/http"
)

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	fmt.Fprintf(w, "status: healthy")
}
func handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	fmt.Fprintf(w, "cpu_usage 42.5 \nmemory_usage 61.2")
}
func handle404(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	fmt.Fprintf(w, "not found")
}
func main() {
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/metrics", handleMetrics)
	http.HandleFunc("/", handle404)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
