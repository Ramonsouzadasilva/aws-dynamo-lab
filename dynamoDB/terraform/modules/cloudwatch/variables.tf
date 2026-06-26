variable "log_group_name" {
  description = "Name of the CloudWatch Log Group for API logs"
  type        = string
  default     = "/aws/api/goals-tasks-api"
}

variable "retention_in_days" {
  description = "Number of days to retain log events"
  type        = number
  default     = 7
}

variable "tags" {
  description = "Tags to assign to resources"
  type        = map(string)
  default     = {}
}
