FROM golang:1.25-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/stk-backend ./cmd/server

FROM alpine:3.21
WORKDIR /app
RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /out/stk-backend /app/stk-backend
COPY --from=builder /src/docs /app/docs

USER app
EXPOSE 8080
CMD ["/app/stk-backend"]
