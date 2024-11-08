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

4. No diretório raiz do projeto, crie um arquivo `.env` com o seguinte conteúdo:

   ```plaintext
   DATABASE_URL=postgresql://<USUARIO>:<SENHA>@<HOST>:<PORTA>/<NOME_DO_BANCO>
   PORT=8080
   ```

5. Instale uma biblioteca para carregar as variáveis de ambiente, como [Godotenv](github.com/joho/godotenv)

   ```shel
   go get github.com/joho/godotenv
   ```

6. No arquivo main.go, adicione o código para carregar as variáveis de ambiente:

   ```go
   import (
       "os"
       "log"
       "github.com/joho/godotenv"
   )

   func main() {
       // Carrega as variáveis do arquivo .env
       if err := godotenv.Load(); err != nil {
           log.Println("Nenhum arquivo .env encontrado")
       }

       connStr := os.Getenv("DATABASE_URL")
       if connStr == "" {
           log.Fatalf("Variável de ambiente DATABASE_URL não está configurada")
       }

       // ... restante do código
   }
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
  "user_id": "userID",
  "username": "nome_usuario",
  "password": "senha_segura"
}
```

    Resposta de Sucesso: 200 OK

3. Adicionar um Score:

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
