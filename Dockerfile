FROM golang:1.25.6-alpine AS builder

ARG SERVICE
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o /bin/service \
    ./cmd/${SERVICE}

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /bin/service /service

EXPOSE 8080

ENTRYPOINT ["/service"]