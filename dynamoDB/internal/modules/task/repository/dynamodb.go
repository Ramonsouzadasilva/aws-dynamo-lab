package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	appErrors "github.com/ramon/goals-tasks-api/internal/errors"
	"github.com/ramon/goals-tasks-api/internal/modules/task/entity"
)

type DynamoTaskRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoTaskRepository(client *dynamodb.Client, tableName string) *DynamoTaskRepository {
	return &DynamoTaskRepository{
		client:    client,
		tableName: tableName,
	}
}

type TaskRow struct {
	PK          string            `dynamodbav:"PK"`
	SK          string            `dynamodbav:"SK"`
	Type        string            `dynamodbav:"Type"`
	ID          string            `dynamodbav:"id"`
	GoalID      string            `dynamodbav:"goal_id,omitempty"`
	UserID      string            `dynamodbav:"user_id"`
	Title       string            `dynamodbav:"title"`
	Description string            `dynamodbav:"description"`
	Status      entity.TaskStatus `dynamodbav:"status"`
	DueDate     time.Time         `dynamodbav:"due_date"`
	IsRecurring bool              `dynamodbav:"is_recurring"`
	CreatedAt   time.Time         `dynamodbav:"created_at"`
	UpdatedAt   time.Time         `dynamodbav:"updated_at"`

	// GSIs
	GSI1_PK string `dynamodbav:"GSI1_PK"`
	GSI1_SK string `dynamodbav:"GSI1_SK"`
	GSI2_PK string `dynamodbav:"GSI2_PK"`
	GSI2_SK string `dynamodbav:"GSI2_SK"`
}

func toTaskRow(task *entity.Task) TaskRow {
	var pk string
	if task.GoalID != "" {
		pk = fmt.Sprintf("GOAL#%s", task.GoalID)
	} else {
		pk = fmt.Sprintf("USER#%s", task.UserID)
	}
	sk := fmt.Sprintf("TASK#%s", task.ID)

	gsi1PK := fmt.Sprintf("USER#%s", task.UserID)
	gsi1SK := fmt.Sprintf("STATUS#%s#TASK#%s", task.Status, task.ID)

	gsi2PK := fmt.Sprintf("USER#%s", task.UserID)
	gsi2SK := fmt.Sprintf("DATE#%s#TASK#%s", task.DueDate.Format(time.RFC3339), task.ID)

	return TaskRow{
		PK:          pk,
		SK:          sk,
		Type:        "TASK",
		ID:          task.ID,
		GoalID:      task.GoalID,
		UserID:      task.UserID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		DueDate:     task.DueDate,
		IsRecurring: task.IsRecurring,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		GSI1_PK:     gsi1PK,
		GSI1_SK:     gsi1SK,
		GSI2_PK:     gsi2PK,
		GSI2_SK:     gsi2SK,
	}
}

func toTaskDomain(row *TaskRow) *entity.Task {
	return &entity.Task{
		ID:          row.ID,
		GoalID:      row.GoalID,
		UserID:      row.UserID,
		Title:       row.Title,
		Description: row.Description,
		Status:      row.Status,
		DueDate:     row.DueDate,
		IsRecurring: row.IsRecurring,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func (r *DynamoTaskRepository) Create(ctx context.Context, task *entity.Task) error {
	row := toTaskRow(task)
	av, err := attributevalue.MarshalMap(row)
	if err != nil {
		return err
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	})
	return err
}

func (r *DynamoTaskRepository) CreateMultiple(ctx context.Context, tasks []*entity.Task) error {
	if len(tasks) == 0 {
		return nil
	}

	writeRequests := make([]types.WriteRequest, 0, len(tasks))
	for _, task := range tasks {
		row := toTaskRow(task)
		av, err := attributevalue.MarshalMap(row)
		if err != nil {
			return err
		}
		writeRequests = append(writeRequests, types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: av,
			},
		})
	}

	// DynamoDB BatchWriteItem supports up to 25 items per batch
	for i := 0; i < len(writeRequests); i += 25 {
		end := i + 25
		if end > len(writeRequests) {
			end = len(writeRequests)
		}

		batch := writeRequests[i:end]
		_, err := r.client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				r.tableName: batch,
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *DynamoTaskRepository) Update(ctx context.Context, task *entity.Task) error {
	return r.Create(ctx, task) // PutItem replaces the item
}

