// Command lintreport converte o JSON do golangci-lint em um relatório legível
// (console, HTML e Markdown) com um quality gate simplificado, no estilo do
// SonarQube. Coloque este arquivo em tools/lintreport/main.go.
//
// Uso:
//
//	golangci-lint run ./... --output.json.path=reports/lint.json
//	go run ./tools/lintreport -in reports/lint.json -out reports
//
// Comparação com execução anterior: a cada execução o programa grava
// <out>/snapshot.json. Na execução seguinte ele usa esse arquivo como base
// (ou o arquivo indicado em -baseline) e marca o que é novo e o que foi
// corrigido. Com -max-new-smells o gate também limita itens novos.
//
// Links para o Azure Repos: no pipeline são montados automaticamente; fora
// dele, use -repo-url e -ref.
//
// Códigos de saída: 0 = aprovado, 1 = reprovado no quality gate, 2 = erro.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const numCats = 3

// ---------------------------------------------------------------
// Entrada (JSON do golangci-lint)
// ---------------------------------------------------------------

type issue struct {
	FromLinter  string   `json:"FromLinter"`
	Text        string   `json:"Text"`
	SourceLines []string `json:"SourceLines"`
	Pos         struct {
		Filename string `json:"Filename"`
		Line     int    `json:"Line"`
		Column   int    `json:"Column"`
	} `json:"Pos"`
}

type report struct {
	Issues []issue `json:"Issues"`
}

// ---------------------------------------------------------------
// Classificação
// ---------------------------------------------------------------

type category int

const (
	catSecurity category = iota
	catBug
	catSmell
)

func (c category) String() string {
	return [...]string{"Segurança", "Bugs", "Manutenção"}[c]
}

func catClassOf(c category) string {
	return [...]string{"sec", "bug", "smell"}[c]
}

func categorize(linter, rule string) category {
	switch linter {
	case "gosec":
		return catSecurity
	case "govet", "errcheck", "errorlint", "bodyclose", "sqlclosecheck", "ineffassign":
		return catBug
	case "staticcheck":
		if strings.HasPrefix(rule, "SA") {
			return catBug
		}
	}
	return catSmell
}

// plainRules traduz regras comuns para uma frase que qualquer pessoa entende.
var plainRules = map[string]string{
	"exported":         "Item público sem comentário de documentação.",
	"package-comments": "Pacote sem comentário de documentação no formato esperado.",
	"unused-parameter": "Parâmetro nunca usado; renomeie para _ ou remova.",
	"var-naming":       "Nome fora do padrão de nomenclatura do Go.",
	"error-strings":    "Mensagem de erro fora do padrão (minúscula, sem pontuação final).",
	"G104":             "Erro retornado e ignorado sem tratamento.",
	"G114":             "Servidor HTTP sem timeouts configurados.",
	"G201":             "Consulta SQL montada com formatação de texto (risco de SQL injection).",
	"G202":             "Consulta SQL montada por concatenação (risco de SQL injection).",
	"G304":             "Abertura de arquivo com caminho vindo de variável.",
	"G306":             "Arquivo criado com permissões muito abertas.",
	"G404":             "Gerador de números aleatórios fraco para uso sensível.",
}

type finding struct {
	Category category
	Linter   string
	Rule     string
	Message  string
	Plain    string
	File     string
	Line     int
	Col      int
	Snippet  string
	FP       string // impressão digital estável, para comparar execuções
	New      bool   // não existia na execução usada como base
	URL      string // link para o código no Azure Repos (se disponível)
}

// Name devolve o nome da regra (ou do linter, se a regra não vier no texto).
func (f finding) Name() string {
	if f.Rule != "" {
		return f.Rule
	}
	return f.Linter
}

// Explain devolve a explicação em linguagem simples, ou a mensagem original.
func (f finding) Explain() string {
	if f.Plain != "" {
		return f.Plain
	}
	return f.Message
}

// CatID, CatClass e SearchText são usados pelo template HTML.
func (f finding) CatID() int { return int(f.Category) }

func (f finding) CatClass() string { return catClassOf(f.Category) }

func (f finding) SearchText() string {
	return strings.ToLower(strings.Join(
		[]string{f.File, f.Name(), f.Explain(), f.Message, f.Snippet, f.Category.String()}, " "))
}

// splitRule separa "regra: mensagem" quando o prefixo parece um nome de regra.
func splitRule(text string) (rule, msg string) {
	if i := strings.Index(text, ": "); i > 0 && !strings.ContainsAny(text[:i], " ()") {
		return text[:i], text[i+2:]
	}
	return "", text
}

// fingerprint não usa o número da linha, para que editar o arquivo acima do
// problema não o transforme em "novo".
func fingerprint(f finding) string {
	sum := sha256.Sum256([]byte(strings.Join(
		[]string{f.Linter, f.Name(), f.File, f.Message, f.Snippet}, "\x1f")))
	return hex.EncodeToString(sum[:16])
}

