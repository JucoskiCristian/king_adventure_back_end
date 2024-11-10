package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	// "github.com/joho/godotenv" // se for usar localhost desmente essa linha
	"github.com/rs/cors"
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
	// Descomente esse bloco para usar o aquivo .env em local host, comentando para deploy no render!
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal("Erro ao carregar o arquivo .env")
	// }

	// Agora que o arquivo foi carregado, acesse as variáveis
	connStr := os.Getenv("DATABASE_URL")

	if connStr == "" {
		log.Fatal("Variável de ambiente DATABASE_URL não está configurada")
	}
	var err error
	
	// Abra a conexão com o banco de dados
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Erro ao conectar com o banco de dados: %v", err)
	}
	defer db.Close()

	// Verifica se a conexão com o banco foi bem-sucedida
	if err = db.Ping(); err != nil {
		log.Fatalf("Erro ao conectar com o banco de dados: %v", err)
	}

	// Define as rotas
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/score", addScoreHandler)
	http.HandleFunc("/scores", getTopScoresHandler)
	http.HandleFunc("/docs", docsHandler)

	// Configuração do CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // Permite requisições da origem do front-end
		AllowedMethods: []string{"GET", "POST"}, // Métodos permitidos
		AllowedHeaders: []string{"Content-Type", "Authorization"}, // Cabeçalhos permitidos
	})

	// Envolva o servidor HTTP com o middleware CORS
	handler := c.Handler(http.DefaultServeMux)

	// Inicia o servidor na porta definida no .env ou na porta 8080 por padrão
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Servidor rodando em http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
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
		// Retorna mensagem informando que o usuário já existe
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, fmt.Sprintf(`{"error": "Usuário '%s' já existe"}`, user.Username), http.StatusConflict)
		return
	}

	// Criptografa a senha
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Erro ao criptografar senha: %v", err)
		http.Error(w, "Erro ao criptografar a senha", http.StatusInternalServerError)
		return
	}

	// Insere o novo usuário no banco de dados com a senha criptografada
	_, err = db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", user.Username, string(hashedPassword))
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

// Função para pegar os top scores
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
	htmlContent := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Documentação da API</title>
    <style>
        /* Reset */
        * { margin: 0; padding: 0; box-sizing: border-box; }
        
        /* Body and Layout */
        body {
            font-family: 'Roboto', sans-serif;
            background-color: #1d1f27;
            color: #eaeaea;
            line-height: 1.6;
            display: flex;
            justify-content: center;
            padding: 20px;
        }
        main {
            max-width: 800px;
            width: 100%;
            padding: 20px;
            background-color: #282a36;
            border-radius: 8px;
            box-shadow: 0 4px 8px rgba(0, 0, 0, 0.3);
        }
        
        /* Header */
        header {
            text-align: center;
            padding: 20px 0;
            background-color: #44475a;
            border-radius: 8px 8px 0 0;
        }
        header h1 {
            color: #f8f8f2;
            font-size: 1.8em;
            font-weight: 700;
        }

        /* Section Titles */
        h2 {
            color: #8be9fd;
            font-size: 1.5em;
            margin-bottom: 10px;
            border-bottom: 2px solid #6272a4;
            padding-bottom: 5px;
        }

        /* Content */
        section {
            margin-bottom: 25px;
        }
        p, ul {
            margin: 10px 0;
            color: #f8f8f2;
        }

        /* Code Blocks */
        code {
            background-color: #44475a;
            color: #50fa7b;
            padding: 5px 10px;
            border-radius: 4px;
            font-size: 0.95em;
            font-family: 'Courier New', Courier, monospace;
        }

        /* Lists */
        ul {
            padding-left: 20px;
            list-style-type: disc;
        }
        ul li {
            margin: 5px 0;
        }

        /* Responsive Typography */
        @media (max-width: 600px) {
            h1 { font-size: 1.5em; }
            h2 { font-size: 1.2em; }
        }
    </style>
</head>
<body>
    <header>
        <h1>Documentação da API</h1>
    </header>
    <main>
        <section>
            <h2>1. Registro de Usuário</h2>
            <p>Rota: <code>POST /register</code></p>
            <p>Descrição: Registra um novo usuário.</p>
            <p>Parâmetros:</p>
            <ul>
                <li><code>username</code>: Nome de usuário.</li>
                <li><code>password</code>: Senha do usuário.</li>
            </ul>
        </section>
        <section>
            <h2>2. Login de Usuário</h2>
            <p>Rota: <code>POST /login</code></p>
            <p>Descrição: Faz login de um usuário.</p>
            <p>Parâmetros:</p>
            <ul>
                <li><code>username</code>: Nome de usuário.</li>
                <li><code>password</code>: Senha do usuário.</li>
            </ul>
        </section>
        <section>
            <h2>3. Adicionar Score</h2>
            <p>Rota: <code>POST /score</code></p>
            <p>Descrição: Adiciona um score para um usuário.</p>
            <p>Parâmetros:</p>
            <ul>
                <li><code>user_id</code>: ID do usuário.</li>
                <li><code>score</code>: Valor do score.</li>
            </ul>
        </section>
        <section>
            <h2>4. Top Scores</h2>
            <p>Rota: <code>GET /scores</code></p>
            <p>Descrição: Retorna os top 10 scores.</p>
        </section>
    </main>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(htmlContent))
}