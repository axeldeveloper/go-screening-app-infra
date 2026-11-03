package web

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"

	"my_tristack_app/internal/model"
	"my_tristack_app/internal/service"
)

//go:embed templates static
var content embed.FS

type Web struct {
	appName string
	tmpl    *template.Template
	items   *service.ItemService
}

func New(appName string, items *service.ItemService) (*Web, error) {
	tmpl, err := template.ParseFS(content, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse web templates: %w", err)
	}
	return &Web{
		appName: appName,
		tmpl:    tmpl,
		items:   items,
	}, nil
}

func (w *Web) Mux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", w.home)
	mux.HandleFunc("GET /web/items", w.renderItems)
	mux.Handle("GET /static/", http.StripPrefix("/static/", Static()))
	return mux
}

func Static() http.Handler {
	sub, err := fs.Sub(content, "static")
	if err != nil {
		panic(fmt.Errorf("missing static assets: %w", err))
	}
	return http.FileServer(http.FS(sub))
}

type homeData struct {
	AppName string
}

func (w *Web) home(wr http.ResponseWriter, r *http.Request) {
	page := homeData{AppName: w.appName}
	if err := w.tmpl.ExecuteTemplate(wr, "index.html", page); err != nil {
		http.Error(wr, "could not render home", http.StatusInternalServerError)
		return
	}
}

type itemsData struct {
	Items []model.Item
}

func (w *Web) renderItems(wr http.ResponseWriter, r *http.Request) {
	items, err := w.items.ListItems(r.Context())
	if err != nil {
		http.Error(wr, "could not get items", http.StatusInternalServerError)
		return
	}
	if err := w.tmpl.ExecuteTemplate(wr, "items.html", itemsData{Items: items}); err != nil {
		http.Error(wr, "could not render items", http.StatusInternalServerError)
		return
	}
}
