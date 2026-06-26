output "topic_arn" {
  description = "The ARN of the SNS topic"
  value       = aws_sns_topic.main.arn
}

output "queue_arn" {
  description = "The ARN of the SQS queue"
  value       = aws_sqs_queue.main.arn
}

output "queue_url" {
  description = "The URL of the SQS queue"
  value       = aws_sqs_queue.main.id
}