func classify(issues []issue) []finding {
	out := make([]finding, 0, len(issues))
	for _, is := range issues {
		rule, msg := splitRule(is.Text)
		f := finding{
			Category: categorize(is.FromLinter, rule),
			Linter:   is.FromLinter,
			Rule:     rule,
			Message:  msg,
			Plain:    plainRules[rule],
			File:     is.Pos.Filename,
			Line:     is.Pos.Line,
			Col:      is.Pos.Column,
		}
		if len(is.SourceLines) > 0 {
			f.Snippet = strings.TrimSpace(is.SourceLines[0])
		}
		f.FP = fingerprint(f)
		out = append(out, f)
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Category != b.Category {
			return a.Category < b.Category
		}
		if a.File != b.File {
			return a.File < b.File
		}
		return a.Line < b.Line
	})
	return out
}

// ---------------------------------------------------------------
// Quality gate
// ---------------------------------------------------------------

type check struct {
	Name   string
	Actual int
	Limit  int // -1 = sem limite (apenas aviso)
	OK     bool
}

func evaluate(counts [numCats]int, maxSec, maxBugs, maxSmells int) ([]check, bool) {
	limits := [numCats]int{maxSec, maxBugs, maxSmells}
	checks := make([]check, 0, len(counts)+1)
	passed := true
	for i := range counts {
		ok := limits[i] < 0 || counts[i] <= limits[i]
		if !ok {
			passed = false
		}
		checks = append(checks, check{category(i).String(), counts[i], limits[i], ok})
	}
	return checks, passed
}

// grade é uma nota simplificada (A a E) pela quantidade de itens.
func grade(n int) string {
	switch {
	case n == 0:
		return "A"
	case n <= 3:
		return "B"
	case n <= 10:
		return "C"
	case n <= 25:
		return "D"
	default:
		return "E"
	}
}

func statusOf(c check) (text, class string) {
	switch {
	case !c.OK:
		return "FALHOU", "st-bad"
	case c.Limit < 0 && c.Actual > 0:
		return "aviso", "st-warn"
	default:
		return "ok", "st-ok"
	}
}

func limitText(l int) string {
	if l < 0 {
		return "sem limite"
	}
	return fmt.Sprint(l)
}

func verdict(passed bool) string {
	if passed {
		return "APROVADO"
	}
	return "REPROVADO"
}

// ---------------------------------------------------------------
// Comparação com a execução anterior
// ---------------------------------------------------------------

type snapItem struct {
	FP      string `json:"fp"`
	Count   int    `json:"n"`
	File    string `json:"file"`
	Rule    string `json:"rule"`
	Message string `json:"msg"`
	Cat     int    `json:"cat"`
}

type snapshot struct {
	Version   int          `json:"version"`
	Generated string       `json:"generated"`
	Counts    [numCats]int `json:"counts"`
	Items     []snapItem   `json:"items"`
}

type diff struct {
	Has        bool
	Source     string
	PrevCounts [numCats]int
	New        int
	NewByCat   [numCats]int
	FixedCount int
	Fixed      []snapItem
}

func loadBaseline(path, outDir string) *snapshot {
	explicit := path != ""
	if !explicit {
		path = filepath.Join(outDir, "snapshot.json")
	}
	data, err := os.ReadFile(path) //nolint:gosec // caminho vem de flag da CLI ou da pasta de saída
	if err != nil {
		if explicit {
			fmt.Fprintf(os.Stderr, "lintreport: base %s não encontrada; seguindo sem comparação\n", path)
		}
		return nil
	}
	var s snapshot
	if err := json.Unmarshal(data, &s); err != nil || s.Version != 1 {
		fmt.Fprintf(os.Stderr, "lintreport: base %s inválida; ignorando\n", path)
		return nil
	}
	return &s
}

// compare marca como novos os achados que excedem a quantidade da base.
func compare(fs []finding, prev *snapshot) diff {
	if prev == nil {
		return diff{}
	}
	d := diff{Has: true, Source: prev.Generated, PrevCounts: prev.Counts}
	prevN := make(map[string]int, len(prev.Items))
	for _, it := range prev.Items {
		prevN[it.FP] = it.Count
	}
	seen := map[string]int{}
	for i := range fs {
		fp := fs[i].FP
		seen[fp]++
		if seen[fp] > prevN[fp] {
			fs[i].New = true
			d.New++
			d.NewByCat[fs[i].Category]++
		}
	}
	for _, it := range prev.Items {
		if n := it.Count - seen[it.FP]; n > 0 {
			d.FixedCount += n
			it.Count = n
			d.Fixed = append(d.Fixed, it)
		}
	}
	sort.Slice(d.Fixed, func(i, j int) bool {
		if d.Fixed[i].File != d.Fixed[j].File {
			return d.Fixed[i].File < d.Fixed[j].File
		}
		return d.Fixed[i].Rule < d.Fixed[j].Rule
	})
	return d
}

