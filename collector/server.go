package collector

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/CodecWhiz/streamlens/cmcd"
)

// Server is the HTTP collector that receives CMCD beacons.
type Server struct {
	buf    *Buffer
	mux    *http.ServeMux
	server *http.Server
}

// NewServer creates a new collector HTTP server.
func NewServer(buf *Buffer, port int) *Server {
	s := &Server{buf: buf, mux: http.NewServeMux()}
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("POST /v1/cmcd", s.handlePostCMCD)
	s.mux.HandleFunc("GET /v1/cmcd", s.handleGetCMCD)

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      s.mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	return s
}

// ListenAndServe starts the server.
func (s *Server) ListenAndServe() error {
	log.Printf("Collector listening on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

// Close gracefully shuts down the server.
func (s *Server) Close() error {
	return s.server.Close()
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

// postBody is the JSON request body for POST /v1/cmcd.
type postBody struct {
	CMCD       string `json:"cmcd"`
	ContentID  string `json:"content_id,omitempty"`
	CDN        string `json:"cdn,omitempty"`
	CountryCode string `json:"country_code,omitempty"`
}

func (s *Server) handlePostCMCD(w http.ResponseWriter, r *http.Request) {
	var body postBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	data, err := cmcd.Parse(body.CMCD)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	event := cmcd.Event{
		Data:        data,
		Timestamp:   time.Now().UnixMilli(),
		ClientIP:    clientIP(r),
		CDN:         body.CDN,
		CountryCode: body.CountryCode,
	}
	if body.ContentID != "" && event.ContentID == "" {
		event.ContentID = body.ContentID
	}

	s.buf.Add(event)
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"accepted"}`))
}

func (s *Server) handleGetCMCD(w http.ResponseWriter, r *http.Request) {
	raw := r.URL.Query().Get("CMCD")
	if raw == "" {
		http.Error(w, `{"error":"missing CMCD parameter"}`, http.StatusBadRequest)
		return
	}

	data, err := cmcd.ParseEncoded(raw)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	event := cmcd.Event{
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
		ClientIP:  clientIP(r),
		CDN:       r.URL.Query().Get("cdn"),
	}

	s.buf.Add(event)
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"accepted"}`))
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
