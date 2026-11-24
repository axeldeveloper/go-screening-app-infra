package main

import (
	"testing"
)

func TestCategorize(t *testing.T) {
	tests := []struct {
		name   string
		linter string
		rule   string
		want   category
	}{
		{"gosec", "gosec", "G404", catSecurity},
		{"govet bug", "govet", "composites", catBug},
		{"errcheck bug", "errcheck", "assign", catBug},
		{"staticcheck SA bug", "staticcheck", "SA1000", catBug},
		{"staticcheck non-SA smell", "staticcheck", "ST1000", catSmell},
		{"revive smell", "revive", "exported", catSmell},
		{"unknown linter smell", "unknown", "rule", catSmell},
		{"empty linter smell", "", "", catSmell},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := categorize(tt.linter, tt.rule)
			if got != tt.want {
				t.Errorf("categorize(%q, %q) = %v, want %v", tt.linter, tt.rule, got, tt.want)
			}
		})
	}
}

func TestCatClassOf(t *testing.T) {
	tests := []struct {
		cat  category
		want string
	}{
		{catSecurity, "sec"},
		{catBug, "bug"},
		{catSmell, "smell"},
	}

	for _, tt := range tests {
		t.Run(tt.cat.String(), func(t *testing.T) {
			got := catClassOf(tt.cat)
			if got != tt.want {
				t.Errorf("catClassOf(%v) = %q, want %q", tt.cat, got, tt.want)
			}
		})
	}
}

func TestCategoryString(t *testing.T) {
	if catSecurity.String() != "Segurança" {
		t.Errorf("catSecurity.String() = %q, want %q", catSecurity.String(), "Segurança")
	}
	if catBug.String() != "Bugs" {
		t.Errorf("catBug.String() = %q, want %q", catBug.String(), "Bugs")
	}
	if catSmell.String() != "Manutenção" {
		t.Errorf("catSmell.String() = %q, want %q", catSmell.String(), "Manutenção")
	}
}

func TestCut(t *testing.T) {
	tests := []struct {
		name string
		s    string
		n    int
		want string
	}{
		{"curto", "oi", 10, "oi"},
		{"exato", "abcde", 5, "abcde"},
		{"longo", "abcdefghij", 5, "abcd…"},
		{"vazio", "", 5, ""},
		{"um caractere", "a", 1, "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cut(tt.s, tt.n)
			if got != tt.want {
				t.Errorf("cut(%q, %d) = %q, want %q", tt.s, tt.n, got, tt.want)
			}
		})
	}
}

func TestGrade(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want string
	}{
		{"zero", 0, "A"},
		{"um", 1, "B"},
		{"tres", 3, "B"},
		{"quatro", 4, "C"},
		{"dez", 10, "C"},
		{"onze", 11, "D"},
		{"vintecinco", 25, "D"},
		{"vinteseis", 26, "E"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := grade(tt.n)
			if got != tt.want {
				t.Errorf("grade(%d) = %q, want %q", tt.n, got, tt.want)
			}
		})
	}
}

func TestLimitText(t *testing.T) {
	tests := []struct {
		name string
		l    int
		want string
	}{
		{"sem limite", -1, "sem limite"},
		{"zero", 0, "0"},
		{"cinquenta", 50, "50"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := limitText(tt.l)
			if got != tt.want {
				t.Errorf("limitText(%d) = %q, want %q", tt.l, got, tt.want)
			}
		})
	}
}

func TestVerdict(t *testing.T) {
	if verdict(true) != "APROVADO" {
		t.Errorf("verdict(true) = %q, want %q", verdict(true), "APROVADO")
	}
	if verdict(false) != "REPROVADO" {
		t.Errorf("verdict(false) = %q, want %q", verdict(false), "REPROVADO")
	}
}