func writeSnapshot(path string, fs []finding, counts [numCats]int) error {
	idx := map[string]*snapItem{}
	var order []string
	for _, f := range fs {
		it, ok := idx[f.FP]
		if !ok {
			it = &snapItem{FP: f.FP, File: f.File, Rule: f.Name(), Message: f.Explain(), Cat: int(f.Category)}
			idx[f.FP] = it
			order = append(order, f.FP)
		}
		it.Count++
	}
	s := snapshot{Version: 1, Generated: time.Now().Format("02/01/2006 15:04"), Counts: counts}
	for _, k := range order {
		s.Items = append(s.Items, *idx[k])
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// ---------------------------------------------------------------
// Links para o Azure Repos
// ---------------------------------------------------------------

type linker struct {
	base string
	ref  string
}

func isCommitSHA(s string) bool {
	if len(s) != 40 {
		return false
	}
	for _, r := range s {
		if !strings.ContainsRune("0123456789abcdef", r) {
			return false
		}
	}
	return true
}

// normalizeRef aceita "GC<sha>", "GB<branch>", um SHA completo ou um nome de branch.
func normalizeRef(r string) string {
	switch {
	case r == "":
		return ""
	case strings.HasPrefix(r, "GC"), strings.HasPrefix(r, "GB"), strings.HasPrefix(r, "GT"):
		return r
	case isCommitSHA(strings.ToLower(r)):
		return "GC" + r
	default:
		return "GB" + r
	}
}

// detectAzureRepo lê as variáveis padrão do Azure Pipelines.
func detectAzureRepo() (base, ref string) {
	if os.Getenv("TF_BUILD") == "" {
		return "", ""
	}
	coll := strings.TrimRight(os.Getenv("SYSTEM_COLLECTIONURI"), "/")
	proj := os.Getenv("SYSTEM_TEAMPROJECT")
	repo := os.Getenv("BUILD_REPOSITORY_NAME")
	if coll == "" || proj == "" || repo == "" {
		return "", ""
	}
	base = coll + "/" + url.PathEscape(proj) + "/_git/" + url.PathEscape(repo)
	sha := os.Getenv("SYSTEM_PULLREQUEST_SOURCECOMMITID")
	if sha == "" {
		sha = os.Getenv("BUILD_SOURCEVERSION")
	}
	if sha != "" {
		ref = "GC" + sha
	}
	return base, ref
}

func newLinker(repoURL, ref string) linker {
	envBase, envRef := detectAzureRepo()
	if repoURL == "" {
		repoURL = envBase
	}
	if ref == "" {
		ref = envRef
	}
	return linker{base: strings.TrimRight(repoURL, "/"), ref: normalizeRef(ref)}
}

func (l linker) link(file string, line int) string {
	if l.base == "" || file == "" {
		return ""
	}
	p := filepath.ToSlash(file)
	if filepath.IsAbs(file) {
		if wd, err := os.Getwd(); err == nil {
			if rel, err := filepath.Rel(wd, file); err == nil && !strings.HasPrefix(rel, "..") {
				p = filepath.ToSlash(rel)
			}
		}
	}
	p = strings.TrimPrefix(p, "./")
	q := url.Values{}
	q.Set("path", "/"+p)
	if l.ref != "" {
		q.Set("version", l.ref)
	}
	q.Set("line", strconv.Itoa(line))
	q.Set("lineEnd", strconv.Itoa(line+1))
	q.Set("lineStartColumn", "1")
	q.Set("lineEndColumn", "1")
	q.Set("lineStyle", "plain")
	q.Set("_a", "contents")
	return l.base + "?" + q.Encode()
}

// ---------------------------------------------------------------
// Agregações
// ---------------------------------------------------------------

type ruleRow struct {
	Name    string
	Explain string
	Cat     category
	Count   int
}

type fileRow struct {
	File  string
	Count int
}

func byRule(fs []finding) []ruleRow {
	idx := map[string]*ruleRow{}
	for _, f := range fs {
		key := f.Linter + "/" + f.Name()
		r, ok := idx[key]
		if !ok {
			r = &ruleRow{Name: f.Name(), Explain: f.Explain(), Cat: f.Category}
			idx[key] = r
		}
		r.Count++
	}
	rows := make([]ruleRow, 0, len(idx))
	for _, r := range idx {
		rows = append(rows, *r)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Count != rows[j].Count {
			return rows[i].Count > rows[j].Count
		}
		return rows[i].Name < rows[j].Name
	})
	return rows
}

func byFile(fs []finding) []fileRow {
	idx := map[string]int{}
	for _, f := range fs {
		idx[f.File]++
	}
	rows := make([]fileRow, 0, len(idx))
	for file, n := range idx {
		rows = append(rows, fileRow{file, n})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Count != rows[j].Count {
			return rows[i].Count > rows[j].Count
		}
		return rows[i].File < rows[j].File
	})
	return rows
}

type group struct {
	File  string
	Count int
	Sec   int
	Bug   int
	Smell int
	New   int
	Open  bool
	Items []finding
}

