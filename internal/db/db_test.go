package db

import (
	"os"
	"testing"
)

func TestDatabaseURL(t *testing.T) {
	tests := []struct {
		name     string
		env      string
		setEnv   bool
		want     string
	}{
		{
			name:    "usa DATABASE_URL do ambiente",
			env:     "postgres://usuario:senha@localhost:5432/banco?sslmode=disable",
			setEnv:  true,
			want:    "postgres://usuario:senha@localhost:5432/banco?sslmode=disable",
		},
		{
			name:    "retorna valor padrão quando DATABASE_URL não definida",
			setEnv:  false,
			want:    "postgres://postgres:pg766312@localhost:5433/my_screening_app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv("DATABASE_URL", tt.env)
			} else {
				os.Unsetenv("DATABASE_URL")
			}

			got := databaseURL()
			if got != tt.want {
				t.Errorf("databaseURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
