package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"noteApi/cmd/speller"
	"noteApi/internal/config"
	"noteApi/internal/database"
	"noteApi/internal/hand"
	"noteApi/internal/logger"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	// Загрузка конфигурации
	cfg := config.LoadConfig()

	// Инициализация логгера
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	logger := logger.InitLogger(logFile)
	defer logFile.Close()

	// Инициализация базы данных
	db, err := database.NewPostgresDB(cfg.DB)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		return
	}
	defer db.Close()

	// Инициализация роутера
	r := mux.NewRouter()

	// Инициализация хендлеров
	userHandler := hand.NewUserHandler(db, logger)
	noteHandler := hand.NewNoteHandler(db, logger)

	// Настройка маршрутов
	r.HandleFunc("/register", userHandler.Register).Methods("POST")
	r.HandleFunc("/login", userHandler.Login).Methods("POST")
	r.HandleFunc("/notes", hand.AuthMiddleware(noteHandler.GetNotes)).Methods("GET")
	r.HandleFunc("/notes", hand.AuthMiddleware(speller.CreateNoteHandler(db))).Methods("POST")
	r.HandleFunc("/notes", hand.AuthMiddleware(noteHandler.DeleteNote)).Methods("DELETE")

	// Создание сервера
	server := &http.Server{
		Addr: ":8000",
		Handler: handlers.CORS(
			handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "OPTIONS"}),
			handlers.AllowedHeaders([]string{"Authorization", "Content-Type"}),
		)(r),
	}

	// Запуск сервера в горутине
	go func() {
		logger.Info("Server started on :8000")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Could not listen on :8000", "error", err)
		}
	}()

	// Обработка сигналов прерывания и завершение работы
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}
	logger.Info("Server exiting")
}
