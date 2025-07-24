package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/phgermanov/tasks/internal/db"
	"github.com/phgermanov/tasks/internal/handlers"
)

func main() {
	repo, err := db.NewTaskRepository("tasks.db")
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer repo.Close()

	taskHandler := handlers.NewTask(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /tasks", taskHandler.Get)
	mux.HandleFunc("POST /tasks", taskHandler.Create)
	mux.HandleFunc("PUT /tasks/{id}", taskHandler.Update)
	mux.HandleFunc("DELETE /tasks/{id}", taskHandler.Delete)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		log.Println("Server starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
