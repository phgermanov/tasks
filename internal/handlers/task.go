package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/phgermanov/tasks/internal/models"
)

type TaskHandler interface {
	Create(ctx context.Context, task *models.Task) error
	GetAll(ctx context.Context) ([]*models.Task, error)
	GetByID(ctx context.Context, id int) (*models.Task, error)
	Update(ctx context.Context, task *models.Task) error
	Delete(ctx context.Context, id int) error
	Close() error
}

type Task struct {
	Repo TaskHandler
}

func (h *Task) Create(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(task.Title) == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	if err := h.Repo.Create(r.Context(), &task); err != nil {
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *Task) Get(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.Repo.GetAll(r.Context())
	if err != nil {
		http.Error(w, "Failed to get tasks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (h *Task) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(task.Title) == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	task.ID = id

	if _, err := h.Repo.GetByID(r.Context(), id); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get task", http.StatusInternalServerError)
		return
	}

	if err := h.Repo.Update(r.Context(), &task); err != nil {
		http.Error(w, "Failed to update task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *Task) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	if _, err := h.Repo.GetByID(r.Context(), id); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get task", http.StatusInternalServerError)
		return
	}

	if err := h.Repo.Delete(r.Context(), id); err != nil {
		http.Error(w, "Failed to delete task", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
