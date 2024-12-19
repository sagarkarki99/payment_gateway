package utils

import (
	"net/http"
	"payment-gateway/internal/models"
)

// HandleError converts an error to an APIError response
func HandleError(err error) models.APIError {
	statusCode := models.GetStatusCode(err)

	return models.APIError{
		StatusCode: statusCode,
		Error:      err.Error(),
	}
}

// WriteErrorResponse writes an error response to the http.ResponseWriter
func WriteErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	errResponse := HandleError(err)
	WriteResponse(w, r, errResponse.StatusCode, errResponse)
}
