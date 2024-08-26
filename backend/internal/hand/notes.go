package hand

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"noteApi/internal/database"
	"noteApi/internal/logger"
	"noteApi/internal/models"
)

type NoteHandler struct {
	db     database.Database
	logger *logger.Logger
}

func NewNoteHandler(db database.Database, logger *logger.Logger) *NoteHandler {
	return &NoteHandler{
		db:     db,
		logger: logger,
	}
}

func (h *NoteHandler) GetNotes(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	username := r.Header.Get("username")

	var userID int
	err := h.db.QueryRow(ctx, "SELECT id FROM users WHERE username=$1", username).Scan(&userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	rows, err := h.db.Query(ctx, "SELECT id, text FROM notes WHERE user_id=$1", userID)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var notes []models.Note
	for rows.Next() {
		var note models.Note
		if err := rows.Scan(&note.ID, &note.Text); err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		notes = append(notes, note)
	}

	json.NewEncoder(w).Encode(notes)
}

func (h *NoteHandler) CreateNote(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	username := r.Header.Get("username")
	text := r.FormValue("text")

	var userID int
	err := h.db.QueryRow(ctx, "SELECT id FROM users WHERE username=$1", username).Scan(&userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	_, err = h.db.Exec(ctx, "INSERT INTO notes (user_id, text) VALUES ($1, $2)", userID, text)
	if err != nil {
		http.Error(w, "Error creating note", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Note created successfully"))
}

func (h *NoteHandler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	username := r.Header.Get("username")
	noteID := r.URL.Query().Get("id")

	var userID int
	err := h.db.QueryRow(ctx, "SELECT id FROM users WHERE username=$1", username).Scan(&userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	_, err = h.db.Exec(ctx, "DELETE FROM notes WHERE id=$1 AND user_id=$2", noteID, userID)
	if err != nil {
		http.Error(w, "Error deleting note", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Note deleted successfully"))
}
