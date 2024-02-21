package handler

import (
	"net/http"
)

func Status(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("{ status: 'okay' }"))
}
