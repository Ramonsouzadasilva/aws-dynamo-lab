package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	appErrors "github.com/ramon/goals-tasks-api/internal/errors"
	"github.com/ramon/goals-tasks-api/internal/modules/user/entity"
)

type DynamoUserRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoUserRepository(client *dynamodb.Client, tableName string) *DynamoUserRepository {
	return &DynamoUserRepository{
		client:    client,
		tableName: tableName,
	}
}

type UserRow struct {
	PK        string    `dynamodbav:"PK"`
	SK        string    `dynamodbav:"SK"`
	Type      string    `dynamodbav:"Type"`
	ID        string    `dynamodbav:"ID"`
	Email     string    `dynamodbav:"Email"`
	Password  string    `dynamodbav:"Password"`
	CreatedAt time.Time `dynamodbav:"CreatedAt"`
}

func (r *DynamoUserRepository) Create(ctx context.Context, user *entity.User) error {
	userPK := fmt.Sprintf("USER#%s", user.ID)
	emailPK := fmt.Sprintf("USER_EMAIL#%s", user.Email)

	userRow := UserRow{
		PK:        userPK,
		SK:        "METADATA",
		Type:      "USER",
		ID:        user.ID,
		Email:     user.Email,
		Password:  user.Password,
		CreatedAt: user.CreatedAt,
	}

	emailRow := UserRow{
		PK:        emailPK,
		SK:        "METADATA",
		Type:      "USER_EMAIL",
		ID:        user.ID,
		Email:     user.Email,
		Password:  "",
		CreatedAt: user.CreatedAt,
	}

	userAV, err := attributevalue.MarshalMap(userRow)
	if err != nil {
		return err
	}

	emailAV, err := attributevalue.MarshalMap(emailRow)
	if err != nil {
		return err
	}

	_, err = r.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Put: &types.Put{
					TableName: aws.String(r.tableName),
					Item:      userAV,
				},
			},
			{
				Put: &types.Put{
					TableName:           aws.String(r.tableName),
					Item:                emailAV,
					ConditionExpression: aws.String("attribute_not_exists(PK)"),
				},
			},
		},
	})

	if err != nil {
		var transactionCanceledErr *types.TransactionCanceledException
		if errors.As(err, &transactionCanceledErr) {
			for _, reason := range transactionCanceledErr.CancellationReasons {
				if reason.Code != nil && *reason.Code == "ConditionalCheckFailed" {
					return appErrors.ErrUserAlreadyExists
				}
			}
		}
		return err
	}

	return nil
}

func (r *DynamoUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	emailPK := fmt.Sprintf("USER_EMAIL#%s", email)

	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: emailPK},
			"SK": &types.AttributeValueMemberS{Value: "METADATA"},
		},
	})
	if err != nil {
		return nil, err
	}

	if out.Item == nil {
		return nil, appErrors.ErrUserNotFound
	}

	var emailRow UserRow
	if err := attributevalue.UnmarshalMap(out.Item, &emailRow); err != nil {
		return nil, err
	}

	return r.GetByID(ctx, emailRow.ID)
}

func (r *DynamoUserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	userPK := fmt.Sprintf("USER#%s", id)

	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: userPK},
			"SK": &types.AttributeValueMemberS{Value: "METADATA"},
		},
	})
	if err != nil {
		return nil, err
	}

	if out.Item == nil {
		return nil, appErrors.ErrUserNotFound
	}

	var row UserRow
	if err := attributevalue.UnmarshalMap(out.Item, &row); err != nil {
		return nil, err
	}

	return entity.RecreateExistingUser(row.ID, row.Email, row.Password, row.CreatedAt), nil
}
