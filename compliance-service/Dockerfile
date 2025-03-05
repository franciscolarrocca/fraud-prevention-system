FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o compliance-service

EXPOSE 8080

CMD ["/app/compliance-service"]
