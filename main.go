package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Shared HTTP client with a hard timeout so a slow upstream can't hang us.
var client = &http.Client{Timeout: 5 * time.Second}

// Only the fields we care about from the GitHub API response.
type githubRepo struct {
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Stars       int    `json:"stargazers_count"`
	Language    string `json:"language"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// GET /api/repo -> calls the GitHub API and returns a trimmed-down result.
func repoHandler(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet,
		"https://api.github.com/repos/golang/go", nil)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "small-go-server")

	resp, err := client.Do(req)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		writeJSON(w, http.StatusBadGateway, map[string]string{
			"error": fmt.Sprintf("upstream returned %d", resp.StatusCode),
		})
		return
	}

	var repo githubRepo
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "bad upstream JSON"})
		return
	}

	writeJSON(w, http.StatusOK, repo)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /api/repo", repoHandler)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Graceful shutdown on Ctrl+C / SIGTERM.
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	log.Printf("listening on http://localhost:%s", port)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}


/*
Created a file, shared files
Here's a single-file Go server using only the standard library (needs Go 1.22+ for the GET /path routing syntax).

Run it:

bash
go mod init smallserver   # first time only
go run main.go



Try it:

bash
curl http://localhost:8080/health
curl http://localhost:8080/api/repo
/api/repo calls the GitHub API from the server and returns a trimmed JSON result (name, description, stars, language). It has a 5s timeout, returns a 502 if the upstream fails, and shuts down gracefully on Ctrl+C. The port defaults to 8080 and can be changed with PORT=9000 go run main.go.

If you meant something different, such as calling a specific API of yours, or accepting a parameter like /api/repo?name=owner/repo, tell me which API and I'll adapt it.

*/