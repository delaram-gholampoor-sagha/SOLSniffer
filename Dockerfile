FROM golang:1.20-alpine
WORKDIR /app
COPY . .
RUN go build -o solsniffer
CMD ["./solsniffer"]