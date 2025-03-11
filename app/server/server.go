package server

import (
	_ "embed"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"text/template"
)

//go:embed index.html
var indexHTML string
var indexTemplate = template.Must(template.New("index").Parse(indexHTML))

type PageData struct {
	Title                string
	BaseURL              string
	MainLinkText         string
	SecondaryLinkText    string
	MainDescription      string
	SecondaryDescription string
}

type Server struct {
	address    string
	baseURL    string
	content    string
	contentMux sync.RWMutex
	pageData   PageData
}

func NewServer(address, baseURL string) *Server {
	return &Server{
		address: address,
		baseURL: baseURL,
		pageData: PageData{
			Title:                "Hegering&nbspGronau-Epe | Termine",
			BaseURL:              baseURL,
			MainLinkText:         "Kalender abonnieren",
			SecondaryLinkText:    "Direkter Download (.ics)",
			MainDescription:      "Der Kalender kann in den meisten Kalender-Apps abonniert werden, dazu einfach auf den Link klicken. Der Kalender wird dann automatisch aktualisiert.",
			SecondaryDescription: "Sollte das Abonnieren nicht funktionieren, kann der Kalender auch direkt heruntergeladen werden.",
		},
	}
}

func (s *Server) UpdateContent(newContent string) {
	s.contentMux.Lock()
	defer s.contentMux.Unlock()
	s.content = newContent
}

func (s *Server) Start() error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		indexTemplate.Execute(w, s.pageData)
	})

	http.HandleFunc("/events.ics", func(w http.ResponseWriter, r *http.Request) {
		s.contentMux.RLock()
		content := s.content
		s.contentMux.RUnlock()

		slog.InfoContext(r.Context(), "serving iCal file",
			"User-Agent", r.Header.Get("User-Agent"),
			"RemoteAddr", r.RemoteAddr,
		)
		w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=events.ics")
		_, _ = w.Write([]byte(content))
	})

	fmt.Printf("Starting server on %s\n", s.address)
	return http.ListenAndServe(s.address, nil)
}
