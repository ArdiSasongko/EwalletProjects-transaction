FROM golang:1.23.2-alpine AS bulder

WORKDIR /app

COPY go.mod .

COPY go.sum .

RUN go mod tidy

COPY . ./

COPY .env ./

RUN go build -o bin/main cmd/*.go

FROM alpine:latest

WORKDIR /app

COPY --from=bulder /app/bin/main ./bin/main

COPY --from=bulder /app/.env ./.env

EXPOSE 4002

EXPOSE 5002

CMD ["./bin/main"]