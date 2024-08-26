package hand

import (
	"errors"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware возвращает middleware функцию, которая проверяет наличие и валидность JWT токена в cookie.
// Если токен отсутствует или недействителен, пользователь получает ответ с ошибкой авторизации.
// В противном случае, middleware устанавливает имя пользователя из токена в заголовки запроса и передает управление следующему обработчику.
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Извлечение JWT токена из cookie.
		cookie, err := r.Cookie("token")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				// Если cookie с токеном отсутствует, возвращаем ошибку авторизации.
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			// Если произошла ошибка при извлечении cookie, возвращаем ошибку запроса.
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Извлечение строки токена из cookie.
		tokenStr := cookie.Value
		claims := &Claims{}

		// Парсинг токена и извлечение его полезной нагрузки.
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			// Проверка подписи токена с использованием секретного ключа.
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			// Если токен не удалось распарсить или он недействителен, выводим ошибку в лог и возвращаем ошибку авторизации.
			log.Println("Token validation failed:", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Установка имени пользователя из токена в заголовки запроса для дальнейшего использования в обработчике.
		r.Header.Set("username", claims.Username)

		// Передача управления следующему обработчику.
		next.ServeHTTP(w, r)
	})
}
