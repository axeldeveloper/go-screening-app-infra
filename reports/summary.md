# Relatório de qualidade — ✅ APROVADO

Gerado em 19/11/2026 21:19 · 29 item(ns)

Comparado com a execução de 19/11/2026 21:17: **3 novo(s)** · 3 corrigido(s)

| Categoria | Itens | Nota | Limite | Status |
|---|---|---|---|---|
| Segurança | 0 | A | 0 | ok |
| Bugs | 0 | A | 0 | ok |
| Manutenção | 29 | E | sem limite | aviso |

## Regras mais frequentes

| Qtd | Regra | O que significa |
|---|---|---|
| 14 | `exported` | Item público sem comentário de documentação. |
| 7 | `unused-parameter` | Parâmetro nunca usado; renomeie para _ ou remova. |
| 3 | `gofmt` | File is not properly formatted |
| 3 | `package-comments` | Pacote sem comentário de documentação no formato esperado. |
| 1 | `blank-imports` | a blank import should be only in a main or test package, or have a comment justifying it |
| 1 | `exitAfterDefer` | log.Fatalf will exit, and `defer cancel()` will not run |

## Detalhes

- **Manutenção** `cmd/api/main.go:74` `exitAfterDefer` — log.Fatalf will exit, and `defer cancel()` will not run
- **Manutenção** `internal/config/config.go:1` `package-comments` — Pacote sem comentário de documentação no formato esperado.
- **Manutenção** `internal/config/config.go:5` `exported` — Item público sem comentário de documentação.
- **Manutenção** `internal/config/config.go:10` `exported` — Item público sem comentário de documentação.
- **Manutenção** `internal/config/config_test.go:10` `gofmt` — File is not properly formatted
- 🆕 **Manutenção** `internal/db/db.go:1` `package-comments` — Pacote sem comentário de documentação no formato esperado.
- **Manutenção** `internal/db/db.go:10` `blank-imports` — a blank import should be only in a main or test package, or have a comment justifying it
- **Manutenção** `internal/db/db.go:24` `exported` — Item público sem comentário de documentação.
- **Manutenção** `internal/db/db.go:41` `exported` — Item público sem comentário de documentação.
- **Manutenção** `internal/db/db_test.go:10` `gofmt` — File is not properly formatted
- **Manutenção** `internal/handler/handler.go:14` `exported` — Item público sem comentário de documentação.
- **Manutenção** `internal/handler/handler.go:18` `exported` — Item público sem comentário de documentação.
- **Manutenção** `internal/handler/handler.go:22` `exported` — Item público sem comentário de documentação.
- **Manutenção** `internal/handler/handler.go:31` `exported` — Item público sem comentário de documentação.
- **Manutenção** `internal/repository/item.go:1` `package-comments` — Pacote sem comentário de documentação no formato esperado.
- **Manutenção** `internal/repository/item.go:12` `exported` — Item público sem comentário de documentação.
- **Manutenção** `internal/repository/item.go:21` `exported` — Item público sem comentário de documentação.
- 🆕 **Manutenção** `internal/service/item_service_test.go:45` `gofmt` — File is not properly formatted
- **Manutenção** `internal/service/item_service_test.go:50` `unused-parameter` — Parâmetro nunca usado; renomeie para _ ou remova.
- **Manutenção** `internal/service/item_service_test.go:58` `unused-parameter` — Parâmetro nunca usado; renomeie para _ ou remova.
- **Manutenção** `internal/service/item_service_test.go:77` `unused-parameter` — Parâmetro nunca usado; renomeie para _ ou remova.
- **Manutenção** `internal/service/item_service_test.go:86` `unused-parameter` — Parâmetro nunca usado; renomeie para _ ou remova.
- **Manutenção** `internal/service/item_service_test.go:92` `unused-parameter` — Parâmetro nunca usado; renomeie para _ ou remova.
- **Manutenção** `internal/web/web.go:17` `exported` — Item público sem comentário de documentação.
- **Manutenção** `internal/web/web.go:23` `exported` — Item público sem comentário de documentação.
- **Manutenção** `internal/web/web.go:35` `exported` — Item público sem comentário de documentação.
- **Manutenção** `internal/web/web.go:43` `exported` — Item público sem comentário de documentação.
- **Manutenção** `internal/web/web.go:55` `unused-parameter` — Parâmetro nunca usado; renomeie para _ ou remova.
- 🆕 **Manutenção** `internal/web/web_test.go:24` `unused-parameter` — Parâmetro nunca usado; renomeie para _ ou remova.
