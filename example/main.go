package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/adzeitor/background"
	"github.com/adzeitor/background/services/inmemory"
)

func slowHandler(w http.ResponseWriter, _ *http.Request) {
	time.Sleep(time.Second * 10)
	log.Println("Completed")
	_, _ = w.Write([]byte("Completed"))
}

func allJobsHandler(service *inmemory.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		job, err := service.Jobs()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(job)
	}
}

func main() {
	service := inmemory.New()
	logger := log.New(os.Stdout, "jobs:", log.LstdFlags)
	bg := background.NewMiddleware(service, logger)

	asyncSlowHandler := bg.InBackground(http.HandlerFunc(slowHandler), "SLOW_HANDLER")
	http.Handle("/slow", asyncSlowHandler)

	http.Handle("/jobs", allJobsHandler(service))

	fmt.Println("to create background job use http://localhost:3000/slow")
	fmt.Println("to track jobs use http://localhost:3000/jobs")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
