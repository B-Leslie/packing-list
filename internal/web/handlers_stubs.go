package web

import "net/http"

func notImpl(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotImplemented) }

