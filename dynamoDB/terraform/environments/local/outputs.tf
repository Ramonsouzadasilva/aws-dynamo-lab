output "dynamodb_table_name" {
  value = module.dynamodb.table_name
}

output "sns_topic_arn" {
  value = module.sqs_sns.topic_arn
}

output "sqs_queue_url" {
  value = module.sqs_sns.queue_url
}

output "iam_role_arn" {
  value = module.iam.role_arn
}

output "cloudwatch_log_group_name" {
  value = module.cloudwatch.log_group_name
}
