FROM golang:1.23-alpine

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o gateway ./cmd/gateway

EXPOSE 8000
CMD ["./gateway"]
