FROM golang:1.26

WORKDIR /code

RUN apt-get install -y ca-certificates 
RUN apt-get install -y python3

RUN ldconfig

COPY . .

RUN GOOS=js GOARCH=wasm go build -o wasm_client/main.wasm wasm_client/main.go

# Tell ldconfig we have new libraries to look at

EXPOSE 8080/tcp

WORKDIR wasm_client
# Run the application
CMD ["python3", "-m", "http.server", "8080"]



