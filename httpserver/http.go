package httpserver

import (
	"fmt"
	"net/http"

	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/soap"
)

type responseWriter struct {
	w             http.ResponseWriter
	outputStarted bool
}

func (w *responseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.outputStarted = true
	logger.LogDebug("writing response: ", string(b))

	return w.w.Write(b)
}

func (w *responseWriter) WriteHeader(code int) {
	w.w.WriteHeader(code)
}

// WriteHeader first set the content-type header and then writes the header code.
func WriteHeader(w http.ResponseWriter, code int) {
	setContentType(w, ContentType)
	w.WriteHeader(code)
}

func setContentType(w http.ResponseWriter, contentType string) {
	w.Header().Set("Content-Type", contentType)
}

func addSOAPHeader(w http.ResponseWriter, contentLength int, contentType string) {
	setContentType(w, contentType)
	w.Header().Set("Content-Length", fmt.Sprint(contentLength))
}

func TransmitXMLReq(request any, w http.ResponseWriter) {
	xmlBytes, err := s.Marshaller.Marshal(request)
	// Adjust namespaces for SOAP 1.2

	if err != nil {
		s.handleError(fmt.Errorf("could not marshal response:: %s", err), w)
	}
	addSOAPHeader(w, len(xmlBytes), s.ContentType)
	w.Write(xmlBytes)

}

func HandleError(err error, w http.ResponseWriter) {
	// has to write a soap fault
	logger.LogDebug("handling error:", err)
	responseEnvelope := &soap.Envelope{}
	xmlBytes, xmlErr := s.Marshaller.Marshal(responseEnvelope)
	if xmlErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "could not marshal soap fault for: %s xmlError: %s\n", err, xmlErr)
		return
	}
	addSOAPHeader(w, len(xmlBytes), s.ContentType)
	w.Write(xmlBytes)
}
