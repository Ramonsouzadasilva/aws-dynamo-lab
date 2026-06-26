resource "aws_sns_topic" "main" {
  name = var.topic_name
  tags = var.tags
}

resource "aws_sqs_queue" "main" {
  name = var.queue_name
  tags = var.tags
}

resource "aws_sns_topic_subscription" "main" {
  topic_arn = aws_sns_topic.main.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.main.arn
}

resource "aws_sqs_queue_policy" "main" {
  queue_url = aws_sqs_queue.main.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect    = "Allow"
        Principal = "*"
        Action    = "sqs:SendMessage"
        Resource  = aws_sqs_queue.main.arn
        Condition = {
          ArnEquals = {
            "aws:SourceArn" = aws_sns_topic.main.arn
          }
        }
      }
    ]
  })
}
