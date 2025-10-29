package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/cliffdoyle/task-api/internal/handlers"
	"github.com/cliffdoyle/task-api/internal/repository"
	"github.com/cliffdoyle/task-api/internal/service"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables from .env file
	// This is useful for local development; in production, variables are typically set directly.
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables.")
	}

	// --- Database Connection ---
	// The DATABASE_URL environment variable will be used to connect to PostgreSQL.
	// For local development, this will point to our Dockerized PostgreSQL.
	// For Azure, it will point to Azure SQL Database.
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer func() {
		if cerr := db.Close(); cerr != nil {
			log.Printf("Error closing database connection: %v", cerr)
		}
	}()

	// Ping the database to verify the connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}
	log.Println("Successfully connected to the database!")

	// --- Initialize Application Layers ---
	taskRepo := repository.NewTaskRepository(db)
	taskService := service.NewTaskService(taskRepo)
	taskHandler := handlers.NewTaskHandler(taskService)

	// --- Setup Routes ---
	r := mux.NewRouter()

	// Task API routes
	r.HandleFunc("/api/tasks", taskHandler.CreateTask).Methods("POST")
	r.HandleFunc("/api/tasks", taskHandler.GetAllTasks).Methods("GET")
	r.HandleFunc("/api/tasks/{id}", taskHandler.GetTask).Methods("GET")
	r.HandleFunc("/api/tasks/{id}", taskHandler.UpdateTask).Methods("PUT")
	r.HandleFunc("/api/tasks/{id}", taskHandler.DeleteTask).Methods("DELETE")

	// Health check endpoint
	r.HandleFunc("/health", healthCheck).Methods("GET")

	// --- Start HTTP Server ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not specified in environment
	}

	log.Printf("Server starting on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, r)) // Use log.Fatal to gracefully exit on server error
}

// healthCheck handler for basic service availability
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
