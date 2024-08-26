# Используем официальный образ Golang для сборки приложения
FROM golang:1.22 AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go mod и sum файлы
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходные коды в рабочую директорию
COPY . .

# Устанавливаем рабочую директорию в месте, где находится main.go
WORKDIR /app/cmd/api

# Компилируем приложение
RUN go build -o /app/main .

# Используем минимальный образ для запуска
FROM ubuntu:jammy

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем скомпилированное приложение из сборочного контейнера
COPY --from=builder /app/main /app/main

# Устанавливаем wait-for-it
RUN apt-get update && apt-get install -y curl && \
    curl -o /usr/local/bin/wait-for-it https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh && \
    chmod +x /usr/local/bin/wait-for-it

# Открываем порт, который приложение будет использовать
EXPOSE 8000

# Запускаем wait-for-it и затем приложение
CMD ["wait-for-it", "db:5432", "--", "./main"]
