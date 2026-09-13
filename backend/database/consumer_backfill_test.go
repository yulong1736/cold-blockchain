package database

import (
	"cold-chain-trace/backend/models"
	"testing"
)

func TestBuildProductConsumerBackfillPairs(t *testing.T) {
	users := []models.User{
		{ID: 1, Username: "alice", Role: models.RoleConsumer},
		{ID: 2, Username: "bob", Role: models.RoleProducer},
		{ID: 3, Username: "cindy", Role: models.RoleConsumer},
	}
	products := []models.Product{
		{ID: 10, Consignee: "alice", ConsumerID: 0},
		{ID: 11, Consignee: "bob", ConsumerID: 0},   // 非消费者，不应回填
		{ID: 12, Consignee: "cindy", ConsumerID: 9}, // 已有 consumer_id，不应回填
		{ID: 13, Consignee: "nobody", ConsumerID: 0},
	}

	pairs := buildProductConsumerBackfillPairs(products, users)
	if len(pairs) != 1 {
		t.Fatalf("want 1 pair, got %d", len(pairs))
	}
	if pairs[0].productID != 10 || pairs[0].userID != 1 {
		t.Fatalf("unexpected pair: %+v", pairs[0])
	}
}
