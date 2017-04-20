bin:
	GOARCH=${ARCH} go build  --ldflags '-extldflags "-static"' -x -o logging .
build:
	sudo docker build -t iotracks/catalog:logging-go-0.1$(TAG) .
push:build
	sudo docker push iotracks/catalog:logging-go-0.1$(TAG)
