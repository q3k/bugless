package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
)

func serveHashedStatic(mux *http.ServeMux, name, contentType string, data []byte) string {
	h := sha256.New()
	h.Write(data)
	hash := hex.EncodeToString(h.Sum(nil))

	path := fmt.Sprintf("/static/%s/%s", hash, name)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Write(data)
	})

	return path
}
