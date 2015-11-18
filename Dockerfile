# Docker image for the Drone rsync plugin
#
#     CGO_ENABLED=0 go build -a -tags netgo
#     docker build --rm=true -t plugins/drone-rsync .

FROM alpine:3.2
RUN apk add -U ca-certificates rsync && rm -rf /var/cache/apk/*
ADD drone-rsync /bin/
ENTRYPOINT ["/bin/drone-rsync"]