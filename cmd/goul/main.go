package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"goul/internal/config"
	"goul/internal/db"
	"goul/internal/ollama"
	"goul/internal/web"
)

func main() {
	cfg := config.Load()

	fmt.Println("==================================================")
	fmt.Println(" [o.o] GOUL - The Homelab Task Dissector")
	fmt.Printf(" Port:     %s\n", cfg.Port)
	fmt.Printf(" Database: %s\n", cfg.DBPath)
	fmt.Printf(" Ollama:   %s\n", cfg.OllamaURL)
	fmt.Printf(" Model:    %s\n", cfg.OllamaModel)
	fmt.Println("==================================================")

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("Fatal: could not open database: %v", err)
	}
	defer database.Close()

	ollamaClient := ollama.NewClient(cfg.OllamaURL, cfg.OllamaModel, 90*time.Second)

	webServer, err := web.NewServer(database, ollamaClient, cfg.OllamaModel)
	if err != nil {
		log.Fatalf("Fatal: could not initialize web server: %v", err)
	}

	httpServer := &http.Server{
		Addr:         cfg.Port,
		Handler:      webServer.Routes(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown channel
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("The ghoul awakens at http://localhost%s\n", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down the crypt...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}
	log.Println("The ghoul rests peacefully.")
}
