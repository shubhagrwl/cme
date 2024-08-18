FROM golang:1.19 AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -o main .

EXPOSE 9090

CMD ["./main"]