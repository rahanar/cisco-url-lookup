run:
	go run main.go

build:
	go build -o main main.go

docker-build:
	docker build . -t url-lookup

docker-run:
	docker run -p 8000:8000 url-lookup:latest
