FROM --platform=$BUILDPLATFORM golang:1.23.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o main ./cmd/server

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .
RUN chmod +x main

EXPOSE 8080

CMD ["./main"]