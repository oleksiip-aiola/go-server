package chatservice

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
)

func ChatService() {
	// Load the TLS certificate and key
	certFile := "certificate.pem"
	keyFile := "certificate.key"
	tlsCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("Failed to load TLS certificates: %v", err)
	}

	// Set up the TLS configuration
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{http3.NextProtoH3}, // Enable HTTP/3
	}

	// Create a new WebTransport server
	wtHandler := webtransport.Server{}

	// Define an HTTP handler that upgrades to WebTransport
	http.HandleFunc("/wt", func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor != 3 || r.Header.Get("Sec-Webtransport-Http3-Draft") == "" {
			http.Error(w, "WebTransport not supported", http.StatusBadRequest)
			return
		}

		// Upgrade the connection to WebTransport
		session, err := wtHandler.Upgrade(w, r)
		if err != nil {
			log.Printf("Failed to upgrade to WebTransport: %v", err)
			return
		}

		log.Println("WebTransport session established")

		// Handle WebTransport streams
		go handleSession(session)
	})

	// Create an HTTP/3 server using the quic-go library
	server := http3.Server{
		Addr:      ":4433",              // Serve on port 4433 (HTTP/3)
		TLSConfig: tlsConfig,            // Use the TLS configuration
		Handler:   http.DefaultServeMux, // Use the default HTTP handler
	}

	log.Println("Starting WebTransport server on :4433")
	err = server.ListenAndServeTLS(certFile, keyFile)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// handleSession handles WebTransport sessions and processes streams
func handleSession(session *webtransport.Session) {
	for {
		stream, err := session.AcceptStream(context.Background())
		if err != nil {
			log.Printf("Error accepting stream: %v", err)
			return
		}

		// Process the stream
		go handleStream(stream)
	}
}

// handleStream processes WebTransport streams
func handleStream(stream webtransport.Stream) {
	defer stream.Close()

	// Echo back data received from the client
	buf := make([]byte, 1024)
	for {
		n, err := stream.Read(buf)
		if err != nil {
			log.Printf("Stream read error: %v", err)
			return
		}

		fmt.Printf("Received: %s\n", string(buf[:n]))

		// Echo the data back to the client
		_, err = stream.Write(buf[:n])
		if err != nil {
			log.Printf("Stream write error: %v", err)
			return
		}
	}
}
