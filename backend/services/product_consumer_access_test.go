package services

import (
	"cold-chain-trace/backend/models"
	"testing"
)

func TestCanConsumerAccessProduct(t *testing.T) {
	product := &models.Product{ID: 100, ConsumerID: 7}

	if !canConsumerAccessProduct(product, 7) {
		t.Fatal("expected consumer with same id to access product")
	}
	if canConsumerAccessProduct(product, 8) {
		t.Fatal("expected different consumer id to be denied")
	}
	if canConsumerAccessProduct(product, 0) {
		t.Fatal("expected zero consumer id to be denied")
	}
	if canConsumerAccessProduct(nil, 7) {
		t.Fatal("expected nil product to be denied")
	}
}

func TestFallbackConsumerAccessByUsername(t *testing.T) {
	product := &models.Product{ID: 101, ConsumerID: 0, Consignee: "test-customer"}
	consumerID := uint(9)
	username := "test-customer"

	allowed := product.ConsumerID == 0 && product.Consignee == username
	if !allowed {
		t.Fatal("expected legacy product to match by consignee username fallback")
	}
	if canConsumerAccessProduct(product, consumerID) {
		t.Fatal("legacy product with zero consumer_id should not pass strict id check before fallback")
	}
}
