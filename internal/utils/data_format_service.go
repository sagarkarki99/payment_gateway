package utils

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"payment-gateway/internal/models"
)

// decodes the incoming request based on content type
func DecodeRequest(r *http.Request, request *models.TransactionRequest) error {
	contentType := r.Header.Get("Content-Type")

	switch contentType {
	case "application/json":
		return json.NewDecoder(r.Body).Decode(request)
	case "text/xml":
		return xml.NewDecoder(r.Body).Decode(request)
	case "application/xml":
		return xml.NewDecoder(r.Body).Decode(request)
	default:
		return fmt.Errorf("unsupported content type")
	}
}

func DecodeCallbackRequest(r *http.Request, request *models.PaymentCallback) error {
	contentType := r.Header.Get("Content-Type")

	switch contentType {
	case "application/json":
		return json.NewDecoder(r.Body).Decode(request)
	case "text/xml":
		return xml.NewDecoder(r.Body).Decode(request)
	case "application/xml":
		return xml.NewDecoder(r.Body).Decode(request)
	default:
		return fmt.Errorf("unsupported content type")
	}
}

func EncodeResponse(contentType string, data interface{}) ([]byte, error) {
	switch contentType {
	case "application/json":
		return json.Marshal(data)
	case "text/xml", "application/xml":
		return xml.Marshal(data)
	default:
		return nil, fmt.Errorf("unsupported content type")
	}
}

// WriteResponse writes a response to the http.ResponseWriter
func WriteResponse(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	contentType := r.Header.Get("Content-Type")

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)

	responseData, err := EncodeResponse(contentType, data)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		return
	}

	if _, err := w.Write(responseData); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}
