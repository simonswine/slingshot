## Build

The easiest way of producing a reproducible build is using the default docker image to build slingshot (of course :laughing:)

The build creates:

* Linux and MacOS X 64bit binaries in _build/
* Coverage report in _test/

The only requirement is a running docker installation for the user that runs the `make` command.

```
make docker
```

## Usage

```
./slingshot cluster create \
  -I "simonswine/slingshot-ip-vagrant-coreos" \
  -C "simonswine/slingshot-cp-ansible-k8s-contrib"
```
