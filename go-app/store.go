package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

var db *sql.DB

func InitDB() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	log.Println("Successfully connected to database!")
	createTable()
}

func createTable() {
	query := `
    CREATE TABLE IF NOT EXISTS tasks (
        id SERIAL PRIMARY KEY,
        title VARCHAR(255) NOT NULL,
        description TEXT,
        status VARCHAR(50) DEFAULT 'pending',
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatalf("Error creating tasks table: %v", err)
	}
}

func GetTasks() ([]Task, error) {
	rows, err := db.Query("SELECT id, title, description, status, created_at, updated_at FROM tasks ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func GetTask(id int) (*Task, error) {
	row := db.QueryRow("SELECT id, title, description, status, created_at, updated_at FROM tasks WHERE id = $1", id)
	var t Task
	if err := row.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}
	return &t, nil
}

func CreateTask(task *Task) error {
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	query := `INSERT INTO tasks (title, description, status, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := db.QueryRow(query, task.Title, task.Description, task.Status, task.CreatedAt, task.UpdatedAt).Scan(&task.ID)
	return err
}

func UpdateTask(id int, task *Task) error {
	task.UpdatedAt = time.Now()
	query := `UPDATE tasks SET title = $1, description = $2, status = $3, updated_at = $4
              WHERE id = $5`
	result, err := db.Exec(query, task.Title, task.Description, task.Status, task.UpdatedAt, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows // Not found
	}
	return nil
}

func DeleteTask(id int) error {
	result, err := db.Exec("DELETE FROM tasks WHERE id = $1", id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows // Not found
	}
	return nil
}