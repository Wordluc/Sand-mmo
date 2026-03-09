FROM golang:1.26

WORKDIR /code

RUN apt-get update && apt-get install -y gcc make build-essential
RUN apt-get install -y libgl1-mesa-dev libxi-dev libxcursor-dev libxrandr-dev libxinerama-dev libwayland-dev libxkbcommon-dev ca-certificates

RUN ldconfig

COPY . .

RUN go build -o server/ server/main.go

# Tell ldconfig we have new libraries to look at

EXPOSE 8000/tcp
EXPOSE 8000/udp

# Run the application
CMD ["server/main"]



