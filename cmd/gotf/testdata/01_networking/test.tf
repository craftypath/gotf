terraform {
  backend "local" {}
}

variable "foo" {}

variable "bar" {}

variable "mapvar" {}

variable "myvar" {}

variable "global_var" {}

variable "env_specific_var" {}

variable "var_from_env_file" {}

output "bar" {
  value = var.bar
}

output "foo" {
  value = var.foo
}

output "mapvar" {
  value = var.mapvar
}

output "global_var" {
  value = var.global_var
}

output "env_specific_var" {
  value = var.env_specific_var
}

output "var_from_env_file" {
  value = var.var_from_env_file
}

output "myvar" {
  value = var.myvar
}

resource "null_resource" "echo" {
  provisioner "local-exec" {
    command = "echo foo=${var.foo}"
  }
  provisioner "local-exec" {
    command = "echo baz=${var.bar}"
  }
}
