package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/marcosbitetti/experiments_with_http_protocols_with_go/internal"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"golang.org/x/net/http2"
)

var transitionData = `{"key":"value"}`
var iterations = 100

func BenchmarkHTTPServer(b *testing.B) {
	defer internal.StopHTTP1Server()
	counter := internal.NewCounter()
	go internal.Http1Server(counter)
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,              // Maximum idle (keep-alive) connections.
			MaxIdleConnsPerHost: 100,              // Maximum idle connections per host.
			IdleConnTimeout:     internal.Timeout, // Idle connection timeout.
			DisableCompression:  false,            //optional
		},
	}
	b.ResetTimer()
	b.StartTimer()
	doDataLoop(b, client, "http://localhost:8080/json", counter)
	b.StopTimer()
}

func BenchmarkHTTP2Server(b *testing.B) {
	defer internal.StopHTTP2Server()
	counter := internal.NewCounter()
	go internal.Http2Server(counter)
	client := &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: false,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			IdleConnTimeout:    internal.Timeout, // Idle connection timeout.
			DisableCompression: false,            //optional
		},
	}
	b.ResetTimer()
	url := "https://localhost:8080/json"
	b.StartTimer()
	doDataLoop(b, client, url, counter)
	b.StopTimer()
}

func BenchmarkHTTP3Server(b *testing.B) {
	defer internal.StopHTTP3Server()
	counter := internal.NewCounter()
	go internal.Http3Server(counter)

	// Create an HTTP/3 client using the quic-go/http3 package
	transport := &http3.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		QUICConfig: &quic.Config{
			MaxIdleTimeout:  internal.Timeout,
			KeepAlivePeriod: internal.Timeout,
			//DisablePathMTUDiscovery: true,
			//Allow0RTT:               true,
		},
	}
	defer transport.Close()

	client := &http.Client{
		Transport: transport,
	}
	b.ResetTimer()
	url := "https://localhost:8080/json"
	b.StartTimer()
	doDataLoop(b, client, url, counter)
	b.StopTimer()
}

func BenchmarkHTTP3ServerWithDefaults(b *testing.B) {
	defer internal.StopHTTP3Server()
	counter := internal.NewCounter()
	go internal.Http3Server(counter)

	// Create an HTTP/3 client using the quic-go/http3 package
	transport := &http3.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	defer transport.Close()

	client := &http.Client{
		Transport: transport,
	}
	b.ResetTimer()
	url := "https://localhost:8080/json"
	b.StartTimer()
	doDataLoop(b, client, url, counter)
	b.StopTimer()
}

func doDataLoop(b *testing.B, client *http.Client, url string, counter *internal.Counter) {
	raw, err := os.ReadFile("./..//data/pokedex.json")
	if err != nil {
		b.Fatal(err)
	}
	var data []map[string]interface{}
	if err := json.Unmarshal(raw, &data); err != nil {
		b.Fatal(err)
	}
	for i := 0; i < iterations; i++ {
		var wg sync.WaitGroup
		wg.Add(len(data))
		counter.ResetCount()
		for _, v := range data {
			delete(v, "id")
			go func(payload map[string]interface{}) {
				defer wg.Done()
				payloadRaw, err := json.Marshal(payload)
				if err != nil {
					b.Fatal(err)
				}
				req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payloadRaw))
				if err != nil {
					b.Fatal(err)
				}
				req.Header.Set("Content-Type", "application/json")
				resp, err := client.Do(req)
				if err != nil {
					b.Fatal(err)
				}
				buf, err := io.ReadAll(resp.Body)
				if err != nil {
					b.Fatal(err)
				}
				if !strings.Contains(string(buf), `"id":`) {
					b.Fatal("failed to fetch data: " + string(buf))
				}
				resp.Body.Close()
			}(v)
		}
		wg.Wait()
	}
}
