FROM golang:1.23-alpine

# Устанавливаем git, так как он нужен для скачивания некоторых модулей
RUN apk add --no-cache git

WORKDIR /app

# Копируем go.mod и go.sum сначала
COPY go.mod go.sum ./

# Качаем зависимости
RUN go mod download

# Копируем остальной код
COPY . .

# Собираем
RUN go build -o main .

CMD ["./main"]
