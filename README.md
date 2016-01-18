# drone-rsync

[![Build Status](http://beta.drone.io/api/badges/drone-plugins/drone-rsync/status.svg)](http://beta.drone.io/drone-plugins/drone-rsync)
[![](https://badge.imagelayers.io/plugins/drone-rsync:latest.svg)](https://imagelayers.io/?images=plugins/drone-rsync:latest 'Get your own badge on imagelayers.io')

Drone plugin for rsyncing files to a remote server

## Usage

Sync files with a remote server using the host machines rsync install:

```
./drone-rsync <<EOF
{
    "workspace": {
        "root": "/drone/src",
        "path": "/drone/src/github.com/drone/drone",
    },
    "vargs": {
        "user": "root",
        "host":" test.drone.io",
        "port": 22,
        "source": "dist/",
        "target": "/path/on/server",
        "delete": false,
        "recursive": false,
        "include": [],
        "exclude": [],
        "filter": []
    }
}
EOF
```
