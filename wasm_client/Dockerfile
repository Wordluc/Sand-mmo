FROM golang:1.26

WORKDIR /code

RUN apt-get install -y python3

COPY . .

WORKDIR wasm_client
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
	GOOS=js GOARCH=wasm go build -o main.wasm main.go

EXPOSE 80/tcp


CMD ["python3", "-m", "http.server", "80"]
