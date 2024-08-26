package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/NickolaiP/notes_app/backend/cmd/speller"
	"github.com/NickolaiP/notes_app/backend/internal/config"
	"github.com/NickolaiP/notes_app/backend/internal/database"
	"github.com/NickolaiP/notes_app/backend/internal/hand"
	"github.com/NickolaiP/notes_app/backend/internal/logger"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	// Загрузка конфигурации приложения из файла
	cfg := config.LoadConfig()

	// Инициализация логгера, который будет выводить логи в стандартный вывод (stdout)
	logger := logger.InitLogger(os.Stdout) // Перенастроили вывод логов в стандартный вывод

	// Инициализация подключения к базе данных
	db, err := database.NewPostgresDB(cfg.DB)
	if err != nil {
		// Логирование ошибки подключения к базе данных и завершение работы программы
		logger.Error("Failed to connect to database", "error", err)
		return
	}
	defer db.Close() // Закрытие подключения к базе данных при завершении работы программы

	// Инициализация маршрутизатора для обработки HTTP-запросов
	r := mux.NewRouter()

	// Инициализация обработчиков запросов
	userHandler := hand.NewUserHandler(db, logger)
	noteHandler := hand.NewNoteHandler(db, logger)

	// Настройка маршрутов для регистрации, входа, получения, создания и удаления заметок
	r.HandleFunc("/register", userHandler.Register).Methods("POST")
	r.HandleFunc("/login", userHandler.Login).Methods("POST")
	r.HandleFunc("/notes", hand.AuthMiddleware(noteHandler.GetNotes)).Methods("GET")
	r.HandleFunc("/notes", hand.AuthMiddleware(speller.CreateNoteHandler(db))).Methods("POST")
	r.HandleFunc("/notes", hand.AuthMiddleware(noteHandler.DeleteNote)).Methods("DELETE")

	// Создание и настройка HTTP-сервера
	server := &http.Server{
		Addr: ":8000",
		Handler: handlers.CORS(
			handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "OPTIONS"}),
			handlers.AllowedHeaders([]string{"Authorization", "Content-Type"}),
		)(r),
	}

	// Запуск сервера в горутине для обработки запросов
	go func() {
		logger.Info("Server started on :8000")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Логирование ошибки, если сервер не смог запуститься
			logger.Error("Could not listen on :8000", "error", err)
		}
	}()

	// Обработка сигналов прерывания (например, Ctrl+C) для корректного завершения работы сервера
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	// Создание контекста с таймаутом для корректного завершения работы сервера
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		// Логирование ошибки, если сервер не смог корректно завершить работу
		logger.Error("Server forced to shutdown", "error", err)
	}
	// Логирование успешного завершения работы сервера
	logger.Info("Server exiting")
}
