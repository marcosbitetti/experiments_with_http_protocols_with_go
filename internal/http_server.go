package internal

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// JSONHandler processes POST requests containing a JSON payload
// and returns a JSEND success response with the parsed payload.
func createHTTP1Handler(counter *Counter) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		payload["id"] = counter.GetCountAndIncrement()
		response := JSendResponse{
			Status: "success",
			Data:   payload,
		}

		responseRaw, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(responseRaw)
		if err != nil {
			log.Println(err)
		}
		//log.Println(fmt.Sprintf("%d", t), r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent(), r.Referer(), r.Proto, r.ContentLength, r.Header.Get("Content-Type"), r.Header.Get("User-Agent"))

		//if err := json.NewEncoder(w).Encode(response); err != nil {
		//	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		//}
	}
}

var server1 *http.Server

// Http1Server starts an HTTP server on port 8080.
func Http1Server(counter *Counter) {
	mux := http.NewServeMux()
	// Route to the JSONHandler on the /json endpoint.
	mux.HandleFunc("/json", createHTTP1Handler(counter))
	log.Println("Starting HTTP server on :8080")

	// Create the server so we can gracefully shut it down later.
	server1 = &http.Server{
		Addr:        ":8080",
		Handler:     mux,
		IdleTimeout: Timeout,
	}

	// ListenAndServe is blocking.
	if err := server1.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not start server: %v", err)
	}
}

// StopHTTPServer gracefully shuts down the server with a timeout.
func StopHTTP1Server() error {
	if server1 != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server1.Shutdown(ctx)
	}
	return nil
}
