variable "table_name" {
  description = "The name of the DynamoDB table"
  type        = string
  default     = "goals_tasks_app"
}

variable "tags" {
  description = "Tags to assign to the resource"
  type        = map(string)
  default     = {}
}
