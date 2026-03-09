FROM golang:1.26

WORKDIR /code

RUN apt-get update && apt-get install -y gcc make build-essential
RUN apt-get install -y libgl1-mesa-dev libxi-dev libxcursor-dev libxrandr-dev libxinerama-dev libwayland-dev libxkbcommon-dev ca-certificates

RUN ldconfig

COPY . .

WORKDIR server 

RUN go build -o . main.go

EXPOSE 8000/tcp
EXPOSE 8000/udp

# Run the application
CMD ["./main"]
