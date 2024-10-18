FROM golang:latest

WORKDIR /app

# Копируем файлы go.mod и go.sum
COPY go.mod .
COPY go.sum .

# Загружаем зависимости
RUN go mod download

# Копируем весь проект
COPY . .

# Собираем приложение
RUN go build -o forum .

# Открываем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./forum"]
