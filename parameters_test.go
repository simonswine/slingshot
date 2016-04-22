package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParametersInventory(t *testing.T) {

	yamlContent := `inventory:
  - name: k8s-masters-1
    hostname: masters-1.k8s.den
    roles:
      - masters
    privateIP: 192.168.51.51
    publicIP: 192.168.51.51
  - name: k8s-workers-1
    roles:
      - workers
    privateIP: 192.168.51.52
    publicIP: 192.168.51.52`

	p := &Parameters{}
	err := p.Parse(yamlContent)
	valErrs := p.validateInventory()

	assert.Nil(t, err, "Unexpected error during parsing")

	// ensure no validation erros
	assert.Equal(t, []error(nil), valErrs)

	// ensure all vars are parse
	assert.Equal(t, 2, len(p.Inventory))
}

func TestParametersMachines(t *testing.T) {
	yamlContent := `cluster:
  machines:
    master:
      count: 2
      cores: 2
      instanceType: m3.medium
      memory: 1024
      roles:
        - masters
    worker:
      count: 1
      cores: 2
      instanceType: t2.large
      memory: 2048
      roles:
        - workers`

	p := &ParametersGeneral{}
	p.Defaults()
	err := p.Parse(yamlContent)
	valErrs := p.Cluster.ValidateMachines()

	assert.Nil(t, err, "Unexpected error during parsing")

	// ensure no validation erros
	assert.Equal(t, []error(nil), valErrs)

	assert.Equal(t, len(p.Cluster.Machines), 2)

	machinesProcessed := 0

	for key, machine := range p.Cluster.Machines {
		if key == "master" {
			assert.Equal(t, []string{"masters"}, *machine.Roles)
			machinesProcessed++
		}
		if key == "worker" {
			assert.Equal(t, []string{"workers"}, *machine.Roles)
			machinesProcessed++
		}
	}

	assert.Equal(t, 2, machinesProcessed, "Not all expected machines were found")
}

func TestParametersCluster(t *testing.T) {
	yamlContent := `cluster:
  kubernetes:
    interface: eth1
    masterApiPort: 8443
    serviceNetwork: 10.240.0.0/16
    dns:
      domainName: cluster.swine.de
      replicas: 3
    networking: noflannel
    flannel:
      subnet: 172.17.0.0
      prefix: 15
      hostPrefix: 23
    addons:
      clusterLogging: true
      clusterMonitoring: true
      kubeUI: true
      kubeDash: true`

	p := &ParametersGeneral{}
	p.Defaults()
	err := p.Parse(yamlContent)
	valErrs := p.Cluster.Validate()

	assert.Nil(t, err, "Unexpected error during parsing")

	// ensure no validation erros
	assert.Equal(t, []error(nil), valErrs)

	assert.Equal(t, "eth1", *p.Cluster.Kubernetes.Interface)
	assert.Equal(t, 8443, p.Cluster.Kubernetes.MasterApiPort)
	assert.Equal(t, "10.240.0.0/16", p.Cluster.Kubernetes.ServiceNetwork)

	assert.Equal(t, "cluster.swine.de", p.Cluster.Kubernetes.Dns.DomainName)
	assert.Equal(t, 3, p.Cluster.Kubernetes.Dns.Replicas)

	assert.Equal(t, "noflannel", p.Cluster.Kubernetes.Networking)

	assert.Equal(t, "172.17.0.0", p.Cluster.Kubernetes.Flannel.Subnet)
	assert.Equal(t, 15, p.Cluster.Kubernetes.Flannel.Prefix)
	assert.Equal(t, 23, p.Cluster.Kubernetes.Flannel.HostPrefix)

	assert.True(t, p.Cluster.Kubernetes.Addons.ClusterLogging)
	assert.True(t, p.Cluster.Kubernetes.Addons.ClusterMonitoring)
	assert.True(t, p.Cluster.Kubernetes.Addons.KubeUI)
	assert.True(t, p.Cluster.Kubernetes.Addons.KubeDash)

}

