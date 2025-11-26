package http

import (
    "encoding/json"
    "net/http"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)

    enc := json.NewEncoder(w)
    _ = enc.Encode(v) // safe to ignore, client probably disconnected
}
