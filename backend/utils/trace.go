package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateTraceID 生成追溯码（纳秒时间戳 + 随机熵，避免同一秒内重复违反 trace_id 唯一索引）
func GenerateTraceID(productID, batchNumber, privateKeySignature string) string {
	var entropy [16]byte
	_, _ = rand.Read(entropy[:])
	data := fmt.Sprintf("%s%s%d%x%s", productID, batchNumber, time.Now().UnixNano(), entropy, privateKeySignature)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// CalculateHash 计算数据哈希
func CalculateHash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// GenerateProductHash 生成商品信息哈希摘要
func GenerateProductHash(productID, batchNumber, origin, productionTime string) string {
	data := fmt.Sprintf("%s%s%s%s", productID, batchNumber, origin, productionTime)
	return CalculateHash(data)
}

// GenerateTemperatureRecordHash 温控记录内容哈希（入库后含主键，防篡改校验用）
func GenerateTemperatureRecordHash(id uint, traceID string, temperature, humidity float64, location string, recordTime time.Time, warehouseID uint, warehouseName string, isAbnormal bool) string {
	data := fmt.Sprintf("temp|%d|%s|%.4f|%.4f|%s|%s|%d|%s|%v",
		id, traceID, temperature, humidity, location, recordTime.UTC().Format(time.RFC3339Nano), warehouseID, warehouseName, isAbnormal)
	return CalculateHash(data)
}

// GenerateTransportNodeHash 物流节点内容哈希（入库后含主键；离开时间零值用 "-"）
func GenerateTransportNodeHash(id uint, traceID, nodeName, location string, arrivalTime, departureTime time.Time, logisticsID uint, logisticsName, status string) string {
	dep := "-"
	if !departureTime.IsZero() {
		dep = departureTime.UTC().Format(time.RFC3339Nano)
	}
	data := fmt.Sprintf("logistics|%d|%s|%s|%s|%s|%s|%d|%s|%s",
		id, traceID, nodeName, location, arrivalTime.UTC().Format(time.RFC3339Nano), dep, logisticsID, logisticsName, status)
	return CalculateHash(data)
}
