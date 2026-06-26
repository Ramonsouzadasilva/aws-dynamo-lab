variable "topic_name" {
  description = "The name of the SNS topic"
  type        = string
  default     = "goal-notifications"
}

variable "queue_name" {
  description = "The name of the SQS queue"
  type        = string
  default     = "goal-notifications-queue"
}

variable "tags" {
  description = "Tags to assign to resources"
  type        = map(string)
  default     = {}
}
