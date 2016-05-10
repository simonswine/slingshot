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

## Build exmaple cluster

```
./slingshot cluster create \
  -I "simonswine/slingshot-ip-vagrant-coreos" \
  -C "simonswine/slingshot-cp-ansible-k8s-contrib" \
  clusterA
```

## Override ansible vars

```
./slingshot cluster create \
  -f - \
  -I simonswine/slingshot-ip-vagrant-coreos:0.0.2 \
  -C simonswine/slingshot-cp-ansible-k8s-contrib:0.0.2 \
  clusterB <<EOF
custom:
  ansible_vars_pre: |
    kube_version: 1.3.0-alpha-petset1
    kube_download_url_base: "https://s3-eu-west-1.amazonaws.com/jetstack.io-kubernetes-builds/release/v{{ kube_version }}/bin/linux/amd64"
EOF
```
