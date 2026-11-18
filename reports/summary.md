# Relatório de qualidade — ✅ APROVADO

Gerado em 18/11/2026 20:16 · 20 item(ns)

Comparado com a execução de 18/11/2026 20:16: **1 novo(s)** · 1 corrigido(s)

| Categoria | Itens | Nota | Limite | Status |
|---|---|---|---|---|
| Segurança | 0 | A | 0 | ok |
| Bugs | 0 | A | 0 | ok |
| Manutenção | 20 | D | sem limite | aviso |

## Regras mais frequentes

| Qtd | Regra | O que significa |
|---|---|---|
| 14 | `exported` | Item público sem comentário de documentação. |
| 3 | `package-comments` | Pacote sem comentário de documentação no formato esperado. |
| 1 | `blank-imports` | a blank import should be only in a main or test package, or have a comment justifying it |
| 1 | `exitAfterDefer` | log.Fatalf will exit, and `defer cancel()` will not run |
| 1 | `unused-parameter` | Parâmetro nunca usado; renomeie para _ ou remova. |

## Detalhes

- **Manutenção** `cmd/api/main.go:74` `exitAfterDefer` — log.Fatalf will exit, and `defer cancel()` will not run
- **Manutenção** `internal/config/config.go:5` `exported` — Item público sem comentário de documentação.
- **Manutenção** `internal/config/config.go:10` `exported` — Item público sem comentário de documentação.
- 🆕 **Manutenção** `internal/db/db.go:1` `package-comments` — Pacote sem comentário de documentação no formato esperado.
- **Manutenção** `internal/db/db.go:10` `blank-imports` — a blank import should be only in a main or test package, or have a comment justifying it
- **Manutenção** `internal/db/db.go:20` `exported` — Item público sem comentário de documentação.
- **Manutenção** `internal/db/db.go:37` `exported` — Item público sem comentário de documentação.
- **Manutenção** `internal/handler/handler.go:1` `package-comments` — Pacote sem comentário de documentação no formato esperado.
- **Manutenção** `internal/handler/handler.go:14` `exported` — Item público sem comentário de documentação.
- **Manutenção** `internal/handler/handler.go:18` `exported` — Item público sem comentário de documentação.
- **Manutenção** `internal/handler/handler.go:22` `exported` — Item público sem comentário de documentação.
- **Manutenção** `internal/handler/handler.go:31` `exported` — Item público sem comentário de documentação.
- **Manutenção** `internal/repository/item.go:1` `package-comments` — Pacote sem comentário de documentação no formato esperado.
- **Manutenção** `internal/repository/item.go:12` `exported` — Item público sem comentário de documentação.
- **Manutenção** `internal/repository/item.go:21` `exported` — Item público sem comentário de documentação.
- **Manutenção** `internal/web/web.go:17` `exported` — Item público sem comentário de documentação.
- **Manutenção** `internal/web/web.go:23` `exported` — Item público sem comentário de documentação.
- **Manutenção** `internal/web/web.go:35` `exported` — Item público sem comentário de documentação.
- **Manutenção** `internal/web/web.go:43` `exported` — Item público sem comentário de documentação.
- **Manutenção** `internal/web/web.go:55` `unused-parameter` — Parâmetro nunca usado; renomeie para _ ou remova.
