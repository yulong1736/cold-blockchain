package models

import (
	"encoding/json"
	"testing"
	"time"
)

// 保证温控 / 物流列表 API 序列化包含存证字段，供前端「区块链状态」列使用。
func TestTemperatureRecordJSONIncludesBlockchainFields(t *testing.T) {
	r := TemperatureRecord{
		ID:                 1,
		TraceID:            "trace-x",
		Temperature:        2.5,
		BlockchainHash:     "deadbeef",
		BlockchainTxHash:   "0xabc123",
		RecordTime:         time.Date(2026, 4, 7, 12, 0, 0, 0, time.UTC),
		WarehouseID:        10,
		WarehouseName:      "wh1",
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m["blockchain_hash"] != "deadbeef" {
		t.Fatalf("blockchain_hash: got %v", m["blockchain_hash"])
	}
	if m["blockchain_tx_hash"] != "0xabc123" {
		t.Fatalf("blockchain_tx_hash: got %v", m["blockchain_tx_hash"])
	}
}

func TestTransportNodeJSONIncludesBlockchainFields(t *testing.T) {
	n := TransportNode{
		ID:                 2,
		TraceID:            "trace-y",
		NodeName:           "港A",
		BlockchainHash:     "cafef00d",
		BlockchainTxHash:   "0xdef456",
		ArrivalTime:        time.Date(2026, 4, 7, 15, 0, 0, 0, time.UTC),
		LogisticsID:        20,
		LogisticsName:      "物流1",
		Status:             "in_transit",
	}
	b, err := json.Marshal(n)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m["blockchain_hash"] != "cafef00d" {
		t.Fatalf("blockchain_hash: got %v", m["blockchain_hash"])
	}
	if m["blockchain_tx_hash"] != "0xdef456" {
		t.Fatalf("blockchain_tx_hash: got %v", m["blockchain_tx_hash"])
	}
}
