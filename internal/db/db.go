package db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/phgermanov/tasks/internal/models"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(dbPath string) (*TaskRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	repo := &TaskRepository{db: db}
	if err := repo.createTable(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *TaskRepository) createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
		completed BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := r.db.Exec(query)
	return err
}

func (r *TaskRepository) Create(ctx context.Context, task *models.Task) error {
	query := `
	INSERT INTO tasks (title, description, completed, created_at)
	VALUES (?, ?, ?, ?)`

	task.CreatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query, task.Title, task.Description, task.Completed, task.CreatedAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	task.ID = int(id)
	return nil
}

func (r *TaskRepository) GetAll(ctx context.Context) ([]*models.Task, error) {
	query := `SELECT id, title, description, completed, created_at FROM tasks ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Completed, &task.CreatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

func (r *TaskRepository) GetByID(ctx context.Context, id int) (*models.Task, error) {
	query := `SELECT id, title, description, completed, created_at FROM tasks WHERE id = ?`

	task := &models.Task{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID, &task.Title, &task.Description, &task.Completed, &task.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (r *TaskRepository) Update(ctx context.Context, task *models.Task) error {
	query := `
	UPDATE tasks
	SET title = ?, description = ?, completed = ?
	WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, task.Title, task.Description, task.Completed, task.ID)
	return err
}

func (r *TaskRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM tasks WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *TaskRepository) Close() error {
	return r.db.Close()
}
