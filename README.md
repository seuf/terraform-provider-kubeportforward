# terraform-provider-kubeportforward

This provider enables Kube port forwarding in Terraform.

*This provider does not support Terraform v0.12 yet. There were some changes made that makes the upgrade non-trivial.*

This provider is inspired by :

* [stefansundin/terraform-provider-ssh](https://github.com/stefansundin/terraform-provider-ssh).
* [txn2/kubefwd](https://github.com/txn2/kubefwd)

## Example

```terraform
provider "kubeportforward" {
  version = "~> 0.0"
}

data "kubeportforward" "grafana" {
  kube_config = "/home/seuf/.kube/config"
  context     = "k3s_default"
  namespace   = "monitoring"
  service     = "grafana"
  local_port  = "3000"
  remote_port = "3000"
}

provider "grafana" {
  url     = "http://localhost:3000/"
  auth    = "1234567890abcdefghijklmop"
  version = "~> 1.4"
}

output "service" {
  value      = "${data.kubeportforward.grafana.service}"
  depends_on = ["data.kubeportforward.grafana"]
}

resource "grafana_data_source" "influxdb" {
  type          = "influxdb"
  name          = "test_influxdb"
  url           = "http://influxdb.example2.net:8086/"
  username      = "foo"
  password      = "bar"
  database_name = "mydb"
  depends_on    = ["data.kubeportforward.grafana"]
}
```

Each resource which need the kube port forward need to explicitely add a dependency to the kubeportforward provider.

## Installation

On Linux:

```shell
mkdir -p terraform.d/plugins/linux_amd64
wget https://github.com/seuf/terraform-provider-kubeportforward/releases/download/v0.0.1/terraform-provider-kubeportforward_v0.0.1_linux_amd64.zip
unzip terraform-provider-kubeportforward_v0.0.1_linux_amd64.zip -d terraform.d/plugins/linux_amd64
rm terraform-provider-kubeportforward_v0.0.1_linux_amd64.zip
terraform init
```

On Mac:

```shell
mkdir -p terraform.d/plugins/darwin_amd64
wget https://github.com/seuf/terraform-provider-kubeportforward/releases/download/v0.0.1/terraform-provider-kubeportforward_v0.0.1_darwin_amd64.zip
unzip terraform-provider-kubeportforward_v0.0.1_darwin_amd64.zip -d terraform.d/plugins/darwin_amd64
rm terraform-provider-kubeportforward_v0.0.1_darwin_amd64.zip
terraform init
```

## Build

Refering to [client-go install](https://github.com/kubernetes/client-go/blob/master/INSTALL.md). Fix the following packages version :

```shell
go get k8s.io/client-go@v11.0.0
go get k8s.io/api@kubernetes-1.14.0
go get k8s.io/apimachinery@kubernetes-1.14.0
```

Then you can build the binary :

```shell
make linux
```

## TODO

* Note that the Windows binary is completely untested!
