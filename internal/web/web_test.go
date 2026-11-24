package web

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"my_screening_app/internal/model"
	"my_screening_app/internal/repository"
	"my_screening_app/internal/service"
)

// mockItemRepository implementa repository.ItemRepository para testes web.
type mockWebItemRepository struct {
	items   []model.Item
	createErr error
	listErr   error
}

func (m *mockWebItemRepository) Create(ctx context.Context, item *model.Item) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.items = append(m.items, *item)
	return nil
}

func (m *mockWebItemRepository) List(ctx context.Context) ([]model.Item, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.items, nil
}

var _ repository.ItemRepository = (*mockWebItemRepository)(nil)

func TestNew(t *testing.T) {
	repo := &mockWebItemRepository{}
	svc := service.NewItemService(repo)

	w, err := New("Minha App", svc)
	if err != nil {
		t.Fatalf("New() retornou erro: %v", err)
	}
	if w == nil {
		t.Fatal("New() retornou nil")
	}
	if w.appName != "Minha App" {
		t.Errorf("appName = %q, want %q", w.appName, "Minha App")
	}
}

func TestMux(t *testing.T) {
	repo := &mockWebItemRepository{}
	svc := service.NewItemService(repo)

	w, err := New("App", svc)
	if err != nil {
		t.Fatalf("New() retornou erro: %v", err)
	}

	mux := w.Mux()
	if mux == nil {
		t.Fatal("Mux() retornou nil")
	}

	// Verifica que as rotas estão registradas fazendo requests
	tests := []struct {
		name     string
		method   string
		path     string
		wantCode int
	}{
		{
			name:     "rota home",
			method:   http.MethodGet,
			path:     "/",
			wantCode: http.StatusOK,
		},
		{
			name:     "rota items",
			method:   http.MethodGet,
			path:     "/web/items",
			wantCode: http.StatusOK,
		},
		{
			name:     "rota inexistente retorna 404",
			method:   http.MethodGet,
			path:     "/nao-existe",
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Errorf("código = %d, want %d", rec.Code, tt.wantCode)
			}
		})
	}
}

func TestHome(t *testing.T) {
	repo := &mockWebItemRepository{}
	svc := service.NewItemService(repo)

	w, err := New("Minha App", svc)
	if err != nil {
		t.Fatalf("New() retornou erro: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	w.home(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("código = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Minha App") {
		t.Errorf("body não contém 'Minha App': %q", body)
	}
}

func TestRenderItems(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*mockWebItemRepository)
		wantCode  int
		wantBody  string
	}{
		{
			name: "renderiza itens com sucesso",
			setup: func(m *mockWebItemRepository) {
				m.items = []model.Item{
					{ID: "1", Name: "Item A"},
				}
			},
			wantCode: http.StatusOK,
			wantBody: "Item A",
		},
		{
			name: "renderiza lista vazia",
			setup: func(m *mockWebItemRepository) {
				m.items = []model.Item{}
			},
			wantCode: http.StatusOK,
			wantBody: "",
		},
		{
			name: "erro ao listar itens retorna 500",
			setup: func(m *mockWebItemRepository) {
				m.listErr = errors.New("falha no banco")
			},
			wantCode: http.StatusInternalServerError,
			wantBody: "could not get items",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockWebItemRepository{}
			tt.setup(repo)

			svc := service.NewItemService(repo)
			w, err := New("App", svc)
			if err != nil {
				t.Fatalf("New() retornou erro: %v", err)
			}

			req := httptest.NewRequest(http.MethodGet, "/web/items", nil)
			rec := httptest.NewRecorder()

			w.renderItems(rec, req)

			if rec.Code != tt.wantCode {
				t.Errorf("código = %d, want %d", rec.Code, tt.wantCode)
			}
			body := rec.Body.String()
			if tt.wantBody != "" && !strings.Contains(body, tt.wantBody) {
				t.Errorf("body = %q, quer contém %q", body, tt.wantBody)
			}
		})
	}
}

func TestStatic(t *testing.T) {
	h := http.StripPrefix("/static/", Static())
	if h == nil {
		t.Fatal("Static() retornou nil handler")
	}

	req := httptest.NewRequest(http.MethodGet, "/static/css/style.css", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("código = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestHomeTemplateError(t *testing.T) {
	repo := &mockWebItemRepository{}
	svc := service.NewItemService(repo)

	// Cria um Web sem templates parseados para forçar erro no ExecuteTemplate
	w := &Web{
		appName: "App",
		tmpl:    template.New("vazio"),
		items:   svc,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	w.home(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("código = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestRenderItemsTemplateError(t *testing.T) {
	repo := &mockWebItemRepository{}
	svc := service.NewItemService(repo)

	w := &Web{
		appName: "App",
		tmpl:    template.New("vazio"),
		items:   svc,
	}

	req := httptest.NewRequest(http.MethodGet, "/web/items", nil)
	rec := httptest.NewRecorder()

	w.renderItems(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("código = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