func buildGroups(fs []finding) []group {
	idx := map[string]*group{}
	var order []string
	for _, f := range fs {
		g, ok := idx[f.File]
		if !ok {
			g = &group{File: f.File}
			idx[f.File] = g
			order = append(order, f.File)
		}
		g.Items = append(g.Items, f)
		g.Count++
		switch f.Category {
		case catSecurity:
			g.Sec++
		case catBug:
			g.Bug++
		default:
			g.Smell++
		}
		if f.New {
			g.New++
		}
	}
	out := make([]group, 0, len(order))
	for _, k := range order {
		g := *idx[k]
		sort.SliceStable(g.Items, func(i, j int) bool { return g.Items[i].Line < g.Items[j].Line })
		g.Open = g.Sec+g.Bug+g.New > 0 || len(order) <= 4
		out = append(out, g)
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		switch {
		case a.Sec != b.Sec:
			return a.Sec > b.Sec
		case a.Bug != b.Bug:
			return a.Bug > b.Bug
		case a.Count != b.Count:
			return a.Count > b.Count
		default:
			return a.File < b.File
		}
	})
	return out
}

// ---------------------------------------------------------------
// Visão para HTML e Markdown
// ---------------------------------------------------------------

type catView struct {
	Name       string
	Count      int
	Grade      string
	Limit      string
	Delta      string
	DeltaClass string
}

type gateRow struct {
	Name      string
	Actual    int
	LimitText string
	Status    string
	Class     string
	Grade     string
}

type chip struct {
	ID    int
	Name  string
	Count int
	Class string
}

type fixedRow struct {
	Cat     string
	File    string
	Rule    string
	Message string
	Count   int
}

type view struct {
	Generated   string
	Passed      bool
	Total       int
	Cats        []catView
	Gate        []gateRow
	Chips       []chip
	Rules       []ruleRow
	Files       []fileRow
	Groups      []group
	Findings    []finding
	HasBaseline bool
	BaselineSrc string
	NewCount    int
	FixedCount  int
	Fixed       []fixedRow
}

func delta(cur, prev int) (text, class string) {
	switch {
	case cur < prev:
		return fmt.Sprintf("▼ %d", prev-cur), "good"
	case cur > prev:
		return fmt.Sprintf("▲ %d", cur-prev), "worse"
	default:
		return "= igual", "same"
	}
}

func buildView(fs []finding, checks []check, passed bool, rules []ruleRow, files []fileRow, d diff) view {
	v := view{
		Generated:   time.Now().Format("02/01/2006 15:04"),
		Passed:      passed,
		Total:       len(fs),
		Rules:       rules,
		Files:       files,
		Groups:      buildGroups(fs),
		Findings:    fs,
		HasBaseline: d.Has,
		BaselineSrc: d.Source,
		NewCount:    d.New,
		FixedCount:  d.FixedCount,
	}
	for i, c := range checks {
		st, cl := statusOf(c)
		grd := "-"
		if i < numCats {
			grd = grade(c.Actual)
			cv := catView{Name: c.Name, Count: c.Actual, Grade: grd, Limit: "limite: " + limitText(c.Limit)}
			if d.Has {
				cv.Delta, cv.DeltaClass = delta(c.Actual, d.PrevCounts[i])
			}
			v.Cats = append(v.Cats, cv)
			v.Chips = append(v.Chips, chip{ID: i, Name: c.Name, Count: c.Actual, Class: catClassOf(category(i))})
		}
		v.Gate = append(v.Gate, gateRow{c.Name, c.Actual, limitText(c.Limit), st, cl, grd})
	}
	for _, it := range d.Fixed {
		cat := catSmell.String()
		if it.Cat >= 0 && it.Cat < numCats {
			cat = category(it.Cat).String()
		}
		v.Fixed = append(v.Fixed, fixedRow{cat, it.File, it.Rule, it.Message, it.Count})
	}
	return v
}

// ---------------------------------------------------------------
// Saídas
// ---------------------------------------------------------------

