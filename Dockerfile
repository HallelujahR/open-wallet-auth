FROM node:20-alpine AS admin-builder

WORKDIR /src/admin-web

COPY admin-web/package*.json ./
RUN npm ci

COPY admin-web/ ./
RUN npm run build

FROM golang:1.23-alpine AS builder

WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/open-wallet-auth ./cmd/server

FROM alpine:3.20

RUN addgroup -S app && adduser -S app -G app
WORKDIR /app

COPY --from=builder /out/open-wallet-auth /app/open-wallet-auth
COPY --from=admin-builder /src/admin-web/dist /app/admin-web/dist
COPY configs/config.example.yaml /app/configs/config.example.yaml
COPY migrations /app/migrations

USER app
EXPOSE 8080

ENTRYPOINT ["/app/open-wallet-auth"]
