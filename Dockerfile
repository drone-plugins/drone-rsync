# Docker image for the rsync plugin
#
#     docker build --rm=true -t plugins/drone-rsync .

FROM library/golang:1.4

ADD . /go/src/github.com/drone-plugins/drone-rsync/


RUN apt-get update && apt-get -y install rsync


RUN go get github.com/drone-plugins/drone-rsync/... && \
    go install github.com/drone-plugins/drone-rsync


RUN apt-get update -qq           && \
	apt-get -y install rsync     && \
	rm -rf /var/lib/apt/lists/*

ENTRYPOINT ["/go/bin/drone-rsync"]
