# Constituição do Projeto

Este documento define os princípios que norteiam as decisões de arquitetura e implementação de `my-screening-app`. É a fonte de verdade: `openspec/constitution.md` e o campo `context` de `openspec/config.yaml` são resumos derivados dele. Toda alteração de princípio começa aqui.

## 1. Arquitetura em camadas

O sistema é organizado em camadas, com as dependências apontando sempre para a camada imediatamente inferior:

`config → db → repository → service → handler/web`, tendo `cmd/api` como raiz de composição.

- `cmd/api/main.go` monta as dependências na ordem acima e registra as rotas; nenhuma regra de negócio vive no entrypoint.
- `internal/handler` adapta requisições e respostas HTTP e não contém regras de negócio.
- `internal/service` concentra as regras de negócio e validações (por exemplo, `validateName` e `ErrInvalidInput`).
- `internal/repository` é a única camada que acessa dados, por meio do código gerado pelo sqlc em `internal/db/sqlc`.
- `internal/model` define as estruturas de domínio compartilhadas.
- `internal/db` e `internal/config` são infraestrutura consumida na composição, não lugar para lógica de negócio.

Um pacote NÃO DEVE importar detalhes de infraestrutura de camada superior ou paralela. Serviços dependem de interfaces, como `repository.ItemRepository`, para permitir a substituição da implementação em testes.

## 2. Fronteiras de responsabilidade

- Cada pacote em `internal/` tem uma única responsabilidade e não pode ser importado de fora do módulo `my_screening_app`.
- Handlers traduzem erros de domínio em códigos HTTP; detalhes de implementação não vazam na resposta.
- Repositórios convertem entre os tipos do sqlc e `internal/model`, sem interpretar regras de negócio.
- Novos casos de uso seguem o fluxo já existente: modelo → query SQL → sqlc → repositório → serviço → handler/rota, com migração em `db/migrations` sempre que o esquema mudar.

## 3. Qualidade e testes

- Código Go é formatado com `gofmt` e verificado com `go vet` e `golangci-lint`. Erros são enriquecidos com `%w` e `context.Context` é propagado nas chamadas de banco e serviços.
- Testes ficam ao lado do código que exercitam, usam tabelas de casos com `t.Run` e comparações com `errors.Is`, como em `internal/service/item_service_test.go`.
- `make test` executa `go test -race ./...`; `make vet`, `make fmt` e `make generate` são os comandos de verificação e geração de código.
- Nenhuma mudança é considerada pronta sem passar pelos comandos acima.

## 4. Segurança

- Toda entrada HTTP é limitada e validada antes de chegar ao domínio: `http.MaxBytesReader` no handler (1 MiB) e validação do nome no serviço (não vazio, no máximo 200 caracteres).
- Erros internos são tratados e respondidos de forma genérica ao cliente; detalhes de banco ou pilha de chamadas não são expostos.
- Configurações sensíveis, como `DATABASE_URL`, vêm de variáveis de ambiente; nenhuma credencial é fixada no código-fonte. `env.example` documenta as variáveis sem segredos reais.
- Consultas são parametrizadas pelo sqlc; entrada do usuário nunca é concatenada em SQL.
- O servidor define timeouts de leitura/escrita e realiza desligamento graceful por sinal (SIGINT/SIGTERM).

## 5. Convenções de documentação

- Esta constituição e os guias do projeto são escritos em português; identificadores de código, comandos, caminhos e nomes de pacotes permanecem em inglês.
- Documentação descreve o estado real do código. Divergências encontradas são registradas como pendências e resolvidas em change separada, nunca silenciadas.
- O README existente permanece em inglês até que uma change específica decida sua tradução.

## 6. Revisão da constituição

Mudanças que alteram a arquitetura, as fronteiras entre camadas ou as garantias de segurança DEVEM revisar este documento e, se necessário, atualizá-lo na mesma change. `CONSTITUTION.md` é a única fonte de verdade; `openspec/constitution.md` e o campo `context` de `openspec/config.yaml` devem ser propagados a partir dele para evitar divergência.
