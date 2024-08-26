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

var jwtKey = []byte(os.Getenv("JWT_KEY"))

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type UserHandler struct {
	db     database.Database
	logger *logger.Logger
}

func NewUserHandler(db database.Database, logger *logger.Logger) *UserHandler {
	return &UserHandler{
		db:     db,
		logger: logger,
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	username := r.FormValue("username")
	password := r.FormValue("password")

	// Хэширование пароля перед сохранением в БД
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	// Вставка нового пользователя в базу данных
	var userID int
	err = h.db.QueryRow(ctx, "INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id",
		username, string(hashedPassword)).Scan(&userID)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("User registered successfully"))
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	username := r.FormValue("username")
	password := r.FormValue("password")

	var storedPassword string
	var userID int
	err := h.db.QueryRow(ctx, "SELECT id, password FROM users WHERE username=$1", username).Scan(&userID, &storedPassword)
	if err == sql.ErrNoRows || bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password)) != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Установка времени истечения токена
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// Создание токена с использованием HMAC-SHA256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Установка токена в cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

	w.WriteHeader(http.StatusOK)
}
