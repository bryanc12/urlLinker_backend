FROM golang:1.23.4-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o app .

FROM alpine:3.21.0 

COPY --from=builder /app/app /usr/local/bin/app

EXPOSE 3000

CMD ["app"]