package internal

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// JSONHandler processes POST requests containing a JSON payload
// and returns a JSEND success response with the parsed payload.
func createHTTP2Handler(counter *Counter) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Write([]byte("done"))
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.ProtoMajor != 2 {
			http.Error(w, "HTTP/2 Required", http.StatusBadRequest)
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
	}
}

type http2handler struct{}

func (h *http2handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

var server2 *http.Server

// Http1Server starts an HTTP server on port 8080.
func Http2Server(counter *Counter) {
	// Grant certificates if they don't exist
	//err := certificates.RequestCertificates("server.crt", "server.key")
	//if err != nil {
	//	log.Fatalf("failed to request certificates: %v", err)
	//}

	handler := http.HandlerFunc(createHTTP2Handler(counter))

	h2s := &http2.Server{
		MaxConcurrentStreams: 250,
		IdleTimeout:          1 * time.Second,
	}

	h2cHandler := h2c.NewHandler(handler, h2s)

	log.Println("Starting HTTP server on :8080")
	// shut it down later.
	server2 = &http.Server{
		Addr:    ":8080",
		Handler: h2cHandler,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: "localhost",
			//Certificates: []tls.Certificate{cert},
			InsecureSkipVerify: true,
			VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
				//log.Printf("Peer certificates: %+v", rawCerts)
				return nil
			},
			VerifyConnection: func(cs tls.ConnectionState) error {
				//log.Printf("Connection state: %+v", cs)
				return nil
			},
		},
		IdleTimeout: Timeout,
		//TLSConfig: tlsCfg,
		//TLSConfig: &tls.Config{
		//	//Certificates: []tls.Certificate{cert},
		//},
		//TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	// ListenAndServe is blocking.
	//if err := server2.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	if err := server2.ListenAndServeTLS("server.crt", "server.key"); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not start server: %v", err)
	}
}

// StopHTTPServer gracefully shuts down the server with a timeout.
func StopHTTP2Server() error {
	if server2 != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server2.Shutdown(ctx)
	}
	return nil
}
