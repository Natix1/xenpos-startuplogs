FROM golang:1.23-alpine

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o app

RUN chmod +x app
CMD ["./app"]