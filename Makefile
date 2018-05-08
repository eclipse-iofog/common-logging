## Build binary
bin:
	go build  --ldflags '-extldflags "-static"' -x -o logging .
## Build with version number for test purposes
build:
	sudo docker build -t iofog/common-logging:$(TAG) .
## Push with version number for test purposes
push:build
    sudo docker push iofog/common-logging:$(TAG)
## Tag latest to verified version number
latest:
    sudo docker tag iofog/common-logging:$(TAG) iofog/common-logging
## Push latest
push-latest:latest
    sudo docker push iofog/common-logging

## Same cmds for arm
build-arm:
    sudo docker build -t iofog/common-logging-arm:$(TAG) -f Dockerfile-arm .
push-arm:build-arm
    sudo docker push iofog/common-logging-arm:$(TAG)
latest-arm:
    sudo docker tag iofog/common-logging-arm:$(TAG) iofog/common-logging-arm
push-latest-arm:latest-arm
    sudo docker push iofog/common-logging-arm