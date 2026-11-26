# Spec Delta

## Purpose

Gerenciar o ciclo de vida de usuários do sistema: cadastro, consulta, atualização e remoção, com e-mail único e validação de entrada.

## ADDED Requirements

### Requirement: Create user
O sistema SHALL criar um usuário quando recebido nome e e-mail válidos, exigindo que o e-mail não esteja cadastrado.

#### Scenario: Criar usuário com sucesso
- **WHEN** a requisição informar nome e e-mail não cadastrado
- **THEN** o sistema cria o usuário e retorna os dados do usuário criado

#### Scenario: E-mail duplicado
- **WHEN** a requisição informar um e-mail já cadastrado
- **THEN** o sistema recusa a criação e retorna erro de e-mail duplicado

#### Scenario: Entrada inválida
- **WHEN** a requisição informar nome vazio ou e-mail inválido
- **THEN** o sistema recusa a criação e retorna erro de validação

### Requirement: List users
O sistema SHALL listar todos os usuários cadastrados.

#### Scenario: Listar usuários
- **WHEN** o cliente solicita a listagem de usuários
- **THEN** o sistema retorna todos os usuários cadastrados

#### Scenario: Nenhum usuário cadastrado
- **WHEN** não há usuários cadastrados
- **THEN** o sistema retorna uma lista vazia

### Requirement: Get user
O sistema SHALL retornar um usuário pelo identificador, ou erro quando não existir.

#### Scenario: Usuário encontrado
- **WHEN** o cliente solicita um usuário por identificador existente
- **THEN** o sistema retorna os dados do usuário

#### Scenario: Usuário não encontrado
- **WHEN** o cliente solicita um usuário por identificador inexistente
- **THEN** o sistema retorna erro de usuário não encontrado

### Requirement: Update user
O sistema SHALL atualizar nome ou e-mail de um usuário existente, exigindo que o novo e-mail não esteja cadastrado por outro usuário.

#### Scenario: Atualizar usuário com sucesso
- **WHEN** a requisição informar identificador existente e dados válidos
- **THEN** o sistema atualiza o usuário e retorna os dados atualizados

#### Scenario: Usuário inexistente na atualização
- **WHEN** a requisição informar identificador inexistente
- **THEN** o sistema retorna erro de usuário não encontrado

#### Scenario: E-mail duplicado na atualização
- **WHEN** a requisição informar e-mail já cadastrado por outro usuário
- **THEN** o sistema recusa a atualização e retorna erro de e-mail duplicado

### Requirement: Delete user
O sistema SHALL remover um usuário pelo identificador, ou erro quando não existir.

#### Scenario: Remover usuário com sucesso
- **WHEN** a requisição informar identificador existente
- **THEN** o sistema remove o usuário e confirma a remoção

#### Scenario: Remover usuário inexistente
- **WHEN** a requisição informar identificador inexistente
- **THEN** o sistema retorna erro de usuário não encontrado