func TestParametersAuthentication(t *testing.T) {

	yamlContent := `authentication:
  ssh:
      user: core
      # this is the vagrant insecure key
      privateKey: |
        -----BEGIN RSA PRIVATE KEY-----
        MIIEogIBAAKCAQEA6NF8iallvQVp22WDkTkyrtvp9eWW6A8YVr+kz4TjGYe7gHzI
        w+niNltGEFHzD8+v1I2YJ6oXevct1YeS0o9HZyN1Q9qgCgzUFtdOKLv6IedplqoP
        kcmF0aYet2PkEDo3MlTBckFXPITAMzF8dJSIFo9D8HfdOV0IAdx4O7PtixWKn5y2
        hMNG0zQPyUecp4pzC6kivAIhyfHilFR61RGL+GPXQ2MWZWFYbAGjyiYJnAmCP3NO
        Td0jMZEnDkbUvxhMmBYSdETk1rRgm+R4LOzFUGaHqHDLKLX+FIPKcF96hrucXzcW
        yLbIbEgE98OHlnVYCzRdK8jlqm8tehUc9c9WhQIBIwKCAQEA4iqWPJXtzZA68mKd
        ELs4jJsdyky+ewdZeNds5tjcnHU5zUYE25K+ffJED9qUWICcLZDc81TGWjHyAqD1
        Bw7XpgUwFgeUJwUlzQurAv+/ySnxiwuaGJfhFM1CaQHzfXphgVml+fZUvnJUTvzf
        TK2Lg6EdbUE9TarUlBf/xPfuEhMSlIE5keb/Zz3/LUlRg8yDqz5w+QWVJ4utnKnK
        iqwZN0mwpwU7YSyJhlT4YV1F3n4YjLswM5wJs2oqm0jssQu/BT0tyEXNDYBLEF4A
        sClaWuSJ2kjq7KhrrYXzagqhnSei9ODYFShJu8UWVec3Ihb5ZXlzO6vdNQ1J9Xsf
        4m+2ywKBgQD6qFxx/Rv9CNN96l/4rb14HKirC2o/orApiHmHDsURs5rUKDx0f9iP
        cXN7S1uePXuJRK/5hsubaOCx3Owd2u9gD6Oq0CsMkE4CUSiJcYrMANtx54cGH7Rk
        EjFZxK8xAv1ldELEyxrFqkbE4BKd8QOt414qjvTGyAK+OLD3M2QdCQKBgQDtx8pN
        CAxR7yhHbIWT1AH66+XWN8bXq7l3RO/ukeaci98JfkbkxURZhtxV/HHuvUhnPLdX
        3TwygPBYZFNo4pzVEhzWoTtnEtrFueKxyc3+LjZpuo+mBlQ6ORtfgkr9gBVphXZG
        YEzkCD3lVdl8L4cw9BVpKrJCs1c5taGjDgdInQKBgHm/fVvv96bJxc9x1tffXAcj
        3OVdUN0UgXNCSaf/3A/phbeBQe9xS+3mpc4r6qvx+iy69mNBeNZ0xOitIjpjBo2+
        dBEjSBwLk5q5tJqHmy/jKMJL4n9ROlx93XS+njxgibTvU6Fp9w+NOFD/HvxB3Tcz
        6+jJF85D5BNAG3DBMKBjAoGBAOAxZvgsKN+JuENXsST7F89Tck2iTcQIT8g5rwWC
        P9Vt74yboe2kDT531w8+egz7nAmRBKNM751U/95P9t88EDacDI/Z2OwnuFQHCPDF
        llYOUI+SpLJ6/vURRbHSnnn8a/XG+nzedGH5JGqEJNQsz+xT2axM0/W/CRknmGaJ
        kda/AoGANWrLCz708y7VYgAtW2Uf1DPOIYMdvo6fxIB5i9ZfISgcJ/bbCUkFrhoH
        +vq/5CIWxCPp0f85R4qxxQ5ihxJ0YDQT9Jpx4TMss4PSavPaBH3RXow5Ohe+bYoQ
        NE5OgEXk2wVfZczCZpigBKbKZHNYcelXtTt/nP3rsCuGcM4h53s=
        -----END RSA PRIVATE KEY-----
`

	p := &ParametersGeneral{}
	p.Defaults()
	err := p.Parse(yamlContent)
	valErrs := p.Authentication.Validate()

	assert.Nil(t, err, "Unexpected error during parsing")

	// ensure no validation erros
	assert.Equal(t, []error(nil), valErrs)
}
