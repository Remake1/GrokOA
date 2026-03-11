FROM golang:1.26-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/crackoa-api ./cmd/app

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /out/crackoa-api /usr/local/bin/crackoa-api
COPY config.yaml /app/config.yaml

EXPOSE 8080

CMD ["crackoa-api"]
