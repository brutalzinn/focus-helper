package database

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Init abre a conexão com o banco de dados e cria as tabelas.
func Init(filepath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}

	createTablesSQL := `CREATE TABLE IF NOT EXISTS wellbeing_checks (id INTEGER PRIMARY KEY, timestamp DATETIME, question TEXT, answer TEXT);`
	_, err = db.Exec(createTablesSQL)
	if err != nil {
		return nil, err
	}

	log.Println("Banco de dados inicializado com sucesso.")
	return db, nil
}

// LogWellbeingCheck salva uma resposta no banco de dados.
func LogWellbeingCheck(db *sql.DB, question, answer string) {
	stmt, err := db.Prepare("INSERT INTO wellbeing_checks(timestamp, question, answer) VALUES(?, ?, ?)")
	if err != nil {
		log.Printf("Erro ao preparar statement: %v", err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(time.Now(), question, answer)
	if err != nil {
		log.Printf("Erro ao inserir log de bem-estar: %v", err)
	}
}
