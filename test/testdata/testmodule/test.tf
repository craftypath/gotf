terraform {
  backend "local" {}
}

variable "foo" {}

variable "baz" {}

output "foo" {
  value = var.foo
}

output "baz" {
  value = var.baz
}

resource "null_resource" "echo" {
  provisioner "local-exec" {
    command = "echo foo=${var.foo}"
  }
  provisioner "local-exec" {
    command = "echo baz=${var.baz}"
  }
}
