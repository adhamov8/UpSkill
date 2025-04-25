FROM golang:1.23-alpine

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o userservice ./cmd/user-service


RUN mkdir -p static/avatars
#COPY --chmod=755 static ./static

EXPOSE 8081
CMD ["./userservice"]
