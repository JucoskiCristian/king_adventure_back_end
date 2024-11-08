package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/crypto/bcrypt"

	_ "github.com/lib/pq"
)

type User struct {
	ID         int    `json:"id"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Status     string `json:"status"`
	Created_at string `json:"created_at"`
	Updated_at string `json:"updated_at"`
}

type Score struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Score    int    `json:"score"`
}

var db *sql.DB

func main() {
	var err error
	connStr := os.Getenv("DATABASE_URL")

	if connStr == "" {
		log.Fatalf("Variável de ambiente DATABASE_URL não está configurada")
	}

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Erro ao conectar com o banco de dados: %v", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("Erro ao conectar com o banco de dados: %v", err)
	}

	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/score", addScoreHandler)
	http.HandleFunc("/scores", getTopScoresHandler)
	http.HandleFunc("/docs", docsHandler)

	port := os.Getenv("PORT")
	if port == "" {
			port = "8080"
	}
	
	fmt.Printf("Servidor rodando em http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}

// Função para registrar um usuário com verificação de duplicidade
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
		return
	}

	// Verifica se o username já existe no banco de dados
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username=$1)", user.Username).Scan(&exists)
	if err != nil {
		log.Printf("Erro ao verificar duplicidade de usuário: %v", err)
		http.Error(w, "Erro ao verificar duplicidade de usuário", http.StatusInternalServerError)
		return
	}

	if exists {
		http.Error(w, "Usuário já existe", http.StatusConflict)
		return
	}

	// Insere o novo usuário no banco de dados
	_, err = db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", user.Username, user.Password)
	if err != nil {
		log.Printf("Erro ao inserir usuário: %v", err)
		http.Error(w, "Erro ao registrar usuário", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Usuário %s registrado com sucesso", user.Username)
}

// Função para fazer login
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
			return
	}

	var userID int
	var hashedPassword string
	err := db.QueryRow("SELECT id, password FROM users WHERE username = $1", user.Username).Scan(&userID, &hashedPassword)
	if err != nil {
			if err == sql.ErrNoRows {
					http.Error(w, "Credenciais inválidas", http.StatusUnauthorized)
			} else {
					log.Printf("Erro ao consultar banco de dados: %v", err)
					http.Error(w, "Erro ao acessar o banco de dados", http.StatusInternalServerError)
			}
			return
	}

	// Compara a senha recebida com o hash armazenado
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(user.Password)); err != nil {
			http.Error(w, "Credenciais inválidas", http.StatusUnauthorized)
			return
	}

	// Retorna o ID e o nome do usuário após o login bem-sucedido
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
			"message":  "Login bem-sucedido",
			"user_id":  userID,
			"username": user.Username,
	})
}


// Função para adicionar um score
func addScoreHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var score Score
	if err := json.NewDecoder(r.Body).Decode(&score); err != nil {
		http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
		return
	}

	// Insere o score no banco de dados
	_, err := db.Exec("INSERT INTO score (user_id, score) VALUES ($1, $2)", score.UserID, score.Score)
	if err != nil {
		log.Printf("Erro ao inserir score: %v", err)
		http.Error(w, "Erro ao registrar score", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Score %d registrado com sucesso para o usuário ID %d", score.Score, score.UserID)
}

func getTopScoresHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query(`
		SELECT users.id, users.username, score.score
		FROM score
		JOIN users ON score.user_id = users.id
		ORDER BY score.score DESC
		LIMIT 10
	`)
	if err != nil {
		log.Printf("Erro ao consultar scores: %v", err)
		http.Error(w, "Erro ao listar scores", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var scores []Score
	for rows.Next() {
		var score Score
		if err := rows.Scan(&score.UserID, &score.Username, &score.Score); err != nil {
			log.Printf("Erro ao escanear score: %v", err)
			http.Error(w, "Erro ao listar scores", http.StatusInternalServerError)
			return
		}
		scores = append(scores, score)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(scores); err != nil {
		http.Error(w, "Erro ao gerar resposta JSON", http.StatusInternalServerError)
		return
	}
}
// Função para servir a página de documentação
func docsHandler(w http.ResponseWriter, r *http.Request) {
	htmlContent := `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Documentação da API</title>
    <style>
      body {
        font-family: Arial, sans-serif;
        margin: 0;
        padding: 0;
        background-color: #f7f7f7;
      }
      header {
        background-color: #333;
        color: #fff;
        padding: 20px;
        text-align: center;
      }
      .container {
        padding: 20px;
        max-width: 800px;
        margin: 0 auto;
        background-color: #fff;
        box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
      }
      h1,h2 {
        color: #fff;
				
      }
      .endpoint {
        margin-bottom: 20px;
      }
      .endpoint h3 {
        margin-top: 0;
      }
      .method {
        display: inline-block;
        padding: 5px 10px;
        margin-bottom: 10px;
        font-size: 0.9em;
        font-weight: bold;
        color: #fff;
        border-radius: 5px;
      }
      .post {
        background-color: #4caf50;
      }
      .get {
        background-color: #2196f3;
      }
      .description {
        margin: 10px 0;
      }
      pre {
        background-color: #f4f4f4;
        padding: 10px;
        border-radius: 5px;
        overflow-x: auto;
      }
    </style>
  </head>
  <body>
    <header>
      <h1>Documentação da API</h1>
      <h2>
        Bem-vindo à documentação da API. Abaixo estão os detalhes para cada endpoint
        disponível.
      </h2>
    </header>
    <div class="container">
      <div class="endpoint">
        <h3>/register</h3>
        <span class="method post">POST</span>
        <p class="description">Registrar novos usuários.</p>
        <h4>Corpo da Requisição</h4>
        <pre>
{
    "username": "string",
    "password": "string"
}</pre>
        <h4>Resposta</h4>
        <pre>
{
    "message": "Usuário 'NOME_DO_USUÁRIO' registrado com sucesso"
}</pre>
      </div>

      <div class="endpoint">
        <h3>/login</h3>
        <span class="method post">POST</span>
        <p class="description">
          Retorna uma mensagem de sucesso para login, junto com o ID e nome de usuário.
        </p>
        <h4>Corpo da Requisição</h4>
        <pre>
{
    "username": "string",
    "password": "string"
}</pre>
        <h4>Resposta</h4>
        <pre>
{
  "message":  "Login bem-sucedido",
  "user_id":  "integer",
  "username": "string"
}</pre>
      </div>

      <div class="endpoint">
        <h3>/score</h3>
        <span class="method post">POST</span>
        <p class="description">Adicionar uma pontuação para um usuário.</p>
        <h4>Corpo da Requisição</h4>
        <pre>
{
    "user_id": "integer",
    "score": "integer"
}</pre>
        <h4>Resposta</h4>
        <pre>
{
    "message": "Pontuação 'score' registrada com sucesso para o usuário ID 'user_id'"
}</pre>
      </div>

      <div class="endpoint">
        <h3>/scores</h3>
        <span class="method get">GET</span>
        <p class="description">Retorna o top 10 de pontuações em ordem decrescente.</p>
        <h4>Resposta</h4>
        <pre>
[
    {
        "user_id": "integer",
        "username": "string",
        "score": "integer"
    },
    ...
]</pre>
      </div>
    </div>
  </body>
</html>
	`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(htmlContent))
}