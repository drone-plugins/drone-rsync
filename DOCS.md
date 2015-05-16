Use the rsync plugin to deploy files to a server using rsync over
ssh. The following parameters are used to configuration this plugin:

* **user** - connects as this user
* **host** - connects to this host address
* **port** - connects to this host port
* **source** - source path from which files are copied
* **target** - target path to which files are copied
* **delete** - delete extraneous files from the target dir
* **recursive** - recursively transfer all files

The following is a sample Docker configuration in your .drone.yml file:

```yaml
deploy:
  rsync:
    user: root
    host: 127.0.0.1
    port: 22
    source: copy/files/from
    target: send/files/to
    delete: false
    recursive: true
```