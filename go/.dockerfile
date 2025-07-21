# 1. First stage: Build the Go binary using an official Go image
FROM golang:1.23-bookworm AS builder

# 2. Set the working directory inside the container
WORKDIR /app

# 3. Copy go.mod and go.sum first, so dependencies can be cached
COPY go.mod go.sum ./

# 4. Download Go module dependencies (cached if unchanged)
RUN go mod download

# 5. Copy the rest of your application source code into the container
COPY . .

# 6. Build the Go binary, outputting to /app/server (you can change 'server')
RUN go build -v -o server

# 7. Second stage: Start with a minimal Debian image for running the binary
FROM debian:bookworm-slim

# 8. (Optional) Install certificates, required by HTTPS clients in Go
# RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
#     ca-certificates && \
#     rm -rf /var/lib/apt/lists/*

# 9. Copy the binary built in the first stage to the new image
COPY --from=builder /app/server /app/server

# 10. Set the command to run your application
CMD ["/app/server"]
