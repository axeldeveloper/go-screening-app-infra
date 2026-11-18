# Tasks

## 1. Constituição na raiz

- [x] 1.1 Criar `CONSTITUTION.md` na raiz, em português, com os princípios completos derivados do código existente: arquitetura em camadas (`config → db → repository → service → handler/web`), fronteiras de responsabilidade, qualidade e testes, segurança e convenções de documentação; verifique que o arquivo existe e que cada princípio pode ser conferido em `cmd/api/`, `internal/` ou `Makefile`.
- [x] 1.2 Incluir em `CONSTITUTION.md` a regra de que mudanças de arquitetura exigem revisão do próprio documento; verifique que a seção está presente e que todo o texto está em português.

## 2. Integração com OpenSpec

- [x] 2.1 Criar `openspec/constitution.md` declarando `CONSTITUTION.md` da raiz como única fonte de verdade, listando os títulos dos princípios e apontando para o arquivo completo; verifique que o arquivo existe e que não reproduz o conteúdo integral da raiz.
- [x] 2.2 Atualizar `openspec/config.yaml` com um resumo dos princípios no campo `context`, preservando `schema: spec-driven` e `githubCopilot.cloudAgent: true`; verifique que `openspec context --json` carrega o texto do resumo e que ele permanece bem abaixo de 51.200 bytes.

## 3. Guia para agentes (`AGENTS.md`)

- [x] 3.1 Criar `AGENTS.md` na raiz, em português, cobrindo: layout de diretórios, fluxo `config → db → repository → service → handler/web` com ponto de entrada `cmd/api`, comandos do `Makefile` (`make test`, `make vet`, `make fmt`, `make generate`, `make migrate-up`, `make migrate-down` e `MIGRATE_URL`), convenções Go (`gofmt`, `go vet`, `golangci-lint`, `%w`, `context.Context`, identificadores em inglês) e o ciclo OpenSpec propor → aplicar → arquivar; verifique que o arquivo existe e que cada comando documentado consta no `Makefile`.
- [x] 3.2 Registrar em `AGENTS.md` as divergências conhecidas entre README e código (porta padrão `3000` no README contra `8000` em `config.go`, e DSN/credenciais diferentes entre README, `Makefile` e código) como pendências a confirmar, sem alterar código; verifique que as pendências estão listadas e que nenhum arquivo `.go` foi modificado.

## 4. Verificação final de consistência

- [x] 4.1 Conferir que os três locais (`CONSTITUTION.md`, `openspec/constitution.md` e `openspec/config.yaml`) não se contradizem e que a raiz continua sendo a fonte de verdade; verifique comparando os títulos dos princípios nos três arquivos.
- [x] 4.2 Executar `git diff --stat` e confirmar que apenas os quatro arquivos de documentação foram criados/alterados, sem mudanças em código, rotas ou dependências; verifique que `go test ./...` continua passando.

## Workflow follow-up

- Revisar e aprovar os documentos antes de arquivar a change.
- Arquivar a change quando a revisão estiver concluída; para isso, iniciar o workflow de arquivamento correspondente.
- Se decidido, abrir uma change separada para corrigir as divergências de configuração/documentação registradas na tarefa 3.2.