func cut(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

func printConsole(fs []finding, checks []check, passed bool, rules []ruleRow, files []fileRow, outDir string, d diff) {
	line := strings.Repeat("=", 70)
	fmt.Println(line)
	fmt.Printf(" RELATÓRIO DE QUALIDADE — %s\n", verdict(passed))
	fmt.Println(line)
	fmt.Printf(" %-20s %6s  %4s  %-10s  %s\n", "CATEGORIA", "ITENS", "NOTA", "LIMITE", "STATUS")
	for i, c := range checks {
		st, _ := statusOf(c)
		g := "-"
		if i < numCats {
			g = grade(c.Actual)
		}
		fmt.Printf(" %-20s %6d  %4s  %-10s  %s\n", c.Name, c.Actual, g, limitText(c.Limit), st)
	}
	if d.Has {
		fmt.Printf("\n Desde a execução de %s: %d novo(s), %d corrigido(s)\n", d.Source, d.New, d.FixedCount)
	}
	if len(fs) == 0 {
		fmt.Println("\n Nenhum achado. Código limpo para as regras ativas.")
		return
	}

	fmt.Println("\n Regras mais frequentes")
	for i, r := range rules {
		if i >= 8 {
			break
		}
		fmt.Printf("  %4d  %-18s %s\n", r.Count, r.Name, cut(r.Explain, 70))
	}

	fmt.Println("\n Arquivos com mais itens")
	for i, f := range files {
		if i >= 5 {
			break
		}
		fmt.Printf("  %4d  %s\n", f.Count, f.File)
	}

	fmt.Println("\n Detalhes")
	for i, f := range fs {
		if i >= 30 {
			fmt.Printf("  ... e mais %d (veja %s)\n", len(fs)-30, filepath.Join(outDir, "quality.html"))
			break
		}
		tag := ""
		if f.New {
			tag = " NOVO"
		}
		fmt.Printf("  [%s]%s %s:%d  %s — %s\n", f.Category, tag, f.File, f.Line, f.Name(), cut(f.Explain(), 80))
	}
	fmt.Printf("\n Relatório completo: %s\n", filepath.Join(outDir, "quality.html"))
}

func writeMarkdown(path string, v view) error {
	var b strings.Builder
	state := "✅ APROVADO"
	if !v.Passed {
		state = "❌ REPROVADO"
	}
	fmt.Fprintf(&b, "# Relatório de qualidade — %s\n\n", state)
	fmt.Fprintf(&b, "Gerado em %s · %d item(ns)\n\n", v.Generated, v.Total)
	if v.HasBaseline {
		fmt.Fprintf(&b, "Comparado com a execução de %s: **%d novo(s)** · %d corrigido(s)\n\n", v.BaselineSrc, v.NewCount, v.FixedCount)
	}
	b.WriteString("| Categoria | Itens | Nota | Limite | Status |\n|---|---|---|---|---|\n")
	for _, g := range v.Gate {
		fmt.Fprintf(&b, "| %s | %d | %s | %s | %s |\n", g.Name, g.Actual, g.Grade, g.LimitText, g.Status)
	}
	if len(v.Rules) > 0 {
		b.WriteString("\n## Regras mais frequentes\n\n| Qtd | Regra | O que significa |\n|---|---|---|\n")
		for i, r := range v.Rules {
			if i >= 10 {
				break
			}
			fmt.Fprintf(&b, "| %d | `%s` | %s |\n", r.Count, r.Name, strings.ReplaceAll(r.Explain, "|", "/"))
		}
		b.WriteString("\n## Detalhes\n\n")
		for i, f := range v.Findings {
			if i >= 100 {
				fmt.Fprintf(&b, "\n… e mais %d no relatório HTML.\n", len(v.Findings)-100)
				break
			}
			mark := ""
			if f.New {
				mark = "🆕 "
			}
			fmt.Fprintf(&b, "- %s**%s** `%s:%d` `%s` — %s\n", mark, f.Category, f.File, f.Line, f.Name(), f.Explain())
		}
	}
	return os.WriteFile(path, []byte(b.String()), 0o600)
}

const htmlTpl = `<!doctype html>
<html lang="pt-BR"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Relatório de qualidade</title>
<style>
:root{--bg:#f4f6f9;--fg:#1b2433;--card:#fff;--mut:#667085;--line:#e3e7ee;--ok:#16794a;--bad:#c62828;--warn:#a85d00;--sec:#c62828;--bug:#d9480f;--smell:#2563c9;--new:#7c3aed;--chip:#eef1f6}
@media (prefers-color-scheme:dark){:root{--bg:#0f141c;--fg:#e7eaf0;--card:#172030;--mut:#98a2b3;--line:#29344a;--ok:#3ecf8e;--bad:#ff6b6b;--warn:#f0b04a;--sec:#ff6b6b;--bug:#ff9a5c;--smell:#6ea8fe;--new:#b197fc;--chip:#1f2a3d}}
[hidden]{display:none!important}
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--fg);font:15px/1.5 system-ui,-apple-system,sans-serif}
main{max-width:1040px;margin:0 auto;padding:24px 16px 48px}
h1{font-size:22px;margin:0 0 2px}
h2{font-size:16px;margin:30px 0 8px}
.sub{color:var(--mut);font-size:13px}
.verdict{display:inline-block;padding:4px 14px;border-radius:999px;color:#fff;font-weight:700;margin:12px 0;letter-spacing:.02em}
.verdict.ok{background:var(--ok)}
.verdict.bad{background:var(--bad)}
.banner{background:var(--card);border:1px solid var(--line);border-left:4px solid var(--new);border-radius:8px;padding:10px 14px;margin:4px 0 14px;font-size:14px}
.banner .n{color:var(--new);font-weight:700}
.banner .f{color:var(--ok);font-weight:700}
.cards{display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:12px}
.card{background:var(--card);border:1px solid var(--line);border-radius:10px;padding:14px}
.card b{font-size:30px;display:block;line-height:1.2}
.grade{float:right;font-weight:800;font-size:22px}
.delta{font-size:12px;font-weight:700;margin-left:6px}
.delta.good{color:var(--ok)}
.delta.worse{color:var(--bad)}
.delta.same{color:var(--mut)}
.wrap{overflow-x:auto}
table{width:100%;border-collapse:collapse;background:var(--card)}
.wrap{border:1px solid var(--line);border-radius:10px}
th,td{text-align:left;padding:8px 10px;border-bottom:1px solid var(--line);vertical-align:top;font-size:14px}
tr:last-child td{border-bottom:0}
th{color:var(--mut);font-weight:600;font-size:13px}
code{font-family:ui-monospace,Menlo,Consolas,monospace;font-size:13px}
pre{margin:6px 0 0;padding:6px 8px;background:var(--chip);border-radius:6px;font:12.5px/1.4 ui-monospace,Menlo,Consolas,monospace;overflow-x:auto;white-space:pre}
a{color:var(--smell)}
.st-ok{color:var(--ok)}
.st-bad{color:var(--bad);font-weight:700}
.st-warn{color:var(--warn)}
.b{display:inline-block;font-size:12px;padding:1px 8px;border-radius:999px;border:1px solid var(--line);white-space:nowrap}
.b.sec{color:var(--sec);border-color:var(--sec)}
.b.bug{color:var(--bug);border-color:var(--bug)}
.b.smell{color:var(--smell);border-color:var(--smell)}
.b.new{background:var(--new);border-color:var(--new);color:#fff;font-weight:700}
.bar{position:sticky;top:0;z-index:5;background:var(--bg);padding:10px 0 8px;border-bottom:1px solid var(--line);margin-top:26px}
.row{display:flex;flex-wrap:wrap;gap:8px;align-items:center}
.row2{display:flex;flex-wrap:wrap;gap:8px;align-items:center;justify-content:space-between;margin-top:8px}
.chip{border:1px solid var(--line);background:var(--chip);color:var(--fg);padding:5px 12px;border-radius:999px;cursor:pointer;font:inherit;font-size:13px}
.chip::before{content:"";display:inline-block;width:8px;height:8px;border-radius:50%;margin-right:6px;background:var(--mut)}
.chip.sec::before{background:var(--sec)}
.chip.bug::before{background:var(--bug)}
.chip.smell::before{background:var(--smell)}
.chip.off{opacity:.45;text-decoration:line-through}
select,input[type=search]{font:inherit;font-size:14px;padding:6px 10px;border:1px solid var(--line);border-radius:8px;background:var(--card);color:var(--fg)}
input[type=search]{flex:1;min-width:200px}
.chk{font-size:14px;display:flex;gap:6px;align-items:center}
.btn{font:inherit;font-size:13px;padding:4px 10px;border:1px solid var(--line);border-radius:8px;background:var(--card);color:var(--fg);cursor:pointer}
details.grp,details.fixed{margin-top:10px;background:var(--card);border:1px solid var(--line);border-radius:10px;overflow:hidden}
details>summary{cursor:pointer;padding:10px 12px;display:flex;gap:10px;align-items:center;list-style:none}
details>summary::-webkit-details-marker{display:none}
details>summary::before{content:"▸";color:var(--mut)}
details[open]>summary::before{content:"▾"}
details[open]>summary{border-bottom:1px solid var(--line)}
.gfile{flex:1;word-break:break-all}
.badges{display:flex;gap:6px;flex-wrap:wrap}
.gcount{font-weight:700;background:var(--chip);border-radius:999px;padding:0 10px;min-width:28px;text-align:center}
details .wrap{border:0;border-radius:0}
td.loc{white-space:nowrap}
tr.rrow{cursor:pointer}
tr.rrow:hover td{background:var(--chip)}
@media (max-width:600px){main{padding:16px 12px 40px}th:nth-child(2),td:nth-child(2){display:none}}
</style></head><body><main>
<h1>Relatório de qualidade</h1>
<div class="sub">Gerado em {{.Generated}} · {{.Total}} item(ns) encontrado(s)</div>
<div class="verdict {{if .Passed}}ok{{else}}bad{{end}}">{{if .Passed}}APROVADO{{else}}REPROVADO{{end}}</div>
{{if .HasBaseline}}<div class="banner">Comparado com a execução de <b>{{.BaselineSrc}}</b>: <span class="n">{{.NewCount}} novo(s)</span> · <span class="f">{{.FixedCount}} corrigido(s)</span></div>{{end}}

<div class="cards">
{{range .Cats}}<div class="card"><span class="grade">{{.Grade}}</span><div class="sub">{{.Name}}</div><b>{{.Count}}</b><span class="sub">{{.Limit}}</span>{{if .Delta}}<span class="delta {{.DeltaClass}}">{{.Delta}}</span>{{end}}</div>
{{end}}</div>

<h2>Quality gate</h2>
<div class="wrap"><table>
<tr><th>Condição</th><th>Atual</th><th>Limite</th><th>Status</th></tr>
{{range .Gate}}<tr><td>{{.Name}}</td><td>{{.Actual}}</td><td>{{.LimitText}}</td><td class="{{.Class}}">{{.Status}}</td></tr>
{{end}}</table></div>

{{if .Groups}}
<div class="bar">
<div class="row">
{{range .Chips}}<button type="button" class="chip {{.Class}}" data-cat="{{.ID}}">{{.Name}} <b>{{.Count}}</b></button>
{{end}}<select id="rule" aria-label="Filtrar por regra"><option value="">Todas as regras</option>{{range .Rules}}<option value="{{.Name}}">{{.Name}} ({{.Count}})</option>{{end}}</select>
{{if .HasBaseline}}<label class="chk"><input type="checkbox" id="onlynew"> só novos</label>{{end}}
<input id="q" type="search" placeholder="Buscar arquivo, regra ou texto…" aria-label="Buscar">
</div>
<div class="row2"><span class="sub">Mostrando <b id="shown">{{.Total}}</b> de {{.Total}}</span>
<span><button type="button" class="btn" id="expand">Expandir tudo</button> <button type="button" class="btn" id="collapse">Recolher tudo</button></span></div>
</div>

{{range .Groups}}<details class="grp"{{if .Open}} open{{end}}>
<summary><code class="gfile">{{.File}}</code><span class="badges">{{if .Sec}}<span class="b sec">{{.Sec}} segurança</span>{{end}}{{if .Bug}}<span class="b bug">{{.Bug}} bugs</span>{{end}}{{if .Smell}}<span class="b smell">{{.Smell}} manutenção</span>{{end}}{{if .New}}<span class="b new">{{.New}} novo(s)</span>{{end}}</span><span class="gcount">{{.Count}}</span></summary>
<div class="wrap"><table>
{{range .Items}}<tr class="item" data-cat="{{.CatID}}" data-rule="{{.Name}}" data-new="{{if .New}}1{{else}}0{{end}}" data-text="{{.SearchText}}">
<td class="loc">{{if .URL}}<a href="{{.URL}}" target="_blank" rel="noopener noreferrer"><code>linha {{.Line}}</code></a>{{else}}<code>linha {{.Line}}</code>{{end}}</td>
<td><span class="b {{.CatClass}}">{{.Category.String}}</span></td>
<td><code>{{.Name}}</code></td>
<td>{{.Explain}}{{if .New}} <span class="b new">novo</span>{{end}}{{if .Snippet}}<pre>{{.Snippet}}</pre>{{end}}</td></tr>
{{end}}</table></div></details>
{{end}}

{{if .Fixed}}<details class="fixed"><summary><span class="gfile">Corrigidos desde a última execução</span><span class="gcount">{{.FixedCount}}</span></summary>
<div class="wrap"><table>
{{range .Fixed}}<tr><td>{{.Cat}}</td><td><code>{{.File}}</code></td><td><code>{{.Rule}}</code></td><td>{{.Message}}{{if gt .Count 1}} (×{{.Count}}){{end}}</td></tr>
{{end}}</table></div></details>{{end}}

<h2>Regras mais frequentes <span class="sub">(clique para filtrar)</span></h2>
<div class="wrap"><table>
<tr><th>Qtd</th><th>Categoria</th><th>Regra</th><th>O que significa</th></tr>
{{range .Rules}}<tr class="rrow" data-rule="{{.Name}}"><td>{{.Count}}</td><td>{{.Cat.String}}</td><td><code>{{.Name}}</code></td><td>{{.Explain}}</td></tr>
{{end}}</table></div>
{{end}}
</main>
<script>
(function () {
  var catOn = {'0': true, '1': true, '2': true};
  var rows = [].slice.call(document.querySelectorAll('tr.item'));
  var groups = [].slice.call(document.querySelectorAll('details.grp'));
  var q = document.getElementById('q');
  var rs = document.getElementById('rule');
  var onlyNew = document.getElementById('onlynew');
  var shown = document.getElementById('shown');
  if (!q || !rs || !shown) { return; }

  function apply() {
    var term = q.value.trim().toLowerCase();
    var rule = rs.value;
    var nOnly = onlyNew ? onlyNew.checked : false;
    var total = 0;
    rows.forEach(function (r) {
      var ok = catOn[r.dataset.cat] &&
        (rule === '' || r.dataset.rule === rule) &&
        (!nOnly || r.dataset.new === '1') &&
        (term === '' || r.dataset.text.indexOf(term) !== -1);
      r.hidden = !ok;
      if (ok) { total++; }
    });
    groups.forEach(function (g) {
      var n = g.querySelectorAll('tr.item:not([hidden])').length;
      g.hidden = (n === 0);
      var c = g.querySelector('.gcount');
      if (c) { c.textContent = n; }
    });
    shown.textContent = total;
  }

  [].slice.call(document.querySelectorAll('.chip')).forEach(function (c) {
    c.addEventListener('click', function () {
      var off = c.classList.toggle('off');
      catOn[c.dataset.cat] = !off;
      apply();
    });
  });
  [].slice.call(document.querySelectorAll('tr.rrow')).forEach(function (r) {
    r.addEventListener('click', function () {
      rs.value = r.dataset.rule;
      apply();
      rs.scrollIntoView({behavior: 'smooth', block: 'center'});
    });
  });
  q.addEventListener('input', apply);
  rs.addEventListener('change', apply);
  if (onlyNew) { onlyNew.addEventListener('change', apply); }
  document.getElementById('expand').addEventListener('click', function () {
    groups.forEach(function (g) { g.open = true; });
  });
  document.getElementById('collapse').addEventListener('click', function () {
    groups.forEach(function (g) { g.open = false; });
  });
})();
</script>
</body></html>
`

func writeHTML(path string, v view) error {
	t, err := template.New("report").Parse(htmlTpl)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, v); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o600)
}

