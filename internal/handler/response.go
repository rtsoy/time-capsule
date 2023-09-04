package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"time-capsule/internal/service"
)

var statusCodes = map[error]int{
	service.ErrDBFailure:           http.StatusInternalServerError, // 500
	service.ErrPasswordHashFailure: http.StatusInternalServerError,

	service.ErrUsernameDuplicate: http.StatusConflict, // 409
	service.ErrEmailDuplicate:    http.StatusConflict,

	service.ErrNotFound: http.StatusNotFound, // 404

	service.ErrForbidden: http.StatusForbidden,

	service.ErrInvalidToken:       http.StatusUnauthorized, // 401
	service.ErrInvalidCredentials: http.StatusUnauthorized,
	service.ErrTokenExpired:       http.StatusUnauthorized,

	service.ErrInvalidTime:      http.StatusBadRequest, // 400
	service.ErrInvalidEmail:     http.StatusBadRequest,
	service.ErrInvalidUsername:  http.StatusBadRequest,
	service.ErrInvalidPassword:  http.StatusBadRequest,
	service.ErrShortMessage:     http.StatusBadRequest,
	service.ErrOpenTimeTooEarly: http.StatusBadRequest,
}

type errorResponse struct {
	Message string `json:"message"`
}

func newErrorResponse(w http.ResponseWriter, err error, code ...int) {
	resp := errorResponse{Message: err.Error()}
	bytes, _ := json.Marshal(resp)

	var (
		statusCode int
		ok         bool
	)

	if len(code) < 1 {
		statusCode, ok = statusCodes[err]
		if !ok {
			statusCode = http.StatusInternalServerError
		}
	} else {
		statusCode = code[0]
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(bytes)
}

func newJSONResponse(w http.ResponseWriter, v any, code ...int) {
	bytes, _ := json.Marshal(v)

	var statusCode int
	if len(code) < 1 {
		statusCode = http.StatusOK
	} else {
		statusCode = code[0]
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(bytes)
}

func handleRequestError(w http.ResponseWriter, err error) {
	if errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF) {
		newErrorResponse(w, errors.New("invalid json"), http.StatusBadRequest)
		return
	}

	switch err.(type) {
	case *time.ParseError:
		newErrorResponse(w, errors.New("invalid timestamp"), http.StatusBadRequest)
		return
	case *json.SyntaxError, *json.UnmarshalTypeError, *json.InvalidUnmarshalError, *json.UnsupportedValueError:
		newErrorResponse(w, errors.New("invalid json"), http.StatusBadRequest)
		return
	default:
		log.Println(err)
		newErrorResponse(w, errors.New("internal server error"), http.StatusInternalServerError)
		return
	}
}
