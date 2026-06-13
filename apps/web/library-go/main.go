package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"vyzorix-backend/database"
)

func main() {
	// 1. Init Database tables
	database.InitDB()

	// 2. Setup ServeMux & Register routes (defined in router.go)
	mux := http.NewServeMux()
	RegisterRoutes(mux)

	// Bind Go API backend to port 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Vyzorix Go authentications engine active on http://0.0.0.0:%s\n", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, enableCORS(mux)); err != nil {
		log.Fatalf("Go Server crash: %v", err)
	}
}

// enableCORS Middleware is a developer comfort feature to bypass browser security blocks
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
