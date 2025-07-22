FROM golang:1.23-bookworm AS builder
WORKDIR /seed
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o seed

FROM debian:bookworm-slim
WORKDIR /seed
COPY --from=builder /seed/seed /seed/seed
CMD ["/seed/seed"]
