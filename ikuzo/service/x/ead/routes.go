package ead

import "github.com/go-chi/chi"

func (s *Service) Routes(pattern string, r chi.Router) {
	// r.Delete("/api/datasets/{spec}", s.CancelTask)
	r.Delete("/api/ead/{spec}/mets/{UUID}", s.DaoClient.HandleDelete)

	r.Get("/api/ead/tasks", s.Tasks)
	r.Get("/api/ead/tasks/{id}", s.GetTask)
	r.Get("/api/ead/{spec}/mets/{UUID}", s.DaoClient.DownloadXML)
	r.Get("/api/ead/{spec}/mets/{UUID}.json", s.DaoClient.DownloadConfig)

	r.Post("/api/ead", s.handleUpload)
	r.Post("/api/ead/{spec}/mets/{UUID}", s.DaoClient.Index)
}
