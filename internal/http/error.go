package http

import (
    "net/http"
)

type ErrorResponse struct {
    Error string `json:"error"`
	Meta      map[string]string `json:"meta,omitempty"`
}

func WriteError(w http.ResponseWriter, status int, msg string, meta map[string]string) {
    WriteJSON(w, status, ErrorResponse{Error: msg, Meta: meta})
}

func BadRequest(w http.ResponseWriter, msg string, meta map[string]string) {
    WriteError(w, http.StatusBadRequest, msg, meta)
}

func Unauthorized(w http.ResponseWriter, msg string, meta map[string]string) {
    WriteError(w, http.StatusUnauthorized, msg, meta)
}

func Forbidden(w http.ResponseWriter, msg string, meta map[string]string) {
    WriteError(w, http.StatusForbidden, msg, meta)
}

func NotFound(w http.ResponseWriter, msg string, meta map[string]string) {
    WriteError(w, http.StatusNotFound, msg, meta)
}

func InternalError(w http.ResponseWriter, msg string, meta map[string]string) {
    WriteError(w, http.StatusInternalServerError, msg, meta)
}

func TooManyRequests(w http.ResponseWriter, msg string, meta map[string]string) {
    WriteError(w, http.StatusTooManyRequests, msg, meta)
}