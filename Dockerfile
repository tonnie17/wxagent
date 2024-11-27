FROM golang:1.21.8-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod tidy && go build -o server cmd/server/main.go


FROM alpine:3.18

COPY --from=builder /app/server /server

RUN chmod +x /server

EXPOSE 8082

CMD ["/server"]
