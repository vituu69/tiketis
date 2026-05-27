FROM golang:1.24-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/ledger ./cmd/main.go
RUN GOBIN=/out go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.3

#minimal runtime image
FROM alpine:3.21
WORKDIR /app

COPY --from=builder /out/ingressos /usr/local/bin/compra
COPY --from=builder /out/migrate /usr/local/bin/migrate
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /src/docs ./docs
COPY --from=builder /src/postgres/migration ./postgres/migration
COPY docker-entrypoint /usr/local/bin/entrypoint

ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt \
    TZ=UTC

RUN chmod +x /usr/local/bin/entrypoint && \
	sed -i 's/\r$//' /usr/local/bin/entrypoint && \
	adduser -D -u 10001 appuser && \
	chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

ENTRYPOINT [ "/bin/sh", "/usr/local/bin/entrypoint" ]