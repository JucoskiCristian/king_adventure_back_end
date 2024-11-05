# Projeto API de Usuários e Scores

Este projeto é uma API em Go que permite gerenciar usuários e registrar pontuações (scores). A API inclui rotas para registrar, fazer login, listar usuários, adicionar scores, e listar os 10 menores scores associados a cada usuário.

## Pré-requisitos

- **Go**: Instale a versão mais recente do Go (https://golang.org/dl/).
- **Banco de Dados PostgreSQL**: Configure um banco de dados PostgreSQL no Supabase ou em um servidor local.

## Configuração

1. Clone o repositório e navegue até o diretório do projeto:

   ```sh
   git clone https://github.com/JucoskiCristian/king_adventure_back_end.git
   cd king_adventure_back_end
   ```

2. Instale as dependências:

   ```sh
   go mod init
   go mod tidy
   ```

3. No banco de dados, crie as tabelas necessárias executando as seguintes consultas SQL:

   ```sql
     CREATE TABLE users (
     id SERIAL PRIMARY KEY,
     username VARCHAR(50) UNIQUE NOT NULL,
     password VARCHAR(50) NOT NULL,
     status VARCHAR(10) DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
   );

   CREATE TABLE score (
     id SERIAL PRIMARY KEY,
     user_id INT REFERENCES users(id) ON DELETE CASCADE,
     score INT NOT NULL,
     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
   );
   ```

4. Configure a string de conexão no código Go:

   No arquivo main.go, substitua connStr pelas credenciais de acesso ao seu banco PostgreSQL.

   ```go
    connStr := "postgresql://postgres:<SENHA>@<HOST>:<PORTA>/postgres"
   ```

## Executando a Aplicação

Para iniciar o servidor local, execute o comando:

```sh
go run main.go
```

O servidor será iniciado em http://localhost:8080.

## Rotas Disponíveis

1. Registrar um Usuário

- Rota: POST /register

- Descrição: Registra um novo usuário.

- Exemplo de Requisição:

```json
{
  "username": "nome_usuario",
  "password": "senha_segura"
}
```

    Resposta de Sucesso: 201 Created

2. Fazer Login

- Rota: POST /login

- Descrição: Realiza o login de um usuário.

- Exemplo de Requisição:

```json
{
  "username": "nome_usuario",
  "password": "senha_segura"
}
```

    Resposta de Sucesso: 200 OK

3. Listar Todos os Usuários

- Rota: GET /users

- Descrição: Lista todos os usuários cadastrados, incluindo id, username, status, created_at e updated_at.

- Resposta de Exemplo:

```json
[
  {
    "id": 1,
    "username": "nome_usuario",
    "status": "active",
    "created_at": "2023-10-30T12:00:00Z",
    "updated_at": "2023-10-30T12:00:00Z"
  },
  {
    "id": 2,
    "username": "outro_usuario",
    "status": "inactive",
    "created_at": "2023-10-30T12:00:00Z",
    "updated_at": "2023-10-30T12:00:00Z"
  }
]
```

4. Adicionar um Score:

- Rota: POST /score

- Descrição: Adiciona um score para um usuário específico.

- Exemplo de Requisição:

```json
{
  "user_id": 1,
  "score": 100
}
```

    Resposta de Sucesso: 201 Created

5. Listar os 10 Maiores Scores

- Rota: GET /scores

- Descrição: Lista os 10 maiores scores em ordem decrescente, incluindo o nome do usuário, created_at e updated_at do score.

- Resposta de Exemplo:

```json
[
  {
    "user_id": 1,
    "username": "nome_usuario",
    "score": 50
  },
  {
    "user_id": 2,
    "username": "outro_usuario",
    "score": 60
  }
]
```
