FROM golang:1.15-alpine as builder

# Fetch certs to allow use of TLS.
RUN apk add --update --no-cache ca-certificates

WORKDIR /build

# Fetch dependencies.
COPY go.mod go.sum ./
RUN go mod download

# Copy code.
COPY cmd ./cmd
COPY internal ./internal
COPY docs ./docs
COPY .env .

# Build the command inside the container.
RUN CGO_ENABLED=0 GOOS=linux go build -v -o main ./cmd/server/main.go

FROM scratch

# Copy the binary to the production image from the builder stage.
COPY --from=builder /build/main /main
COPY --from=builder /build/.env /.env
COPY --from=builder /build/docs /docs

# Copy CA certificates (allows HTTPS calls).
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 9001

# Entrypoint.
ENTRYPOINT ["/main"]
