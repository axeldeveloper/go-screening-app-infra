# project-constitution Specification

## Purpose

Define os princípios que governam as decisões de arquitetura e implementação do projeto e os disponibiliza tanto na raiz do repositório quanto no contexto do OpenSpec, para que toda mudança seja avaliada contra um referencial comum e estável.

## Requirements

### Requirement: Constituição publicada na raiz
O projeto MUST manter um arquivo `CONSTITUTION.md` na raiz do repositório contendo os princípios que governam decisões de arquitetura, qualidade, segurança e documentação.

#### Scenario: Consulta às diretrizes do projeto
- **WHEN** um colaborador ou agente procura as diretrizes do projeto na raiz do repositório
- **THEN** encontra `CONSTITUTION.md` com os princípios vigentes

### Requirement: Princípios alinhados à arquitetura em camadas
A constituição MUST refletir a arquitetura real do projeto, definindo a ordem das camadas `config → db → repository → service → handler/web` e a regra de que cada camada depende da imediatamente inferior.

#### Scenario: Decisão sobre onde alocar lógica
- **WHEN** um agente precisa decidir onde implementar uma regra de negócio
- **THEN** a constituição direciona a implementação para a camada de serviço e indica que handlers apenas adaptam requisições HTTP

#### Scenario: Dependências entre camadas
- **WHEN** um agente propõe uma dependência de uma camada superior em detalhes de infraestrutura de outra camada
- **THEN** a constituição classifica essa dependência como violação dos princípios definidos

### Requirement: Integração da constituição ao OpenSpec
A constituição MUST estar disponível como restrição de planejamento no OpenSpec por meio de `openspec/constitution.md` e da configuração do projeto, mantendo uma única fonte de verdade com o arquivo da raiz.

#### Scenario: Planejamento de uma mudança
- **WHEN** um agente cria ou implementa uma change OpenSpec
- **THEN** os princípios da constituição são carregados como contexto e considerados nas decisões de planejamento

#### Scenario: Atualização dos princípios
- **WHEN** um princípio é alterado no arquivo da raiz
- **THEN** a versão integrada ao OpenSpec reflete a mesma alteração, sem divergência de conteúdo

### Requirement: Idioma da constituição
A constituição MUST ser escrita em português, de forma legível por colaboradores e agentes, independentemente do idioma dos documentos pré-existentes no repositório.

#### Scenario: Leitura da constituição
- **WHEN** um agente ou colaborador lê `CONSTITUTION.md`
- **THEN** todo o conteúdo está redigido em português
