package hand

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/NickolaiP/notes_app/backend/internal/database"
	"github.com/NickolaiP/notes_app/backend/internal/logger"
	"github.com/NickolaiP/notes_app/backend/internal/models"
)

// NoteHandler обрабатывает запросы, связанные с заметками (создание, получение и удаление).
type NoteHandler struct {
	db     database.Database
	logger *logger.Logger
}

// NewNoteHandler создает новый экземпляр NoteHandler с заданной базой данных и логгером.
// Аргументы:
//
//	db - интерфейс базы данных для выполнения запросов.
//	logger - логгер для записи сообщений о событиях и ошибках.
//
// Возвращает:
//
//	*NoteHandler - новый экземпляр NoteHandler.
func NewNoteHandler(db database.Database, logger *logger.Logger) *NoteHandler {
	return &NoteHandler{
		db:     db,
		logger: logger,
	}
}

// GetNotes обрабатывает запрос на получение всех заметок для текущего пользователя.
// Аргументы:
//
//	w - http.ResponseWriter для отправки ответа клиенту.
//	r - http.Request, содержащий запрос от клиента.
//
// Возвращает:
//
//	Ответ с JSON массивом заметок или ошибкой.
func (h *NoteHandler) GetNotes(w http.ResponseWriter, r *http.Request) {
	// Устанавливаем тайм-аут для запроса к базе данных.
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Получаем имя пользователя из заголовка запроса.
	username := r.Header.Get("username")

	// Запрашиваем идентификатор пользователя из базы данных.
	var userID int
	err := h.db.QueryRow(ctx, "SELECT id FROM users WHERE username=$1", username).Scan(&userID)
	if err != nil {
		// Если пользователь не найден, возвращаем ошибку 404.
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Запрашиваем все заметки пользователя из базы данных.
	rows, err := h.db.Query(ctx, "SELECT id, text FROM notes WHERE user_id=$1", userID)
	if err != nil {
		// Если произошла ошибка при выполнении запроса, возвращаем ошибку 500.
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Считываем все заметки и сохраняем их в слайс.
	var notes []models.Note
	for rows.Next() {
		var note models.Note
		if err := rows.Scan(&note.ID, &note.Text); err != nil {
			// Если произошла ошибка при чтении строки, возвращаем ошибку 500.
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		notes = append(notes, note)
	}

	// Кодируем заметки в JSON и отправляем клиенту.
	json.NewEncoder(w).Encode(notes)
}

// CreateNote обрабатывает запрос на создание новой заметки для текущего пользователя.
// Аргументы:
//
//	w - http.ResponseWriter для отправки ответа клиенту.
//	r - http.Request, содержащий запрос от клиента.
//
// Возвращает:
//
//	Ответ с сообщением об успешном создании заметки или ошибкой.
func (h *NoteHandler) CreateNote(w http.ResponseWriter, r *http.Request) {
	// Устанавливаем тайм-аут для запроса к базе данных.
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Получаем имя пользователя и текст заметки из заголовков и формы запроса.
	username := r.Header.Get("username")
	text := r.FormValue("text")

	// Запрашиваем идентификатор пользователя из базы данных.
	var userID int
	err := h.db.QueryRow(ctx, "SELECT id FROM users WHERE username=$1", username).Scan(&userID)
	if err != nil {
		// Если пользователь не найден, возвращаем ошибку 404.
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Вставляем новую заметку в базу данных.
	_, err = h.db.Exec(ctx, "INSERT INTO notes (user_id, text) VALUES ($1, $2)", userID, text)
	if err != nil {
		// Если произошла ошибка при выполнении запроса, возвращаем ошибку 500.
		http.Error(w, "Error creating note", http.StatusInternalServerError)
		return
	}

	// Отправляем клиенту сообщение об успешном создании заметки.
	w.Write([]byte("Note created successfully"))
}

// DeleteNote обрабатывает запрос на удаление заметки для текущего пользователя.
// Аргументы:
//
//	w - http.ResponseWriter для отправки ответа клиенту.
//	r - http.Request, содержащий запрос от клиента.
//
// Возвращает:
//
//	Ответ с сообщением об успешном удалении заметки или ошибкой.
func (h *NoteHandler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	// Устанавливаем тайм-аут для запроса к базе данных.
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Получаем имя пользователя из заголовка запроса и идентификатор заметки из параметров URL.
	username := r.Header.Get("username")
	noteID := r.URL.Query().Get("id")

	// Запрашиваем идентификатор пользователя из базы данных.
	var userID int
	err := h.db.QueryRow(ctx, "SELECT id FROM users WHERE username=$1", username).Scan(&userID)
	if err != nil {
		// Если пользователь не найден, возвращаем ошибку 404.
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Удаляем заметку из базы данных, если она принадлежит указанному пользователю.
	_, err = h.db.Exec(ctx, "DELETE FROM notes WHERE id=$1 AND user_id=$2", noteID, userID)
	if err != nil {
		// Если произошла ошибка при выполнении запроса, возвращаем ошибку 500.
		http.Error(w, "Error deleting note", http.StatusInternalServerError)
		return
	}

	// Отправляем клиенту сообщение об успешном удалении заметки.
	w.Write([]byte("Note deleted successfully"))
}
