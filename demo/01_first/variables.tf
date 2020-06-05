variable "global_message" {
  description = "A global message to print to the console"
  type        = string
}

variable "use_special_chars" {
  description = "Specifies whether special characters should be used in random string"
  type        = bool
}

variable "module_specific_messages" {
  description = "Messages to print to the console"
  type = list(string)
}
