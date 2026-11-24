package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"my_screening_app/internal/model"
	"my_screening_app/internal/repository"
	"my_screening_app/internal/service"
)

// mockItemRepository implementa repository.ItemRepository para testes.
type mockItemRepository struct {
	items      []model.Item
	createErr  error
	listErr    error
}

func (m *mockItemRepository) Create(ctx context.Context, item *model.Item) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.items = append(m.items, *item)
	return nil
}

func (m *mockItemRepository) List(ctx context.Context) ([]model.Item, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.items, nil
}

// compile-time check
var _ repository.ItemRepository = (*mockItemRepository)(nil)

func TestHandler_ListItems(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*mockItemRepository)
		wantCode int
		wantBody string
	}{
		{
			name:      "lista itens com sucesso",
			setup: func(m *mockItemRepository) {
				m.items = []model.Item{
					{ID: "1", Name: "Item A"},
					{ID: "2", Name: "Item B"},
				}
			},
			wantCode: http.StatusOK,
			wantBody: `"Item A"`,
		},
		{
			name:      "lista vazia quando não há itens",
			setup: func(m *mockItemRepository) {
				m.items = []model.Item{}
			},
			wantCode: http.StatusOK,
			wantBody: `[]`,
		},
		{
			name:      "erro ao listar itens",
			setup: func(m *mockItemRepository) {
				m.listErr = errors.New("falha no banco")
			},
			wantCode: http.StatusInternalServerError,
			wantBody: `{"error":"could not list items"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockItemRepository{}
			tt.setup(repo)

			svc := service.NewItemService(repo)
			h := NewHandler(svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = &http.Request{}

			h.ListItems(c)

			if w.Code != tt.wantCode {
				t.Errorf("código = %d, want %d", w.Code, tt.wantCode)
			}
			body := w.Body.String()
			if tt.wantBody != "" && !strings.Contains(body, tt.wantBody) {
				t.Errorf("body = %q, quer contém %q", body, tt.wantBody)
			}
		})
	}
}

func TestHandler_CreateItem(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*mockItemRepository)
		body      string
		wantCode  int
		wantBody  string
	}{
		{
			name:      "cria item com sucesso",
			setup:     func(m *mockItemRepository) {},
			body:      `{"name":"Novo Item"}`,
			wantCode:  http.StatusCreated,
			wantBody:  `"Novo Item"`,
		},
		{
			name:      "nome vazio retorna 400",
			setup:     func(m *mockItemRepository) {},
			body:      `{"name":""}`,
			wantCode:  http.StatusBadRequest,
			wantBody:  `"error"`,
		},
		{
			name:      "corpo inválido retorna 400",
			setup:     func(m *mockItemRepository) {},
			body:      `não é json`,
			wantCode:  http.StatusBadRequest,
			wantBody:  `"error"`,
		},
		{
			name:      "erro no repositório retorna 500",
			setup: func(m *mockItemRepository) {
				m.createErr = errors.New("falha ao inserir")
			},
			body:      `{"name":"Item"}`,
			wantCode:  http.StatusInternalServerError,
			wantBody:  `"error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockItemRepository{}
			tt.setup(repo)

			svc := service.NewItemService(repo)
			h := NewHandler(svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = &http.Request{
				Method: http.MethodPost,
				Body:   io.NopCloser(strings.NewReader(tt.body)),
				Header: http.Header{"Content-Type": []string{"application/json"}},
			}

			h.CreateItem(c)

			if w.Code != tt.wantCode {
				t.Errorf("código = %d, want %d", w.Code, tt.wantCode)
			}
			body := w.Body.String()
			if tt.wantBody != "" && !strings.Contains(body, tt.wantBody) {
				t.Errorf("body = %q, quer contém %q", body, tt.wantBody)
			}
		})
	}
}
