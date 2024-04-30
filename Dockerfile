FROM golang:1.22.2-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o carsApp ./cmd/cars

FROM alpine:latest

RUN mkdir /app

COPY --from=builder /app/carsApp /app/carsApp

CMD [ "/app/carsApp" ]