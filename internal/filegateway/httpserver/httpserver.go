package httpserver

import (
	"context"
	"fmt"
	"github.com/denismitr/shardstore/internal/common/logger"
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
	lg     logger.Logger
	router *chi.Mux
	p      processor
}

func NewServer(cfg *config.Config, lg logger.Logger, p processor) *Server {
	s := &Server{cfg: cfg, p: p, lg: lg}
	s.setupRoutes()
	return s
}

func (s *Server) uploadFile(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File Upload Endpoint Hit")

	if err := r.ParseMultipartForm(s.cfg.MaxFileSize); err != nil {
		s.lg.Error(fmt.Errorf("error parsing updloaded file: %w", err))
		http.Error(w, http.StatusText(400), 400)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		s.lg.Error(fmt.Errorf("error retrieving updloaded file: %w", err))
		http.Error(w, http.StatusText(400), 400)
		return
	}

	defer func() {
		if err := file.Close(); err != nil {
			s.lg.Error(err)
		}
	}()

	s.lg.Debugf("Uploaded File: %+v\n", header.Filename)
	s.lg.Debugf("File Size: %+v\n", header.Size)
	s.lg.Debugf("MIME Header: %+v\n", header.Header)

	if err := s.p.Process(r.Context(), file, header); err != nil {
		http.Error(w, http.StatusText(500), 500)
		s.lg.Error(fmt.Errorf("error processing updloaded file: %w", err))
		return
	}

	s.lg.Debugf("Successfully Uploaded File")
}

func (s *Server) setupRoutes() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Put("/upload", s.uploadFile)
	s.router = r
}

func (s *Server) Start() error {
	return http.ListenAndServe(fmt.Sprintf(":%d", s.cfg.HTTPPort), s.router)
}
