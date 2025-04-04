package internal

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

// JSONHandler processes POST requests containing a JSON payload
// and returns a JSEND success response with the parsed payload.
func createHTTP3Handler(counter *Counter) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := server3.SetQUICHeaders(w.Header()); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if r.Method == http.MethodGet {
			w.Write([]byte("done"))
			return
		}
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
	}
}

var server3 *http3.Server

// Http3Server starts an HTTP/3 server on port 8443.
func Http3Server(counter *Counter) {
	handler := http.HandlerFunc(createHTTP3Handler(counter))

	// Configure TLS
	tlsConfig := &tls.Config{
		//MinVersion:         tls.VersionTLS13, // HTTP/3 requires TLS 1.3
		ServerName:         "localhost",
		InsecureSkipVerify: true,
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			return nil
		},
		VerifyConnection: func(cs tls.ConnectionState) error {
			return nil
		},
	}

	// Configure QUIC
	quicConfig := &quic.Config{
		MaxIdleTimeout:        Timeout,
		MaxIncomingStreams:    1000,
		MaxIncomingUniStreams: 1000,
		// You can add more QUIC-specific configurations here
	}

	// Create HTTP/3 server
	server3 = &http3.Server{
		Port:       8080,
		Addr:       ":8080",
		TLSConfig:  tlsConfig,
		QUICConfig: quicConfig,
		Handler:    handler,
	}

	log.Println("Starting HTTP/3 server on :8080")

	// ListenAndServe is blocking
	if err := server3.ListenAndServeTLS("server.crt", "server.key"); err != nil {
		log.Fatalf("Could not start HTTP/3 server: %v", err)
	}
}

// StopHTTP3Server gracefully shuts down the server with a timeout.
func StopHTTP3Server() error {
	if server3 != nil {
		_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server3.Close()
	}
	return nil
}
