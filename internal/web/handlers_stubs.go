package web

import "net/http"

func notImpl(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotImplemented) }

func (s *Server) getTripsIndex(w http.ResponseWriter, r *http.Request)        { notImpl(w, r) }
func (s *Server) getTripNew(w http.ResponseWriter, r *http.Request)           { notImpl(w, r) }
func (s *Server) postTripCreate(w http.ResponseWriter, r *http.Request)       { notImpl(w, r) }
func (s *Server) getTripDetail(w http.ResponseWriter, r *http.Request)        { notImpl(w, r) }
func (s *Server) patchTrip(w http.ResponseWriter, r *http.Request)            { notImpl(w, r) }
func (s *Server) deleteTrip(w http.ResponseWriter, r *http.Request)           { notImpl(w, r) }
func (s *Server) attachBundle(w http.ResponseWriter, r *http.Request)         { notImpl(w, r) }
func (s *Server) detachBundle(w http.ResponseWriter, r *http.Request)         { notImpl(w, r) }
func (s *Server) addExtra(w http.ResponseWriter, r *http.Request)             { notImpl(w, r) }
func (s *Server) overrideItem(w http.ResponseWriter, r *http.Request)         { notImpl(w, r) }
func (s *Server) togglePack(w http.ResponseWriter, r *http.Request)           { notImpl(w, r) }
func (s *Server) inviteMember(w http.ResponseWriter, r *http.Request)         { notImpl(w, r) }
func (s *Server) removeMember(w http.ResponseWriter, r *http.Request)         { notImpl(w, r) }
func (s *Server) listBundles(w http.ResponseWriter, r *http.Request)          { notImpl(w, r) }
func (s *Server) createBundle(w http.ResponseWriter, r *http.Request)         { notImpl(w, r) }
func (s *Server) editBundle(w http.ResponseWriter, r *http.Request)           { notImpl(w, r) }
func (s *Server) updateBundle(w http.ResponseWriter, r *http.Request)         { notImpl(w, r) }
func (s *Server) deleteBundle(w http.ResponseWriter, r *http.Request)         { notImpl(w, r) }
func (s *Server) bundleAddItem(w http.ResponseWriter, r *http.Request)        { notImpl(w, r) }
func (s *Server) bundleRemoveItem(w http.ResponseWriter, r *http.Request)     { notImpl(w, r) }
func (s *Server) bundleAddChild(w http.ResponseWriter, r *http.Request)       { notImpl(w, r) }
func (s *Server) bundleRemoveChild(w http.ResponseWriter, r *http.Request)    { notImpl(w, r) }
func (s *Server) getTrash(w http.ResponseWriter, r *http.Request)             { notImpl(w, r) }
func (s *Server) restore(w http.ResponseWriter, r *http.Request)              { notImpl(w, r) }
func (s *Server) purge(w http.ResponseWriter, r *http.Request)                { notImpl(w, r) }
func (s *Server) export(w http.ResponseWriter, r *http.Request)               { notImpl(w, r) }
func (s *Server) importJSON(w http.ResponseWriter, r *http.Request)           { notImpl(w, r) }
