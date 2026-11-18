# Design

## Context

O projeto é um serviço Go scaffoldado com TriStack (Gin, sqlc, Postgres, golang-migrate) cuja arquitetura em camadas já está implementada em `cmd/api` e `internal/` (config → db → repository → service → handler/web). Não existem specs nem documentos de governança. A motivação está em `proposal.md`; os requisitos observáveis estão em `specs/project-constitution/spec.md` e `specs/agent-guidance/spec.md`.

Restrição relevante: o fluxo de planejamento do OpenSpec carrega contexto a partir de `openspec/config.yaml` (campo `context`), e não lê automaticamente um arquivo `constitution.md`. Qualquer integração precisa passar por esse campo para influenciar as propostas.

## Goals / Non-Goals

**Goals:**

- Estabelecer `CONSTITUTION.md` na raiz como fonte de verdade dos princípios do projeto.
- Tornar esses princípios efetivamente visíveis ao planejamento OpenSpec, sem exigir que o usuário os releia manualmente.
- Produzir um `AGENTS.md` com instruções verificáveis contra o estado real do repositório (comandos do Makefile, layout de pacotes e fluxo OpenSpec).
- Manter todo o conteúdo em português.

**Non-Goals:**

- Alterar código de aplicação, endpoints, dependências ou configuração de runtime.
- Traduzir o `README.md` existente.
- Corrigir divergências pré-existentes entre documentação e código (isso será registrado, não resolvido aqui).
- Introduzir novas ferramentas ou scripts de validação automática de documentação.

## Decisions

### 1. Fonte de verdade única na raiz, com espelhos mínimos no OpenSpec

`CONSTITUTION.md` na raiz contém os princípios completos. `openspec/constitution.md` é um arquivo curto que declara a raiz como autoridade, lista os títulos dos princípios e aponta para o arquivo completo. O campo `context` de `openspec/config.yaml` recebe um resumo dos princípios em português, com a instrução de consultar `CONSTITUTION.md` para os detalhes.

- **Por quê:** atende ao pedido de ter o documento na raiz *e* integrado ao OpenSpec, sem criar três documentos concorrentes com conteúdo integral idêntico.
- **Alternativas consideradas:** (a) apenas `openspec/constitution.md` — descartada porque o usuário pediu o arquivo também na raiz; (b) `CONSTITUTION.md` na raiz sem integração — descartada porque os princípios não seriam carregados no planejamento; (c) symlink de `openspec/constitution.md` para a raiz — descartada por fragilidade em diferentes sistemas operacionais e em ambientes que não preservam links simbólicos.

### 2. Conteúdo derivado do código, não de suposições

Os princípios e instruções serão escritos a partir do que existe de fato: `Makefile` (`test`, `vet`, `fmt`, `generate`, `migrate-up`, `migrate-down`), `cmd/api/main.go` (entrypoint, rotas, timeouts do servidor, graceful shutdown), `internal/config`, `internal/db` (pgx, pool de conexões), `internal/repository` (sqlc), `internal/service` (regras e `ErrInvalidInput`) e `internal/handler` (adaptação HTTP com `MaxBytesReader`).

- **Por quê:** evita documentar práticas inexistentes e torna o `AGENTS.md` imediatamente utilizável.
- **Alternativas consideradas:** copiar o README — descartada porque o README descreve o scaffold e contém valores que divergem do código; escrever diretrizes genéricas de Go — descartada por não serem specificamente acionáveis neste repositório.

### 3. Conteúdo em português, identificadores de código em inglês

Toda a prosa dos três arquivos será em português, conforme decidido pelo solicitante. Nomes de pacotes, funções, variáveis, comandos e caminhos permanecem em inglês, mantendo a compatibilidade com o código Go existente.

- **Por quê:** satisfaz a escolha de idioma sem introduzir problemas de convenção ou de tooling no código.
- **Alternativas consideradas:** traduzir identificadores — descartada por quebrar a compatibilidade com Go e com o código já escrito.

## Risks / Trade-offs

- [Duplicação de conteúdo entre raiz, `openspec/constitution.md` e `config.yaml`] → A raiz é a única autoridade; os outros dois são resumos. Toda alteração começa na raiz e é propagada, com uma etapa de verificação nas tarefas.
- [O `context` do `config.yaml` tem limite de tamanho e deve ser um texto simples] → Manter o resumo curto (muito abaixo do limite de 51.200 bytes) e concentrar os detalhes na raiz.
- [Divergências pré-existentes entre README e código: porta padrão `3000` no README contra `8000` em `config.go`; DSN e credenciais diferentes entre README/Makefile/código] → Não corrigir nesta change; registrar como pendência e propor uma change separada de documentação/configuração.
- [Nem toda ferramenta de IA lê `AGENTS.md` automaticamente] → Trata-se de uma convenção amplamente adotada; o guia é escrito para ser útil a humanos e agentes que a suportam, sem depender de enforcement.
- [Restrições de aprendizado: os princípios podem ficar desatualizados se o código evoluir] → Revisar os documentos como parte das mudanças que alterem a arquitetura; isso fica registrado no próprio guia.

## Migration Plan

Esta é uma mudança apenas de documentação, sem impacto em runtime.

1. Criar `CONSTITUTION.md` e `AGENTS.md` na raiz.
2. Criar `openspec/constitution.md` apontando para a raiz.
3. Atualizar `openspec/config.yaml` com o resumo dos princípios no campo `context`, preservando as configurações existentes (`schema`, `githubCopilot`).
4. Verificar que os três locais não se contradizem e que os comandos documentados batem com o `Makefile`.
5. Rollback: remover os arquivos criados e restaurar `openspec/config.yaml`. Nenhuma mudança de dados ou estado de aplicação a desfazer.

## Open Questions

- As divergências já observadas entre README e código (porta padrão e DSN do banco) devem ser corrigidas em uma change separada de configuração/documentação? Esta change apenas as registra.
- O `README.md` deve ser traduzido para português em algum momento para uniformizar o idioma da documentação? Fora do escopo desta change.
