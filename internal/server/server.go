package server

import (
	"context"
	"embed"
	"github.com/creekorful/go-news/internal/database"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

//go:embed res/*
var resDirectory embed.FS

type Server struct {
	db *database.Database

	srv *http.Server
}

type ctx struct {
	Items []database.Item
}

func NewServer(db *database.Database) *Server {
	return &Server{db: db}
}

func (s *Server) Serve(ctx context.Context, address string) error {
	s.srv = &http.Server{
		Addr:         address,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      http.HandlerFunc(s.index),
	}

	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	items, err := s.db.GetItems()
	if err != nil {
		log.Printf("error while rendering index: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	t, err := template.New("index.html.tmpl").ParseFS(resDirectory, filepath.Join("res", "index.html.tmpl"))
	if err != nil {
		log.Printf("error while rendering index: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := t.ExecuteTemplate(w, "index.html.tmpl", ctx{Items: items}); err != nil {
		log.Printf("error while rendering index: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
