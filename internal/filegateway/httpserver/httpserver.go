package httpserver

import (
	"context"
	"fmt"
	"github.com/denismitr/shardstore/internal/filegateway/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"mime/multipart"
	"net/http"
)

type processor interface {
	Process(
		ctx context.Context,
		f multipart.File,
		h *multipart.FileHeader,
	) error
}

type Server struct {
	cfg    *config.Config
	router *chi.Mux
	p      processor
}

func NewServer(cfg *config.Config, p processor) *Server {
	s := &Server{cfg: cfg, p: p}
	s.setupRoutes()
	return s
}

func (s *Server) uploadFile(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File Upload Endpoint Hit")

	if err := r.ParseMultipartForm(s.cfg.MaxFileSize); err != nil {
		fmt.Fprintf(w, "error parsing updloaded file: %s", err.Error())
		http.Error(w, http.StatusText(400), 400)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, http.StatusText(400), 400)
		fmt.Fprintf(w, "error retrieving updloaded file: %s", err.Error())
		return
	}
	defer file.Close()

	fmt.Printf("Uploaded File: %+v\n", header.Filename)
	fmt.Printf("File Size: %+v\n", header.Size)
	fmt.Printf("MIME Header: %+v\n", header.Header)

	if err := s.p.Process(r.Context(), file, header); err != nil {
		http.Error(w, http.StatusText(500), 500)
		fmt.Fprintf(w, "error processing updloaded file: %s", err.Error())
		return
	}

	fmt.Fprintf(w, "Successfully Uploaded File\n")
}

func (s *Server) setupRoutes() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Put("/upload/{bucket:[1-9-a-z-]+}", s.uploadFile)
	s.router = r
}

func (s *Server) Start() error {
	return http.ListenAndServe(fmt.Sprintf(":%d", s.cfg.HTTPPort), s.router)
}
