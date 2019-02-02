package api

import (
	"io"
	"net/http"
)

func init() {
	http.HandleFunc("/health", Health)
}

func Health(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "ok")
}
