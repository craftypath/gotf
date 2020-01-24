terraform {
  backend "local" {}
}

variable "foo" {}

variable "bar" {}

variable "baz" {}

variable "mapvar" {
  type = map(object({
    value1 = string
    value2 = bool
  }))
}

output "foo" {
  value = var.foo
}

output "baz" {
  value = var.baz
}
