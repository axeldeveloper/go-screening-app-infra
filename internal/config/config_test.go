package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name      string
		env       map[string]string
		wantApp   string
		wantPort  string
	}{
		{
			name:  "valores padrões quando variáveis não definidas",
			env:   map[string]string{},
			wantApp: "my-tristack-app",
			wantPort: "8000",
		},
		{
			name:  "valores do ambiente são respeitados",
			env:   map[string]string{"APP_NAME": "minha-app", "PORT": "9000"},
			wantApp: "minha-app",
			wantPort: "9000",
		},
		{
			name:  "apenas APP_NAME definida",
			env:   map[string]string{"APP_NAME": "outra-app"},
			wantApp: "outra-app",
			wantPort: "8000",
		},
		{
			name:  "apenas PORT definida",
			env:   map[string]string{"PORT": "3000"},
			wantApp: "my-tristack-app",
			wantPort: "3000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Limpa variáveis de ambiente do teste anterior
			for k := range tt.env {
				t.Setenv(k, tt.env[k])
			}
			// Remove variáveis não definidas neste caso de teste
			for _, k := range []string{"APP_NAME", "PORT"} {
				if _, ok := tt.env[k]; !ok {
					os.Unsetenv(k)
				}
			}

			cfg := Load()
			if cfg.AppName != tt.wantApp {
				t.Errorf("AppName = %q, want %q", cfg.AppName, tt.wantApp)
			}
			if cfg.Port != tt.wantPort {
				t.Errorf("Port = %q, want %q", cfg.Port, tt.wantPort)
			}
		})
	}
}

func TestGetenv(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		fallback string
		envVal  string
		setEnv  bool
		want    string
	}{
		{
			name:     "variável definida retorna seu valor",
			key:      "APP_NAME",
			fallback: "padrao",
			envVal:   "minha-app",
			setEnv:   true,
			want:     "minha-app",
		},
		{
			name:     "variável não definida retorna fallback",
			key:      "PORT",
			fallback: "8000",
			setEnv:   false,
			want:     "8000",
		},
		{
			name:     "variável vazia retorna fallback",
			key:      "APP_NAME",
			fallback: "padrao",
			envVal:   "",
			setEnv:   true,
			want:     "padrao",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(tt.key, tt.envVal)
			} else {
				os.Unsetenv(tt.key)
			}

			got := getenv(tt.key, tt.fallback)
			if got != tt.want {
				t.Errorf("getenv(%q, %q) = %q, want %q", tt.key, tt.fallback, got, tt.want)
			}
		})
	}
}
