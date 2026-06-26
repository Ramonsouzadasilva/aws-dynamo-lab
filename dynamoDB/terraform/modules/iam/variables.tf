variable "role_name" {
  description = "Name of the IAM role to create"
  type        = string
  default     = "goals-api-execution-role"
}

variable "tags" {
  description = "Tags to assign to resources"
  type        = map(string)
  default     = {}
}
