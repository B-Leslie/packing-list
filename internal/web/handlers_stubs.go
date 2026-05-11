package web

import "net/http"

func notImpl(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotImplemented) }

func (s *Server) getTrash(w http.ResponseWriter, r *http.Request)             { notImpl(w, r) }
func (s *Server) restore(w http.ResponseWriter, r *http.Request)              { notImpl(w, r) }
func (s *Server) purge(w http.ResponseWriter, r *http.Request)                { notImpl(w, r) }
func (s *Server) export(w http.ResponseWriter, r *http.Request)               { notImpl(w, r) }
func (s *Server) importJSON(w http.ResponseWriter, r *http.Request)           { notImpl(w, r) }
