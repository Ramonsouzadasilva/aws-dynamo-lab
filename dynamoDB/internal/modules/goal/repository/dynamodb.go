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
	"github.com/ramon/goals-tasks-api/internal/modules/goal/entity"
)

type DynamoGoalRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoGoalRepository(client *dynamodb.Client, tableName string) *DynamoGoalRepository {
	return &DynamoGoalRepository{
		client:    client,
		tableName: tableName,
	}
}

type GoalRow struct {
	PK          string    `dynamodbav:"PK"`
	SK          string    `dynamodbav:"SK"`
	Type        string    `dynamodbav:"Type"`
	ID          string    `dynamodbav:"id"`
	UserID      string    `dynamodbav:"user_id"`
	Title       string    `dynamodbav:"title"`
	Description string    `dynamodbav:"description"`
	StartDate   time.Time `dynamodbav:"start_date"`
	EndDate     time.Time `dynamodbav:"end_date"`
	IsActive    bool      `dynamodbav:"is_active"`
	CreatedAt   time.Time `dynamodbav:"created_at"`
	UpdatedAt   time.Time `dynamodbav:"updated_at"`

	GSI3_PK string `dynamodbav:"GSI3_PK,omitempty"`
	GSI3_SK string `dynamodbav:"GSI3_SK,omitempty"`
}

func toRow(goal *entity.Goal) GoalRow {
	pk := fmt.Sprintf("USER#%s", goal.UserID)
	sk := fmt.Sprintf("GOAL#%s", goal.ID)

	var gsi3PK, gsi3SK string
	if goal.IsActive {
		gsi3PK = pk
		gsi3SK = "ACTIVE#true"
	} else {
		gsi3PK = pk
		gsi3SK = "ACTIVE#false"
	}

	return GoalRow{
		PK:          pk,
		SK:          sk,
		Type:        "GOAL",
		ID:          goal.ID,
		UserID:      goal.UserID,
		Title:       goal.Title,
		Description: goal.Description,
		StartDate:   goal.StartDate,
		EndDate:     goal.EndDate,
		IsActive:    goal.IsActive,
		CreatedAt:   goal.CreatedAt,
		UpdatedAt:   goal.UpdatedAt,
		GSI3_PK:     gsi3PK,
		GSI3_SK:     gsi3SK,
	}
}

func toDomain(row *GoalRow) *entity.Goal {
	return &entity.Goal{
		ID:          row.ID,
		UserID:      row.UserID,
		Title:       row.Title,
		Description: row.Description,
		StartDate:   row.StartDate,
		EndDate:     row.EndDate,
		IsActive:    row.IsActive,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func (r *DynamoGoalRepository) Create(ctx context.Context, goal *entity.Goal) error {
	row := toRow(goal)
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

func (r *DynamoGoalRepository) Update(ctx context.Context, goal *entity.Goal) error {
	return r.Create(ctx, goal) // PutItem replaces the item
}

func (r *DynamoGoalRepository) Delete(ctx context.Context, userID, goalID string) error {
	// 1. Query all tasks associated with this Goal
	// Tasks for a goal have PK = GOAL#{goalID} and SK = TASK#{taskID}
	taskPK := fmt.Sprintf("GOAL#%s", goalID)

	queryOut, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: taskPK},
		},
	})
	if err == nil && len(queryOut.Items) > 0 {
		// Cascade delete tasks
		for _, item := range queryOut.Items {
			var taskKey struct {
				PK string `dynamodbav:"PK"`
				SK string `dynamodbav:"SK"`
			}
			if err := attributevalue.UnmarshalMap(item, &taskKey); err == nil {
				_, _ = r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
					TableName: aws.String(r.tableName),
					Key: map[string]types.AttributeValue{
						"PK": &types.AttributeValueMemberS{Value: taskKey.PK},
						"SK": &types.AttributeValueMemberS{Value: taskKey.SK},
					},
				})
			}
		}
	}

	// 2. Delete the Goal itself
	goalPK := fmt.Sprintf("USER#%s", userID)
	goalSK := fmt.Sprintf("GOAL#%s", goalID)

	_, err = r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: goalPK},
			"SK": &types.AttributeValueMemberS{Value: goalSK},
		},
	})
	return err
}

func (r *DynamoGoalRepository) GetByID(ctx context.Context, userID, goalID string) (*entity.Goal, error) {
	pk := fmt.Sprintf("USER#%s", userID)
	sk := fmt.Sprintf("GOAL#%s", goalID)

	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk},
			"SK": &types.AttributeValueMemberS{Value: sk},
		},
	})
	if err != nil {
		return nil, err
	}

	if out.Item == nil {
		return nil, appErrors.ErrGoalNotFound
	}

	var row GoalRow
	if err := attributevalue.UnmarshalMap(out.Item, &row); err != nil {
		return nil, err
	}

	return toDomain(&row), nil
}

func (r *DynamoGoalRepository) ListActive(ctx context.Context, userID string) ([]*entity.Goal, error) {
	pk := fmt.Sprintf("USER#%s", userID)

	// Query GSI3 (Active Goals Index)
	out, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("GSI3"),
		KeyConditionExpression: aws.String("GSI3_PK = :pk AND GSI3_SK = :sk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: pk},
			":sk": &types.AttributeValueMemberS{Value: "ACTIVE#true"},
		},
	})
	if err != nil {
		return nil, err
	}

	goals := make([]*entity.Goal, 0, len(out.Items))
	for _, item := range out.Items {
		var row GoalRow
		if err := attributevalue.UnmarshalMap(item, &row); err == nil {
			goals = append(goals, toDomain(&row))
		}
	}

	return goals, nil
}

func (r *DynamoGoalRepository) ListAll(ctx context.Context, userID string) ([]*entity.Goal, error) {
	pk := fmt.Sprintf("USER#%s", userID)
	skPrefix := "GOAL#"

	out, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :skPrefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":       &types.AttributeValueMemberS{Value: pk},
			":skPrefix": &types.AttributeValueMemberS{Value: skPrefix},
		},
	})
	if err != nil {
		return nil, err
	}

	goals := make([]*entity.Goal, 0, len(out.Items))
	for _, item := range out.Items {
		var row GoalRow
		if err := attributevalue.UnmarshalMap(item, &row); err == nil {
			goals = append(goals, toDomain(&row))
		}
	}

	return goals, nil
}
