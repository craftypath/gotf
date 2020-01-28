terraform {
  backend "local" {}
}

variable "foo" {}

variable "bar" {}

variable "mapvar" {}

output "foo" {
  value = var.foo
}

output "mapvar" {
  value   = var.mapvar
}

resource "null_resource" "echo" {
  provisioner "local-exec" {
    command = "echo foo=${var.foo}"
  }
  provisioner "local-exec" {
    command = "echo baz=${var.bar}"
  }
}