func (r *DynamoTaskRepository) GetByID(ctx context.Context, userID, taskID string) (*entity.Task, error) {
	// Since we don't know the GoalID, we query GSI1 which is partitioned by UserID
	gsi1PK := fmt.Sprintf("USER#%s", userID)

	out, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("GSI1_PK = :pk"),
		FilterExpression:       aws.String("id = :id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: gsi1PK},
			":id": &types.AttributeValueMemberS{Value: taskID},
		},
	})
	if err != nil {
		return nil, err
	}

	if len(out.Items) == 0 {
		return nil, appErrors.ErrTaskNotFound
	}

	var row TaskRow
	if err := attributevalue.UnmarshalMap(out.Items[0], &row); err != nil {
		return nil, err
	}

	return toTaskDomain(&row), nil
}

func (r *DynamoTaskRepository) Delete(ctx context.Context, userID, taskID string) error {
	task, err := r.GetByID(ctx, userID, taskID)
	if err != nil {
		return err
	}

	var pk string
	if task.GoalID != "" {
		pk = fmt.Sprintf("GOAL#%s", task.GoalID)
	} else {
		pk = fmt.Sprintf("USER#%s", task.UserID)
	}
	sk := fmt.Sprintf("TASK#%s", task.ID)

	_, err = r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk},
			"SK": &types.AttributeValueMemberS{Value: sk},
		},
	})
	return err
}

func (r *DynamoTaskRepository) ListByStatus(ctx context.Context, userID string, status entity.TaskStatus) ([]*entity.Task, error) {
	gsi1PK := fmt.Sprintf("USER#%s", userID)
	gsi1SKPrefix := fmt.Sprintf("STATUS#%s#", status)

	out, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("GSI1_PK = :pk AND begins_with(GSI1_SK, :skPrefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":       &types.AttributeValueMemberS{Value: gsi1PK},
			":skPrefix": &types.AttributeValueMemberS{Value: gsi1SKPrefix},
		},
	})
	if err != nil {
		return nil, err
	}

	tasks := make([]*entity.Task, 0, len(out.Items))
	for _, item := range out.Items {
		var row TaskRow
		if err := attributevalue.UnmarshalMap(item, &row); err == nil {
			tasks = append(tasks, toTaskDomain(&row))
		}
	}

	return tasks, nil
}

func (r *DynamoTaskRepository) ListByPeriod(ctx context.Context, userID string, start, end time.Time) ([]*entity.Task, error) {
	gsi2PK := fmt.Sprintf("USER#%s", userID)
	startSK := fmt.Sprintf("DATE#%s", start.Format(time.RFC3339))
	endSK := fmt.Sprintf("DATE#%s", end.Format(time.RFC3339))

	out, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("GSI2"),
		KeyConditionExpression: aws.String("GSI2_PK = :pk AND GSI2_SK BETWEEN :startVal AND :endVal"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":       &types.AttributeValueMemberS{Value: gsi2PK},
			":startVal": &types.AttributeValueMemberS{Value: startSK},
			":endVal":   &types.AttributeValueMemberS{Value: endSK},
		},
	})
	if err != nil {
		return nil, err
	}

	tasks := make([]*entity.Task, 0, len(out.Items))
	for _, item := range out.Items {
		var row TaskRow
		if err := attributevalue.UnmarshalMap(item, &row); err == nil {
			tasks = append(tasks, toTaskDomain(&row))
		}
	}

	return tasks, nil
}

func (r *DynamoTaskRepository) ListAllCompleted(ctx context.Context, userID string) ([]*entity.Task, error) {
	return r.ListByStatus(ctx, userID, entity.StatusCompleted)
}

func (r *DynamoTaskRepository) ListAll(ctx context.Context, userID string) ([]*entity.Task, error) {
	gsi1PK := fmt.Sprintf("USER#%s", userID)

	out, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("GSI1_PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: gsi1PK},
		},
	})
	if err != nil {
		return nil, err
	}

	tasks := make([]*entity.Task, 0, len(out.Items))
	for _, item := range out.Items {
		var row TaskRow
		if err := attributevalue.UnmarshalMap(item, &row); err == nil {
			tasks = append(tasks, toTaskDomain(&row))
		}
	}

	return tasks, nil
}

func (r *DynamoTaskRepository) CountTasksByGoal(ctx context.Context, goalID string) (completed, total int, err error) {
	pk := fmt.Sprintf("GOAL#%s", goalID)
	skPrefix := "TASK#"

	out, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :skPrefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":       &types.AttributeValueMemberS{Value: pk},
			":skPrefix": &types.AttributeValueMemberS{Value: skPrefix},
		},
	})
	if err != nil {
		return 0, 0, err
	}

	total = len(out.Items)
	completed = 0
	for _, item := range out.Items {
		var row TaskRow
		if err := attributevalue.UnmarshalMap(item, &row); err == nil {
			if row.Status == entity.StatusCompleted {
				completed++
			}
		}
	}

	return completed, total, nil
}
