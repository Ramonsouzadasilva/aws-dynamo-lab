# Goal API

Uma API REST desenvolvida em **Go (Golang)** para gerenciamento de metas e tarefas, utilizando **DynamoDB** como banco de dados. O ambiente de desenvolvimento é provisionado com **Terraform** e executado localmente através do **LocalStack**, permitindo simular serviços da AWS sem custos.

## Tecnologias

- Go
- Chi Router
- DynamoDB
- AWS SDK for Go v2
- LocalStack
- Terraform
- Docker

## Pré-requisitos

- Go 1.24+
- Docker
- Docker Compose
- Terraform >= 1.8
- AWS CLI (opcional)

## Clonando o projeto

```bash
git clone https://github.com/seu-usuario/goal-api.git

cd goal-api
```

---

# Configuração

Crie um arquivo `.env`.

Exemplo:

```env
APP_PORT=8080

AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=test
AWS_SECRET_ACCESS_KEY=test

DYNAMODB_ENDPOINT=http://localhost:4566
DYNAMODB_TABLE=goals
```

---

# Subindo o LocalStack

```bash
docker compose up -d
```

Verifique se está executando:

```bash
docker ps
```

---

# Provisionando a infraestrutura

Entre na pasta do ambiente:

```bash
cd terraform/environments/local
```

Inicialize o Terraform:

```bash
terraform init
```

Visualize o plano:

```bash
terraform plan
```

Aplique a infraestrutura:

```bash
terraform apply
```

Ao final será criada a tabela do DynamoDB dentro do LocalStack.

---

# Executando a API

Na raiz do projeto:

```bash
go run ./cmd/api
```

A API ficará disponível em:

```
http://localhost:8080
```

---

# Endpoints

| Método | Endpoint    | Descrição      |
| ------ | ----------- | -------------- |
| GET    | /health     | Health Check   |
| POST   | /goals      | Criar meta     |
| GET    | /goals      | Listar metas   |
| GET    | /goals/{id} | Buscar meta    |
| PUT    | /goals/{id} | Atualizar meta |
| DELETE | /goals/{id} | Remover meta   |

---

# Estrutura da infraestrutura

O Terraform é responsável por provisionar:

- DynamoDB
- Tabelas
- Índices (GSI)
- Configuração da AWS LocalStack

Todo o provisionamento é executado localmente utilizando o endpoint do LocalStack.

---

# Comandos úteis

Inicializar Terraform

```bash
terraform init
```

Aplicar infraestrutura

```bash
terraform apply
```

Destruir infraestrutura

```bash
terraform destroy
```

---

# Variáveis de Ambiente

| Variável              | Descrição              |
| --------------------- | ---------------------- |
| APP_PORT              | Porta da aplicação     |
| AWS_REGION            | Região AWS             |
| AWS_ACCESS_KEY_ID     | Access Key             |
| AWS_SECRET_ACCESS_KEY | Secret Key             |
| DYNAMODB_ENDPOINT     | Endpoint do LocalStack |
| DYNAMODB_TABLE        | Nome da tabela         |

---
