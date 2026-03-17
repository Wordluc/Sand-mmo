FROM golang:1.26

WORKDIR /code

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
	go build -o . server/main.go

EXPOSE 8000/tcp
EXPOSE 8000/udp

# Run the application
CMD ["./main"]
