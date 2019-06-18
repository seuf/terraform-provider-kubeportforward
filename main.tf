provider "kubeportforward" {
  version = ">= 0.0"
}

data "kubeportforward" "grafana" {
  kube_config = "/home/thierry/.kube/config"
  namespace   = "monitoring"
  service     = "grafana"
  local_port  = "3000"
  remote_port = "3000"
}

provider "grafana" {
  url        = "http://localhost:3000/"
  auth       = "1234567890abcdefghijklmop"
  version    = "~> 1.4"
  depends_on = ["data.kubeportforward.grafana"]
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
