package infra

import (
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type HealthChecker struct {
	dbClient  *dynamodb.Client
	tableName string
}

func NewHealthChecker(dbClient *dynamodb.Client, tableName string) *HealthChecker {
	return &HealthChecker{
		dbClient:  dbClient,
		tableName: tableName,
	}
}

func (h *HealthChecker) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"UP"}`))
}

func (h *HealthChecker) ReadyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Verify DynamoDB connectivity
	_, err := h.dbClient.DescribeTable(r.Context(), &dynamodb.DescribeTableInput{
		TableName: aws.String(h.tableName),
	})
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"status":"DOWN", "reason":"Conexão com DynamoDB falhou"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"UP"}`))
}
