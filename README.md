# drone-rsync

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
