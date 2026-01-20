package httplogger

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	logger "github.com/ecpartan/soap-server-tr069/log"
)

// RequestLog represents the structure of request logs
type RequestLog struct {
	Time       time.Time   `json:"time"`
	Method     string      `json:"method"`
	URL        string      `json:"url"`
	RemoteAddr string      `json:"remote_addr"`
	UserAgent  string      `json:"user_agent,omitempty"`
	Headers    http.Header `json:"headers,omitempty"`
	Body       interface{} `json:"body,omitempty"`
	BodySize   int         `json:"body_size"`
}

// ResponseLog represents the structure of response logs
type ResponseLog struct {
	Time       time.Time     `json:"time"`
	StatusCode int           `json:"status_code"`
	Status     string        `json:"status"`
	Headers    http.Header   `json:"headers,omitempty"`
	Body       interface{}   `json:"body,omitempty"`
	BodySize   int           `json:"body_size"`
	Duration   time.Duration `json:"duration_ms"`
}

// HTTPLogger is the logging middleware
type HTTPLogger struct {
	logger *log.Logger
}

// NewHTTPLogger creates a new HTTP logger
func NewHTTPLogger() *HTTPLogger {
	err := os.Remove("./output")
	if err != nil {
		logger.LogErr("Error removing file: %v", err)
	}
	file, err := os.OpenFile("./output", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	return &HTTPLogger{
		logger: log.New(file, "", 0),
	}
}

// Middleware returns a middleware function
func (h *HTTPLogger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Read and buffer request body
		var reqBody []byte
		if r.Body != nil {
			reqBody, _ = io.ReadAll(r.Body)
			r.Body.Close()
			logger.LogDebug("Request: %s %s", r.Method, string(reqBody))

			r.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		// Parse request body as JSON if possible
		var reqBodyJSON interface{}
		if len(reqBody) > 0 {
			if json.Valid(reqBody) {
				json.Unmarshal(reqBody, &reqBodyJSON)
			} else {
				reqBodyJSON = string(reqBody)
			}
		}

		// Log request
		requestLog := RequestLog{
			Time:       time.Now(),
			Method:     r.Method,
			URL:        r.URL.String(),
			RemoteAddr: r.RemoteAddr,
			UserAgent:  r.UserAgent(),
			Headers:    r.Header,
			Body:       reqBodyJSON,
			BodySize:   len(reqBody),
		}

		h.logJSON("REQUEST", requestLog)

		// Create custom response writer to capture response
		rw := &responseWriter{
			ResponseWriter: w,
			body:           &bytes.Buffer{},
			statusCode:     http.StatusOK,
		}

		// Call next handler
		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		// Parse response body as JSON if possible
		var respBodyJSON interface{}
		respBody := rw.body.Bytes()
		if len(respBody) > 0 {
			if json.Valid(respBody) {
				json.Unmarshal(respBody, &respBodyJSON)
			} else {
				respBodyJSON = string(respBody)
			}
		}

		// Log response
		responseLog := ResponseLog{
			Time:       time.Now(),
			StatusCode: rw.statusCode,
			Status:     http.StatusText(rw.statusCode),
			Headers:    rw.Header(),
			Body:       respBodyJSON,
			BodySize:   len(respBody),
			Duration:   duration,
		}

		h.logJSON("RESPONSE", responseLog)
	})
}

// Custom response writer to capture response
type responseWriter struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Helper method to log as JSON
func (h *HTTPLogger) logJSON(prefix string, data interface{}) {
	logData := map[string]interface{}{
		"type": prefix,
		"data": data,
	}

	jsonData, err := json.MarshalIndent(logData, "", "  ")
	if err != nil {
		h.logger.Printf("Error marshaling log: %v", err)
		return
	}

	h.logger.Println(string(jsonData))
}

// Simple method for one-line JSON logs
func (h *HTTPLogger) logCompact(prefix string, data interface{}) {
	logData := map[string]interface{}{
		"type": prefix,
		"data": data,
	}

	jsonData, err := json.Marshal(logData)
	if err != nil {
		h.logger.Printf("Error marshaling log: %v", err)
		return
	}

	h.logger.Println(string(jsonData))
}
