FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /progressservice ./cmd/progress-service

FROM scratch
COPY --from=build /progressservice /progressservice
ENTRYPOINT ["/progressservice"]
EXPOSE 8083
