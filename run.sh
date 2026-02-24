echo "Updating server"

docker stop prova
docker rm prova
docker rmi prova
docker build . -t prova
docker run -P --name prova -d prova 

if [[ $1 == "c" ]] ; then
	echo "Run Client"
	go run ./client/main.go
fi
