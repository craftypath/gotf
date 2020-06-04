resource "random_string" "bar" {
  length  = 12
  special = var.use_special_chars
}

resource "null_resource" "global_message" {
  provisioner "local-exec" {
    command = "echo '${var.global_message}'"
  }
}

resource "null_resource" "module_specific_message" {
  for_each = toset(var.module_specific_messages)

  provisioner "local-exec" {
    command = "echo '${each.value}'"
  }
}
