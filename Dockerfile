FROM plugins/base:amd64

LABEL maintainer="Drone.IO Community <drone-dev@googlegroups.com>" \
  org.label-schema.name="Drone Rsync" \
  org.label-schema.vendor="Drone.IO Community" \
  org.label-schema.schema-version="1.0"

RUN apk add --no-cache openssh-client rsync

ADD release/linux/amd64/drone-rsync /bin/
ENTRYPOINT ["/bin/drone-rsync"]
