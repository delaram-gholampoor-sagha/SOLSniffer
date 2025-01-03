FROM golang:1.20-alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy

COPY . .

RUN mkdir -p /app/configs && cp configs/configs.yml /app/configs/

RUN go build -o solsniffer cmd/main.go

EXPOSE 8080

CMD ["./solsniffer"]
