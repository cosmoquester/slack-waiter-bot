FROM golang:1.16.5-alpine AS builder

# Set necessary environmet variables needed for running on scratch
ENV CGO_ENABLED=0

WORKDIR /build

# Download Packages
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

# Build
RUN go build -o slack-waiter-bot .

FROM scratch

WORKDIR /app

# Use only compiled binary
COPY --from=builder /build/slack-waiter-bot .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8080

ENTRYPOINT [ "./slack-waiter-bot" ]
