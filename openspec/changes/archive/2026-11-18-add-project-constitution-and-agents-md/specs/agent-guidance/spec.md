# Spec Delta

## Purpose

Fornece instruções acionáveis para que agentes de codificação e contribuidores trabalhem no repositório de forma consistente, cobrindo a estrutura de diretórios, os comandos de verificação, as convenções de código e o fluxo de trabalho de mudanças.

## ADDED Requirements

### Requirement: Guia de agentes na raiz
O projeto MUST manter um arquivo `AGENTS.md` na raiz do repositório com instruções aplicáveis a agentes de IA que trabalham no código.

#### Scenario: Início de uma tarefa
- **WHEN** um agente começa a trabalhar em uma tarefa neste repositório
- **THEN** encontra `AGENTS.md` na raiz com instruções que se aplicam ao trabalho

### Requirement: Comandos essenciais documentados
O `AGENTS.md` MUST documentar os comandos de build, teste, formatação, análise estática, geração de código e migrações disponíveis no projeto.

#### Scenario: Verificação de uma mudança
- **WHEN** um agente precisa validar uma alteração que fez
- **THEN** encontra os comandos de teste, vet, formatação e geração (`make test`, `make vet`, `make fmt`, `make generate`)

#### Scenario: Preparação do banco local
- **WHEN** um agente precisa aplicar migrações localmente
- **THEN** encontra os comandos `make migrate-up` e `make migrate-down` e a variável `MIGRATE_URL`

### Requirement: Estrutura e fluxo de dados descritos
O `AGENTS.md` MUST descrever o layout de diretórios e a direção das dependências entre camadas, incluindo o ponto de entrada `cmd/api` e o fluxo `config → db → repository → service → handler/web`.

#### Scenario: Adição de um caso de uso
- **WHEN** um agente precisa implementar uma nova funcionalidade
- **THEN** o guia indica a ordem das camadas e os pacotes envolvidos, do entrypoint ao handler

### Requirement: Convenções de código e qualidade
O `AGENTS.md` MUST instruir agentes a seguir as convenções Go do projeto: formatação com `gofmt`, análise com `go vet` e `golangci-lint`, erros contextualizados com `%w`, propagação de `context.Context` em chamadas de banco e serviços, e identificadores em inglês no código.

#### Scenario: Escrita de novo código Go
- **WHEN** um agente implementa um novo arquivo ou pacote Go
- **THEN** segue as convenções de formatação, tratamento de erros e uso de contexto documentadas no guia

### Requirement: Fluxo de trabalho OpenSpec documentado
O `AGENTS.md` MUST descrever o ciclo de trabalho OpenSpec de propor, aplicar e arquivar, e exigir que mudanças de comportamento sejam planejadas como uma change antes da implementação.

#### Scenario: Mudança de comportamento solicitada
- **WHEN** um agente recebe uma solicitação que altera o comportamento observável do sistema
- **THEN** verifica ou cria uma change OpenSpec antes de editar o código de aplicação

### Requirement: Idioma do guia
O `AGENTS.md` MUST ser escrito em português, de modo que agentes e colaboradores possam segui-lo sem tradução.

#### Scenario: Leitura do guia
- **WHEN** um agente lê `AGENTS.md`
- **THEN** todo o conteúdo está redigido em português
