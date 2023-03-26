package httpserver

import (
	"context"
	"fmt"
	"github.com/denismitr/shardstore/internal/common/logger"
	"github.com/denismitr/shardstore/internal/filegateway/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"mime/multipart"
	"net/http"
)

type fileUploader interface {
	Upload(
		ctx context.Context,
		f multipart.File,
		h *multipart.FileHeader,
	) error
}

type fileDownloader interface {
	Download(
		ctx context.Context,
		fileName string,
		w io.Writer,
	) (int, error)
}

type Server struct {
	cfg        *config.Config
	lg         logger.Logger
	router     *chi.Mux
	uploader   fileUploader
	downloader fileDownloader
}

func NewServer(
	cfg *config.Config,
	lg logger.Logger,
	fu fileUploader,
	fd fileDownloader,
) *Server {
	s := &Server{cfg: cfg, uploader: fu, lg: lg, downloader: fd}
	s.setupRoutes()
	return s
}

func (s *Server) downloadFile(w http.ResponseWriter, r *http.Request) {
	file := chi.URLParam(r, "file")
	w.WriteHeader(200)
	w.Header().Set("content-type", "application/force-download") // todo: maybe to store mime types
	w.Header().Set("content-disposition", `attachment; filename="`+file+`"`)
	//w.Header().Set("content-length", fmt.Sprintf("%d", downloaded))

	_, err := s.downloader.Download(r.Context(), file, w)
	if err != nil {
		// todo: check not found error
		http.Error(w, http.StatusText(500), 500)
		return
	}
}

func (s *Server) uploadFile(w http.ResponseWriter, r *http.Request) {
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

	s.lg.Debugf("uploaded file name: %s\n", header.Filename)
	s.lg.Debugf("file size: %d\n", header.Size)
	s.lg.Debugf("MIME header: %+v\n", header.Header)

	if err := s.uploader.Upload(r.Context(), file, header); err != nil {
		http.Error(w, http.StatusText(500), 500)
		s.lg.Error(fmt.Errorf("error processing updloaded file: %w", err))
		return
	}

	s.lg.Debugf("successfully uploaded file")
}

func (s *Server) setupRoutes() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Put("/files/upload", s.uploadFile)
	r.Get("/files/{file}", s.downloadFile)
	s.router = r
}

func (s *Server) Start() error {
	return http.ListenAndServe(fmt.Sprintf(":%d", s.cfg.HTTPPort), s.router)
}
