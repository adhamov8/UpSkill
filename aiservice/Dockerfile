FROM golang:1.23-alpine

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o aiservice ./cmd/ai-service

EXPOSE 8082
CMD ["./aiservice"]
