# Proposta

## Por quê

O projeto `my-screening-app` é um serviço Go (Gin + sqlc + Postgres + golang-migrate, scaffoldado com TriStack) com uma arquitetura em camadas bem definida, mas não possui documentação de governança nem instruções para agentes de IA que trabalham no repositório. Sem esses documentos, cada contribuidor ou agente precisa inferir as fronteiras entre camadas, as convenções de código e o fluxo de trabalho a partir do zero, o que gera decisões inconsistentes e retrabalho.

Esta mudança estabelece a base documental do projeto: uma constitution com os princípios que norteiam as decisões e uma `AGENTS.md` com instruções acionáveis para agentes de codificação.

## O que muda

- Adiciona `CONSTITUTION.md` na raiz com os princípios orientadores do projeto (arquitetura em camadas, fronteiras de responsabilidade, qualidade e testes, segurança, e convenções de documentação), derivados da estrutura real do código.
- Integra a constitution ao OpenSpec: cria `openspec/constitution.md` com os princípios e referencia o arquivo da raiz, e atualiza `openspec/config.yaml` para incluir esses princípios no contexto de planejamento, de modo que sejam considerados em toda proposta.
- Adiciona `AGENTS.md` na raiz com instruções práticas para agentes de IA: comandos de build/teste/gregação de código, layout de diretórios, fluxo de dados entre camadas, regras de estilo (`gofmt`, `go vet`, `golangci-lint`) e o fluxo de trabalho OpenSpec (propor → aplicar → arquivar).
- Registra explicitamente as divergências observadas entre documentação e código (por exemplo, portas padrão e DSN do banco) como pendências a serem confirmadas, sem alterar o comportamento da aplicação.

## Capacidades

### Novas capacidades

- `project-constitution`: os princípios de governança do projeto e sua disponibilização tanto na raiz do repositório quanto no contexto do OpenSpec.
- `agent-guidance`: as instruções em `AGENTS.md` que orientam agentes de codificação a trabalhar de forma consistente no repositório.

### Capacidades modificadas

- Nenhuma. O projeto ainda não possui specs existentes.

## Impacto

- Arquivos novos na raiz: `CONSTITUTION.md`, `AGENTS.md`.
- Arquivos novos/alterados em `openspec/`: `openspec/constitution.md`, `openspec/config.yaml`.
- Nenhuma mudança em código de aplicação, APIs, dependências ou comportamento em runtime.
- Idioma: todo o conteúdo será escrito em português, conforme decidido pelo solicitante, embora o README existente esteja em inglês.