func TestIsCommitSHA(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"sha valido", "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2", true},
		{"muito curto", "abc", false},
		{"muito longo", "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3", false},
		{"nao hexadecimal", "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1bz", false},
		{"vazio", "", false},
		{"maiusculas", "A1B2C3D4E5F6A1B2C3D4E5F6A1B2C3D4E5F6A1B2", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCommitSHA(tt.s)
			if got != tt.want {
				t.Errorf("isCommitSHA(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestNormalizeRef(t *testing.T) {
	tests := []struct {
		name string
		r    string
		want string
	}{
		{"vazio", "", ""},
		{"GC prefix", "GCabc123", "GCabc123"},
		{"GB prefix", "GBmain", "GBmain"},
		{"GT prefix", "GTtag", "GTtag"},
		{"sha completo", "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2", "GCa1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"},
		{"branch nome", "main", "GBmain"},
		{"branch com barra", "feature/x", "GBfeature/x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeRef(tt.r)
			if got != tt.want {
				t.Errorf("normalizeRef(%q) = %q, want %q", tt.r, got, tt.want)
			}
		})
	}
}

func TestSplitRule(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		wantRule string
		wantMsg  string
	}{
		{"com regra", "exported: Item público", "exported", "Item público"},
		{"sem regra", "mensagem simples", "", "mensagem simples"},
		{"vazio", "", "", ""},
		{"apenas dois pontos", ": msg", "", ": msg"},
		{"espaco antes dos pontos", "regra : msg", "", "regra : msg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRule, gotMsg := splitRule(tt.text)
			if gotRule != tt.wantRule {
				t.Errorf("splitRule(%q) rule = %q, want %q", tt.text, gotRule, tt.wantRule)
			}
			if gotMsg != tt.wantMsg {
				t.Errorf("splitRule(%q) msg = %q, want %q", tt.text, gotMsg, tt.wantMsg)
			}
		})
	}
}

func TestFingerprint(t *testing.T) {
	f := finding{
		Linter:  "revive",
		Rule:    "exported",
		Message: "Item público sem comentário",
		File:    "main.go",
		Snippet: "",
	}
	fp1 := fingerprint(f)
	fp2 := fingerprint(f)

	if fp1 == "" {
		t.Error("fingerprint não deve ser vazio")
	}
	if fp1 != fp2 {
		t.Error("fingerprint deve ser determinístico")
	}
	if len(fp1) != 32 {
		t.Errorf("fingerprint deve ter 32 chars (hex de 16 bytes), tem %d", len(fp1))
	}
}

func TestStatusOf(t *testing.T) {
	tests := []struct {
		name      string
		check     check
		wantText  string
		wantClass string
	}{
		{"aprovado", check{Name: "x", Actual: 0, Limit: 10, OK: true}, "ok", "st-ok"},
		{"falhou", check{Name: "x", Actual: 11, Limit: 10, OK: false}, "FALHOU", "st-bad"},
		{"sem limite com itens", check{Name: "x", Actual: 5, Limit: -1, OK: true}, "aviso", "st-warn"},
		{"sem limite sem itens", check{Name: "x", Actual: 0, Limit: -1, OK: true}, "ok", "st-ok"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotText, gotClass := statusOf(tt.check)
			if gotText != tt.wantText {
				t.Errorf("statusOf text = %q, want %q", gotText, tt.wantText)
			}
			if gotClass != tt.wantClass {
				t.Errorf("statusOf class = %q, want %q", gotClass, tt.wantClass)
			}
		})
	}
}

func TestEvaluate(t *testing.T) {
	tests := []struct {
		name     string
		counts   [numCats]int
		limits   [numCats]int
		wantPass bool
	}{
		{"todos dentro do limite", [numCats]int{0, 0, 0}, [numCats]int{10, 10, 10}, true},
		{"um acima do limite", [numCats]int{11, 0, 0}, [numCats]int{10, 10, 10}, false},
		{"sem limite", [numCats]int{100, 0, 0}, [numCats]int{-1, 10, 10}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checks, passed := evaluate(tt.counts, tt.limits[0], tt.limits[1], tt.limits[2])
			if passed != tt.wantPass {
				t.Errorf("evaluate passed = %v, want %v", passed, tt.wantPass)
			}
			if len(checks) != numCats {
				t.Errorf("evaluate checks len = %d, want %d", len(checks), numCats)
			}
		})
	}
}
