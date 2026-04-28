FROM golang:1.22-alpine AS builder

WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/open-wallet-auth ./cmd/server

FROM alpine:3.20

RUN addgroup -S app && adduser -S app -G app
WORKDIR /app

COPY --from=builder /out/open-wallet-auth /app/open-wallet-auth
COPY configs/config.example.yaml /app/configs/config.example.yaml

USER app
EXPOSE 8080

ENTRYPOINT ["/app/open-wallet-auth"]
