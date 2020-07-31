This is release ${DRONE_TAG} of Grizzly (`grr`). Check out the [CHANGELOG](CHANGELOG.md) for detailed release notes.
## Install instructions

#### Binary:
```bash
# download the binary (adapt os and arch as needed)
$ curl -fSL -o "/usr/local/bin/grr" "https://github.com/grafana/grizzly/releases/download/${DRONE_TAG}/grr-linux-amd64"

# make it executable
$ chmod a+x "/usr/local/bin/grr"

# have fun :)
$ grr --help
```
