resource "random_string" "foo" {
  length = var.random_string_length
}

resource "null_resource" "global_message" {
  provisioner "local-exec" {
    command = "echo '${var.global_message}'"
  }
}
