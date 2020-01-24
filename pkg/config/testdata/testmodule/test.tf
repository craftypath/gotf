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
