# Constituição (integração OpenSpec)

Este arquivo é um espelho mínimo da constituição do projeto. A única fonte de verdade é [`CONSTITUTION.md`](../CONSTITUTION.md), na raiz do repositório. Toda alteração de princípio deve ser feita primeiro lá e depois propagada para este arquivo e para o campo `context` de `config.yaml`.

## Princípios

1. Arquitetura em camadas — `config → db → repository → service → handler/web`, com `cmd/api` como raiz de composição.
2. Fronteiras de responsabilidade — cada pacote em `internal/` tem uma única responsabilidade; handlers adaptam HTTP, serviços detêm as regras, repositórios acessam dados.
3. Qualidade e testes — `gofmt`, `go vet`, `golangci-lint`, erros com `%w`, `context.Context`, testes table-driven e `make test` com race detector.
4. Segurança — limites e validação de entrada, respostas genéricas para erros internos, segredos por variáveis de ambiente, SQL parametrizado pelo sqlc, timeouts e desligamento graceful.
5. Convenções de documentação — prosa em português, identificadores e caminhos em inglês; documentação reflete o código e divergências são registradas.
6. Revisão da constituição — mudanças de arquitetura ou segurança exigem revisar e atualizar a constituição na mesma change.

Para os detalhes e a redação completa dos princípios, consulte [`CONSTITUTION.md`](../CONSTITUTION.md).
