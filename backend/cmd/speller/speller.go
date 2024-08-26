package speller

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"noteApi/internal/database"
	"strings"
	"time"
)

// CreateNoteHandler возвращает обработчик HTTP-запросов для создания заметки
// с проверкой орфографии текста и сохранением в базу данных.
func CreateNoteHandler(db database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Создаем контекст с таймаутом для обработки запроса.
		// Если выполнение операции затянется, контекст будет отменен.
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Извлекаем данные из запроса.
		username := r.Header.Get("username")
		text := r.FormValue("text")

		// Проверяем орфографию текста с помощью Яндекс.Спеллер API.
		correctedText, err := checkSpelling(ctx, text)
		if err != nil {
			// Если произошла ошибка при проверке орфографии, возвращаем ошибку 500.
			http.Error(w, "Error checking spelling", http.StatusInternalServerError)
			return
		}

		// Находим ID пользователя по имени.
		var userID int
		err = db.QueryRow(ctx, "SELECT id FROM users WHERE username=$1", username).Scan(&userID)
		if err == sql.ErrNoRows {
			// Если пользователь не найден, возвращаем ошибку 404.
			http.Error(w, "User not found", http.StatusNotFound)
			return
		} else if err != nil {
			// Если произошла ошибка при выполнении запроса, возвращаем ошибку 500.
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		// Сохраняем заметку в базу данных.
		_, err = db.Exec(ctx, "INSERT INTO notes (user_id, text) VALUES ($1, $2)", userID, correctedText)
		if err != nil {
			// Если произошла ошибка при сохранении заметки, возвращаем ошибку 500.
			http.Error(w, "Error creating note", http.StatusInternalServerError)
			return
		}

		// Отправляем успешный ответ.
		w.Write([]byte("Note created successfully"))
	}
}

// checkSpelling проверяет орфографию текста с использованием Яндекс.Спеллер API.
// Возвращает исправленный текст и ошибку, если таковая имеется.
func checkSpelling(ctx context.Context, text string) (string, error) {
	// Подготовка запроса к Яндекс.Спеллер API.
	apiURL := "https://speller.yandex.net/services/spellservice.json/checkText"
	data := url.Values{}
	data.Set("text", text)

	// Создаем новый HTTP-запрос с привязанным контекстом.
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		// Если произошла ошибка при создании запроса, возвращаем ошибку.
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Выполняем запрос к API.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// Если произошла ошибка при выполнении запроса, возвращаем ошибку.
		return "", err
	}
	defer resp.Body.Close()

	// Читаем и обрабатываем ответ от API.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// Если произошла ошибка при чтении ответа, возвращаем ошибку.
		return "", err
	}

	// Парсим JSON-ответ от API.
	var result []struct {
		Word string   `json:"word"`
		S    []string `json:"s"`
		Pos  int      `json:"pos"`
		Row  int      `json:"row"`
		Col  int      `json:"col"`
		Len  int      `json:"len"`
		Code int      `json:"code"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		// Если произошла ошибка при парсинге ответа, возвращаем ошибку.
		return "", err
	}

	// Заменяем слова с ошибками на исправленные версии в тексте.
	for _, item := range result {
		if len(item.S) > 0 {
			// Заменяем слово с ошибкой на первое предложение исправления.
			text = strings.Replace(text, item.Word, item.S[0], 1)
		}
	}

	// Возвращаем исправленный текст.
	return text, nil
}
