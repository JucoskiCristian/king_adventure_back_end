package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Status string `json:"status"`
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
	connStr := "postgresql://postgres:<SENHA>@<HOST>:<PORTA>/postgres" // Substitua com suas credenciais corretas
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Erro ao conectar com o banco de dados: %v", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("Erro ao conectar com o banco de dados: %v", err)
	}

	http.HandleFunc("/register", registerHandler) //Adiciona usuario
	http.HandleFunc("/login", loginHandler) //Faz login
	http.HandleFunc("/users", listUsersHandler)//Retorna todos os usuarios
	http.HandleFunc("/score", addScoreHandler) //Adiciona score
	http.HandleFunc("/scores", getTopScoresHandler) //retona 10 maioras scores  

	fmt.Println("Servidor rodando em http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}

// Função para registrar um usuário
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

	_, err := db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", user.Username, user.Password)
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

	var storedPassword string
	err := db.QueryRow("SELECT password FROM users WHERE username = $1", user.Username).Scan(&storedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Credenciais inválidas", http.StatusUnauthorized)
		} else {
			log.Printf("Erro ao consultar banco de dados: %v", err)
			http.Error(w, "Erro ao acessar o banco de dados", http.StatusInternalServerError)
		}
		return
	}

	if storedPassword != user.Password {
		http.Error(w, "Credenciais inválidas", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Login bem-sucedido para o usuário %s", user.Username)
}

// Função para listar todos os usuários
func listUsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query("SELECT id, username, password, status, created_at, updated_at FROM users")
	if err != nil {
		log.Printf("Erro ao consultar usuários: %v", err)
		http.Error(w, "Erro ao listar usuários", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.Password, &user.Status, &user.Created_at, &user.Updated_at); err != nil {
			log.Printf("Erro ao escanear usuário: %v", err)
			http.Error(w, "Erro ao listar usuários", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, "Erro ao gerar resposta JSON", http.StatusInternalServerError)
		return
	}
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
