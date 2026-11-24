package model

import (
	"testing"
	"time"
)

func TestItem(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name  string
		item  Item
		want  Item
		check func(t *testing.T, got Item)
	}{
		{
			name: "item com todos os campos preenchidos",
			item: Item{
				ID:        "uuid-teste-123",
				Name:      "Meu Item",
				CreatedAt: now,
			},
			want: Item{
				ID:        "uuid-teste-123",
				Name:      "Meu Item",
				CreatedAt: now,
			},
			check: func(t *testing.T, got Item) {
				if got.ID != "uuid-teste-123" {
					t.Errorf("ID = %q, want %q", got.ID, "uuid-teste-123")
				}
				if got.Name != "Meu Item" {
					t.Errorf("Name = %q, want %q", got.Name, "Meu Item")
				}
				if !got.CreatedAt.Equal(now) {
					t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, now)
				}
			},
		},
		{
			name: "item com nome vazio",
			item: Item{
				ID:        "uuid-teste-456",
				Name:      "",
				CreatedAt: now,
			},
			want: Item{
				ID:        "uuid-teste-456",
				Name:      "",
				CreatedAt: now,
			},
			check: func(t *testing.T, got Item) {
				if got.Name != "" {
					t.Errorf("Name = %q, want vazio", got.Name)
				}
			},
		},
		{
			name: "item zero value",
			item: Item{},
			want: Item{},
			check: func(t *testing.T, got Item) {
				if got.ID != "" {
					t.Errorf("ID = %q, want vazio", got.ID)
				}
				if got.Name != "" {
					t.Errorf("Name = %q, want vazio", got.Name)
				}
				if !got.CreatedAt.IsZero() {
					t.Errorf("CreatedAt = %v, want zero", got.CreatedAt)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.item

			if got.ID != tt.want.ID {
				t.Errorf("ID = %q, want %q", got.ID, tt.want.ID)
			}
			if got.Name != tt.want.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.want.Name)
			}
			if !got.CreatedAt.Equal(tt.want.CreatedAt) {
				t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, tt.want.CreatedAt)
			}

			tt.check(t, got)
		})
	}
}
