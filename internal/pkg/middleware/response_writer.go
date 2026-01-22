package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type errorResponse struct {
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

type successResponse struct {
	Data any `json:"data"`
}

func SuccessResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	payload := successResponse{
		Data: data,
	}

	out, err := json.Marshal(payload)
	if err != nil {
		log.Printf("cannot marshal success response payload")
		writeServerErrorResponse(w)
		return
	}

	_, err = w.Write(out)
	if err != nil {
		log.Printf("cannot write json success response")
		writeServerErrorResponse(w)
		return
	}
}

func ErrorResponse(w http.ResponseWriter, httpStatus int, err error) {
	log.Printf("error: %v\n", err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	payload := errorResponse{
		ErrorCode:    httpStatus,
		ErrorMessage: err.Error(),
	}

	out, err := json.Marshal(payload)
	if err != nil {
		log.Printf("error occurred while marshalling response payload %v", err)
		return
	}

	_, err = w.Write(out)
	if err != nil {
		writeServerErrorResponse(w)
		return
	}
}

func writeServerErrorResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	_, err := fmt.Fprintf(w, "{\"message\":%s}", "internal server error")
	if err != nil {
		log.Println("error occurred while writing response")
	}
}
