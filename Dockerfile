FROM golang:1.23-alpine

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY ./src/*.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o app

RUN chmod +x app
CMD ["./app"]