# Prerequisites

- Linux or Mac OS X (not tested yet)
- Docker installed
- User that runs slingshot needs to be member of the docker group


# Usage

## Build cluster

### Vagrant + CoreOS

```
slingshot cluster create \
  -I "jetstack/slingshot-ip-vagrant-coreos:canary" \
  -C "jetstack/slingshot-cp-ansible-k8s-coreos:canary" \
  cluster-vagrant
```

### AWS + CoreOS and override kubernetes version `1.3.0-beta.1`

```
$ slingshot cluster create \
  -f - \
  -I jetstack/slingshot-ip-terraform-aws-coreos:canary \
  -C jetstack/slingshot-cp-ansible-k8s-coreos:canary \
  cluster-aws <<EOF
custom:
  aws_access_key: #AWS_KEY#
  aws_region: eu-west-1
  aws_secret_key: #AWS_SECRET#
  aws_zones: eu-west-1a
general:
  cluster:
    kubernetes:
      version: 1.3.0-beta.1
EOF
```

## Use cluster

### Use kubectl

```
# List nodes
$ slingshot cluster ssh cluster-aws kubectl get nodes
NAME                                           STATUS                     AGE
ip-172-20-128-39.eu-west-1.compute.internal    Ready                      2h
ip-172-20-128-40.eu-west-1.compute.internal    Ready                      2h
ip-172-20-130-173.eu-west-1.compute.internal   Ready,SchedulingDisabled   2h


# List pods
$ slingshot cluster ssh cluster-aws kubectl get pods --all-namespaces
NAMESPACE     NAME                                                                   READY     STATUS    RESTARTS   AGE
kube-system   kube-apiserver-ip-172-20-130-173.eu-west-1.compute.internal            1/1       Running   0          2h
kube-system   kube-controller-manager-ip-172-20-130-173.eu-west-1.compute.internal   1/1       Running   0          2h
kube-system   kube-dns-125411827-47qh5                                               4/4       Running   0          2h
kube-system   kube-proxy-ip-172-20-128-39.eu-west-1.compute.internal                 1/1       Running   0          2h
kube-system   kube-proxy-ip-172-20-128-40.eu-west-1.compute.internal                 1/1       Running   0          2h
kube-system   kube-proxy-ip-172-20-130-173.eu-west-1.compute.internal                1/1       Running   0          2h
kube-system   kube-scheduler-ip-172-20-130-173.eu-west-1.compute.internal            1/1       Running   0          2h
kube-system   kubernetes-dashboard-3638963639-d06ar                                  1/1       Running   0          2h
```

### Connect to a node

```
$ slingshot cluster ssh cluster-aws master
Last login: Mon Jun 20 12:32:51 2016 from 172.20.3.122
CoreOS stable (1010.5.0)
core@ip-172-20-130-173 ~ $
```

### List clusters

```
$ slingshot cluster list
Cluster Name    Infrastructure Provider                                 Config Provider
cluster-aws     jetstack/slingshot-ip-terraform-aws-coreos:canary       jetstack/slingshot-cp-ansible-k8s-coreos:canary
```

### List nodes

```
$ slingshot cluster nodes cluster-aws
Nodes in cluster cluster-aws
Name                    Aliases Roles   Private IP      Public IP
i-e5e69269              master  master  172.20.130.173
i-14e79398              worker1 worker  172.20.128.40
i-17e7939b              worker2 worker  172.20.128.39
i-05b1fdf0e948828e2     bastion bastion 172.20.3.122    52.208.78.101
```

# Build binaries

The easiest way of producing a reproducible build is using the default docker image to build slingshot (of course :laughing:)

The build creates:

* Linux and MacOS X 64bit binaries iin `_build/`
* Coverage report in `_test/`

The only requirement is a running docker installation for the user that runs the `make` command.

```
make docker
```