func oneLine(s string) string {
	return strings.NewReplacer("\n", " ", "\r", " ", "]", ")").Replace(s)
}

// azureLog emite anotações que o Azure DevOps mostra no resumo da build.
func azureLog(fs []finding, summaryPath string) {
	if os.Getenv("TF_BUILD") == "" {
		return
	}
	const maxAnnotations = 50
	for i, f := range fs {
		if i >= maxAnnotations {
			break
		}
		kind := "warning"
		if f.Category != catSmell {
			kind = "error"
		}
		fmt.Printf("##vso[task.logissue type=%s;sourcepath=%s;linenumber=%d;columnnumber=%d;code=%s]%s\n",
			kind, f.File, f.Line, f.Col, f.Name(), oneLine(f.Explain()))
	}
	if abs, err := filepath.Abs(summaryPath); err == nil {
		fmt.Printf("##vso[task.uploadsummary]%s\n", abs)
	}
}

func fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "lintreport: "+format+"\n", a...)
	os.Exit(2)
}

func main() {
	in := flag.String("in", "reports/lint.json", "JSON gerado pelo golangci-lint (--output.json.path)")
	out := flag.String("out", "reports", "pasta de saída dos relatórios")
	maxSec := flag.Int("max-security", 0, "máximo de achados de segurança (-1 = sem limite)")
	maxBugs := flag.Int("max-bugs", 0, "máximo de bugs (-1 = sem limite)")
	maxSmells := flag.Int("max-smells", -1, "máximo de itens de manutenção (-1 = apenas avisa)")
	maxNewSmells := flag.Int("max-new-smells", -1, "máximo de itens de manutenção NOVOS em relação à base (-1 = sem limite)")
	baseline := flag.String("baseline", "", "snapshot.json de uma execução anterior (padrão: <out>/snapshot.json)")
	repoURL := flag.String("repo-url", "", "URL do repositório no Azure Repos para gerar links (automático no pipeline)")
	ref := flag.String("ref", "", "commit (SHA) ou branch para os links (automático no pipeline)")
	flag.Parse()

	data, err := os.ReadFile(*in) //nolint:gosec // caminho vem de flag da CLI
	if err != nil {
		fatalf("não consegui ler %s: %v (o lint rodou e gerou o JSON?)", *in, err)
	}
	var rep report
	if err := json.Unmarshal(data, &rep); err != nil {
		fatalf("JSON inválido em %s: %v", *in, err)
	}

	findings := classify(rep.Issues)
	d := compare(findings, loadBaseline(*baseline, *out))

	lk := newLinker(*repoURL, *ref)
	for i := range findings {
		findings[i].URL = lk.link(findings[i].File, findings[i].Line)
	}

	var counts [numCats]int
	for _, f := range findings {
		counts[f.Category]++
	}
	checks, passed := evaluate(counts, *maxSec, *maxBugs, *maxSmells)
	if d.Has && *maxNewSmells >= 0 {
		c := check{"Manutenção (novos)", d.NewByCat[catSmell], *maxNewSmells, d.NewByCat[catSmell] <= *maxNewSmells}
		checks = append(checks, c)
		if !c.OK {
			passed = false
		}
	}
	rules, files := byRule(findings), byFile(findings)

	printConsole(findings, checks, passed, rules, files, *out, d)

	if err := os.MkdirAll(*out, 0o750); err != nil {
		fatalf("não consegui criar %s: %v", *out, err)
	}
	v := buildView(findings, checks, passed, rules, files, d)
	mdPath := filepath.Join(*out, "summary.md")
	if err := writeMarkdown(mdPath, v); err != nil {
		fatalf("erro ao gravar Markdown: %v", err)
	}
	if err := writeHTML(filepath.Join(*out, "quality.html"), v); err != nil {
		fatalf("erro ao gravar HTML: %v", err)
	}
	if err := writeSnapshot(filepath.Join(*out, "snapshot.json"), findings, counts); err != nil {
		fatalf("erro ao gravar snapshot: %v", err)
	}
	azureLog(findings, mdPath)

	if !passed {
		os.Exit(1)
	}
}
