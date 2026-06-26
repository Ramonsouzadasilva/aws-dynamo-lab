module "dynamodb" {
  source     = "../../modules/dynamodb"
  table_name = "goals_tasks_app"
  tags = {
    Environment = "local"
  }
}

module "sqs_sns" {
  source     = "../../modules/sqs_sns"
  topic_name = "goal-notifications"
  queue_name = "goal-notifications-queue"
  tags = {
    Environment = "local"
  }
}

module "iam" {
  source    = "../../modules/iam"
  role_name = "goals-api-execution-role"
  tags = {
    Environment = "local"
  }
}

module "cloudwatch" {
  source          = "../../modules/cloudwatch"
  log_group_name  = "/aws/api/goals-tasks-api"
  tags = {
    Environment = "local"
  }
}
