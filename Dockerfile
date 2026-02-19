FROM golang:1.24.4-alpine AS build

WORKDIR /build

COPY go.mod ./

RUN go mod download

COPY . .

RUN go build server/main.go

EXPOSE 8000

# Run the application
CMD ["./main"]



