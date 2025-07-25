FROM golang:1.24.3 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /gobalancer ./cmd/main.go

FROM gcr.io/distroless/static

WORKDIR /
COPY --from=builder /gobalancer .
COPY servers.json ./servers.json

ENTRYPOINT ["/gobalancer"]