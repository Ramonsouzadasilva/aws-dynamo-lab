output "log_group_arn" {
  description = "The ARN of the CloudWatch Log Group"
  value       = aws_cloudwatch_log_group.api_logs.arn
}

output "log_group_name" {
  description = "The Name of the CloudWatch Log Group"
  value       = aws_cloudwatch_log_group.api_logs.name
}
