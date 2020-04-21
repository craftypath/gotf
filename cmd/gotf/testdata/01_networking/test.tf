terraform {
  backend "local" {}
}

variable "foo" {}

variable "bar" {}

variable "mapvar" {}

variable "myvar" {}

variable "globalVar" {}

variable "envSpecificVar" {}

output "bar" {
  value = var.bar
}

output "foo" {
  value = var.foo
}

output "mapvar" {
  value   = var.mapvar
}

output "globalVar" {
  value   = var.globalVar
}

output "envSpecificVar" {
  value   = var.envSpecificVar
}

output "myvar" {
  value   = var.myvar
}

resource "null_resource" "echo" {
  provisioner "local-exec" {
    command = "echo foo=${var.foo}"
  }
  provisioner "local-exec" {
    command = "echo baz=${var.bar}"
  }
}
