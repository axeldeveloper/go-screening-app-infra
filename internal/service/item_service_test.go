package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"my_screening_app/internal/model"
	"my_screening_app/internal/repository"
)

func TestValidateName(t *testing.T) {
	longName := strings.Repeat("a", 201)

	tests := []struct {
		name string
		want error
	}{
		{name: "abc", want: nil},
		{name: "  spaced name  ", want: nil},
		{name: "", want: ErrInvalidInput},
		{name: "   ", want: ErrInvalidInput},
		{name: longName, want: ErrInvalidInput},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateName(tt.name)
			if tt.want == nil {
				if err != nil {
					t.Errorf("validateName(%q) = %v, want nil", tt.name, err)
				}
				return
			}
			if !errors.Is(err, tt.want) {
				t.Errorf("validateName(%q) = %v, want %v", tt.name, err, tt.want)
			}
		})
	}
}

// mockRepo implementa repository.ItemRepository para testes de service.
type mockRepo struct {
	items   []model.Item
	createErr error
	listErr   error
}

func (m *mockRepo) Create(ctx context.Context, item *model.Item) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.items = append(m.items, *item)
	return nil
}

func (m *mockRepo) List(ctx context.Context) ([]model.Item, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.items, nil
}

var _ repository.ItemRepository = (*mockRepo)(nil)

func TestItemService_CreateItem(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*mockRepo)
		nameArg string
		wantErr error
		wantItem model.Item
	}{
		{
			name:    "cria item com sucesso",
			setup:   func(m *mockRepo) {},
			nameArg: "Meu Item",
			wantErr: nil,
			wantItem: model.Item{
				Name: "Meu Item",
			},
		},
		{
			name:    "nome vazio retorna erro",
			setup:   func(m *mockRepo) {},
			nameArg: "",
			wantErr: ErrInvalidInput,
		},
		{
			name:    "nome apenas espaços retorna erro",
			setup:   func(m *mockRepo) {},
			nameArg: "   ",
			wantErr: ErrInvalidInput,
		},
		{
			name:    "nome muito longo retorna erro",
			setup:   func(m *mockRepo) {},
			nameArg: strings.Repeat("a", 201),
			wantErr: ErrInvalidInput,
		},
		{
			name:    "erro no repositório propaga",
			setup: func(m *mockRepo) {
				m.createErr = errors.New("falha ao inserir")
			},
			nameArg: "Item",
			wantErr: errors.New("create item: falha ao inserir"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{}
			tt.setup(repo)

			svc := NewItemService(repo)
			item, err := svc.CreateItem(context.Background(), tt.nameArg)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("esperava erro, não recebeu nenhum")
				}
				if errors.Is(err, ErrInvalidInput) {
					if !errors.Is(err, tt.wantErr) {
						t.Errorf("erro = %v, want %v", err, tt.wantErr)
					}
				} else if err.Error() != tt.wantErr.Error() {
					t.Errorf("erro = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("não esperava erro, recebeu: %v", err)
			}
			if item.Name != tt.wantItem.Name {
				t.Errorf("item.Name = %q, want %q", item.Name, tt.wantItem.Name)
			}
			if item.ID == "" {
				t.Error("item.ID não deve estar vazio (UUID gerado)")
			}
			if item.CreatedAt.IsZero() {
				t.Error("item.CreatedAt não deve ser zero")
			}
		})
	}
}

func TestItemService_ListItems(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*mockRepo)
		wantErr error
		wantLen int
	}{
		{
			name:    "lista itens com sucesso",
			setup: func(m *mockRepo) {
				m.items = []model.Item{
					{ID: "1", Name: "Item A"},
					{ID: "2", Name: "Item B"},
				}
			},
			wantErr: nil,
			wantLen: 2,
		},
		{
			name:    "lista vazia",
			setup:   func(m *mockRepo) {},
			wantErr: nil,
			wantLen: 0,
		},
		{
			name:    "erro no repositório propaga",
			setup: func(m *mockRepo) {
				m.listErr = errors.New("falha ao listar")
			},
			wantErr: errors.New("list items: falha ao listar"),
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{}
			tt.setup(repo)

			svc := NewItemService(repo)
			items, err := svc.ListItems(context.Background())

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("esperava erro, não recebeu nenhum")
				}
				if errors.Is(err, ErrInvalidInput) {
					if !errors.Is(err, tt.wantErr) {
						t.Errorf("erro = %v, want %v", err, tt.wantErr)
					}
				} else if err.Error() != tt.wantErr.Error() {
					t.Errorf("erro = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("não esperava erro, recebeu: %v", err)
			}
			if len(items) != tt.wantLen {
				t.Errorf("len(items) = %d, want %d", len(items), tt.wantLen)
			}
		})
	}
}
