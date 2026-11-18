package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"my_screening_app/internal/model"
	"my_screening_app/internal/repository"
)

// ErrInvalidInput is returned when the input to a service method is invalid.
var ErrInvalidInput = errors.New("invalid input")

// ItemService contém as regras de negócio de itens.
type ItemService struct {
	repo repository.ItemRepository
}

// NewItemService cria uma nova instância de ItemService.
func NewItemService(repo repository.ItemRepository) *ItemService {
	return &ItemService{repo: repo}
}

// CreateItem valida a entrada e cria um item.
func (s *ItemService) CreateItem(ctx context.Context, name string) (model.Item, error) {
	if err := validateName(name); err != nil {
		return model.Item{}, err
	}
	item := model.Item{
		ID:        uuid.NewString(),
		Name:      name,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.repo.Create(ctx, &item); err != nil {
		return model.Item{}, fmt.Errorf("create item: %w", err)
	}
	return item, nil
}

// ListItems lista todos os itens.
func (s *ItemService) ListItems(ctx context.Context) ([]model.Item, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	return items, nil
}

func validateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("%w: name must not be empty", ErrInvalidInput)
	}
	if len([]rune(name)) > 200 {
		return fmt.Errorf("%w: name must be at most 200 characters", ErrInvalidInput)
	}
	return nil
}
