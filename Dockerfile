# Multi-stage build for the web forum backend
FROM golang:1.23-bullseye AS builder
WORKDIR /src

# Cache dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/app ./cmd

# Minimal runtime image
FROM debian:12-slim
RUN useradd -m appuser
WORKDIR /home/appuser
COPY --from=builder /out/app /usr/local/bin/app

EXPOSE 3000
ENV PORT=3000
USER appuser
CMD ["/usr/local/bin/app"]
