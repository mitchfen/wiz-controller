package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/mitchfen/wiz-controller/internal/handlers"
	"github.com/mitchfen/wiz-controller/internal/services"
)

func main() {
	configPath := "config.json"
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		configPath = path
	}

	cfg, err := services.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Loaded %d lights", len(cfg.Lights))

	wizSvc := services.NewWizService()
	h := handlers.NewHandlers(cfg, wizSvc)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.Home)
	mux.HandleFunc("POST /api/lights/{ip}/brightness", h.SetBrightness)
	mux.HandleFunc("POST /api/lights/all/brightness", h.SetAllBrightness)
	mux.HandleFunc("POST /api/groups/{group}/brightness", h.SetGroupBrightness)
	mux.HandleFunc("POST /api/sync-scenes", h.SyncScenes)
	mux.HandleFunc("POST /api/scenes/warm", h.SetWarm)
	mux.HandleFunc("POST /api/scenes/daylight", h.SetDaylight)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	port := ":80"
	if p := os.Getenv("PORT"); p != "" {
		port = ":" + p
	}

	log.Printf("Starting server on %s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
