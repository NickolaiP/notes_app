package hand

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"time"

	"noteApi/internal/database"
	"noteApi/internal/logger"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Ключ для подписывания JWT токенов, загружается из переменных окружения.
var jwtKey = []byte(os.Getenv("JWT_KEY"))

// Claims определяет структуру полезной нагрузки JWT токена.
type Claims struct {
	Username             string `json:"username"` // Имя пользователя, для которого выдан токен.
	jwt.RegisteredClaims        // Стандартные зарегистрированные поля JWT.
}

// UserHandler содержит логику для обработки запросов, связанных с пользователями.
type UserHandler struct {
	db     database.Database // Интерфейс для работы с базой данных.
	logger *logger.Logger    // Логгер для записи сообщений и ошибок.
}

// NewUserHandler создает новый экземпляр UserHandler с заданными зависимостями.
func NewUserHandler(db database.Database, logger *logger.Logger) *UserHandler {
	return &UserHandler{
		db:     db,
		logger: logger,
	}
}

// Register обрабатывает запросы на регистрацию нового пользователя.
// Хэширует пароль и сохраняет данные пользователя в базе данных.
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Создание контекста с таймаутом для запроса к базе данных.
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Извлечение имени пользователя и пароля из формы запроса.
	username := r.FormValue("username")
	password := r.FormValue("password")

	// Хэширование пароля с использованием bcrypt перед сохранением в базе данных.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		// Возвращение ошибки, если хэширование не удалось.
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	// Вставка нового пользователя в базу данных и получение его ID.
	var userID int
	err = h.db.QueryRow(ctx, "INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id",
		username, string(hashedPassword)).Scan(&userID)
	if err != nil {
		// Возвращение ошибки, если вставка в базу данных не удалась.
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	// Отправка успешного ответа клиенту.
	w.Write([]byte("User registered successfully"))
}

// Login обрабатывает запросы на авторизацию пользователя.
// Проверяет учетные данные и возвращает JWT токен в cookie при успешной авторизации.
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Создание контекста с таймаутом для запроса к базе данных.
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Извлечение имени пользователя и пароля из формы запроса.
	username := r.FormValue("username")
	password := r.FormValue("password")

	// Получение хэшированного пароля и ID пользователя из базы данных.
	var storedPassword string
	var userID int
	err := h.db.QueryRow(ctx, "SELECT id, password FROM users WHERE username=$1", username).Scan(&userID, &storedPassword)
	if err == sql.ErrNoRows || bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password)) != nil {
		// Возвращение ошибки авторизации, если пользователь не найден или пароль неверный.
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Установка времени истечения токена на 24 часа от текущего времени.
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// Создание нового JWT токена с использованием алгоритма HMAC-SHA256.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		// Возвращение ошибки генерации токена, если что-то пошло не так.
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Установка JWT токена в cookie, который будет использоваться для аутентификации пользователя.
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

	// Отправка успешного ответа клиенту с установкой статуса 200 OK.
	w.WriteHeader(http.StatusOK)
}
