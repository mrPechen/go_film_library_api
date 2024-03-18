FROM golang:latest

WORKDIR /app

COPY . .

RUN go mod download

WORKDIR /app/src

CMD ["go", "run", "main.go"]
