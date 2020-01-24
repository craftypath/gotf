terraform {
  backend "local" {}
}

variable "foo" {}

variable "baz" {}

variable "mapvar" {}

output "foo" {
  value = var.foo
}

output "baz" {
  value = var.baz
}

output "mapvar" {
  value   = var.mapvar
}

resource "null_resource" "echo" {
  provisioner "local-exec" {
    command = "echo foo=${var.foo}"
  }
  provisioner "local-exec" {
    command = "echo baz=${var.baz}"
  }
}
