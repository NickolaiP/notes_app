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

func CreateNoteHandler(db database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Создание контекста с таймаутом для обработки запроса
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		username := r.Header.Get("username")
		text := r.FormValue("text")

		// Проверка текста через Яндекс.Спеллер API
		correctedText, err := checkSpelling(ctx, text)
		if err != nil {
			http.Error(w, "Error checking spelling", http.StatusInternalServerError)
			return
		}

		var userID int
		err = db.QueryRow(ctx, "SELECT id FROM users WHERE username=$1", username).Scan(&userID)
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec(ctx, "INSERT INTO notes (user_id, text) VALUES ($1, $2)", userID, correctedText)
		if err != nil {
			http.Error(w, "Error creating note", http.StatusInternalServerError)
			return
		}

		w.Write([]byte("Note created successfully"))
	}
}

func checkSpelling(ctx context.Context, text string) (string, error) {
	// Подготовка запроса
	apiURL := "https://speller.yandex.net/services/spellservice.json/checkText"
	data := url.Values{}
	data.Set("text", text)

	// Выполнение запроса с использованием контекста
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Чтение и обработка ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Парсинг ответа
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
		return "", err
	}

	// Заменяем слова с ошибками на исправленные версии
	for _, item := range result {
		if len(item.S) > 0 {
			// Заменяем слово в исходном тексте
			text = strings.Replace(text, item.Word, item.S[0], 1)
		}
	}

	return text, nil
}
