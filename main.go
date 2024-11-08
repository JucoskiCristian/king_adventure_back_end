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
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>API Documentation</title>
		<style>
			body { font-family: Arial, sans-serif; margin: 0; padding: 20px; }
			h1 { color: #333; }
			h2 { color: #555; }
			pre { background-color: #f4f4f4; padding: 10px; border-radius: 5px; }
		</style>
	</head>
	<body>
		<h1>API Documentation</h1>
		<p>Bem-vindo à documentação da API. Abaixo, você encontrará detalhes sobre cada rota disponível, o método HTTP e os parâmetros esperados.</p>
		
		<h2>Endpoints</h2>
		
		<h3>1. Register User</h3>
		<p><strong>Rota:</strong> <code>/register</code></p>
		<p><strong>Método:</strong> POST</p>
		<p><strong>Descrição:</strong> Registra um novo usuário.</p>
		<p><strong>Body:</strong></p>
		<pre>{
    "username": "string",
    "password": "string"
}</pre>

		<h3>2. Login</h3>
		<p><strong>Rota:</strong> <code>/login</code></p>
		<p><strong>Método:</strong> POST</p>
		<p><strong>Descrição:</strong> Faz login para o usuário.</p>
		<p><strong>Body:</strong></p>
		<pre>{
    "username": "string",
    "password": "string"
}</pre>

		<h3>3. Add Score</h3>
		<p><strong>Rota:</strong> <code>/score</code></p>
		<p><strong>Método:</strong> POST</p>
		<p><strong>Descrição:</strong> Adiciona um novo score para o usuário.</p>
		<p><strong>Body:</strong></p>
		<pre>{
    "user_id": "integer",
    "score": "integer"
}</pre>

		<h3>4. Top Scores</h3>
		<p><strong>Rota:</strong> <code>/scores</code></p>
		<p><strong>Método:</strong> GET</p>
		<p><strong>Descrição:</strong> Retorna o top 10 scores de usuários.</p>

	</body>
	</html>
	`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(htmlContent))
}