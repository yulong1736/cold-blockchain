package services

import (
	"cold-chain-trace/backend/blockchain"
	"cold-chain-trace/backend/database"
	"cold-chain-trace/backend/models"
	"strings"
	"time"
)

// getProducerTraceIDs 返回某生产商拥有的商品追溯码列表（未删除）
// 含 producer_id=当前用户 的商品，以及 producer_id=0 但由该用户创建的商品（按 product_histories 的 create 记录归属）
// 分步查询避免 PostgreSQL 下 SubQuery 占位符语法错误
func getProducerTraceIDs(producerID uint) ([]string, error) {
	var traceIDs []string
	// 1) 直接属于该生产商的商品
	if err := database.DB.Model(&models.Product{}).
		Where("is_deleted = ? AND producer_id = ?", false, producerID).
		Pluck("trace_id", &traceIDs).Error; err != nil {
		return nil, err
	}
	// 2) producer_id=0 且由该用户创建的商品（通过 product_histories 查 product_id 再取 trace_id）
	var legacyProductIDs []uint
	if err := database.DB.Table(models.ProductHistory{}.TableName()).
		Where("operation_type = ? AND operator_id = ?", "create", producerID).
		Pluck("product_id", &legacyProductIDs).Error; err != nil {
		return nil, err
	}
	if len(legacyProductIDs) == 0 {
		return traceIDs, nil
	}
	ph, args := inClauseUints(legacyProductIDs)
	args = append([]interface{}{false}, args...)
	var extraTraceIDs []string
	if err := database.DB.Model(&models.Product{}).
		Where("is_deleted = ? AND producer_id = 0 AND id IN ("+ph+")", args...).
		Pluck("trace_id", &extraTraceIDs).Error; err != nil {
		return nil, err
	}
	// 合并并去重（同一 trace_id 可能已在 traceIDs 中）
	seen := make(map[string]bool)
	for _, t := range traceIDs {
		seen[t] = true
	}
	for _, t := range extraTraceIDs {
		if !seen[t] {
			seen[t] = true
			traceIDs = append(traceIDs, t)
		}
	}
	return traceIDs, nil
}

// getLogisticsTraceIDs 返回 transport_company = companyName 的商品追溯码列表
func getLogisticsTraceIDs(companyName string) ([]string, error) {
	var traceIDs []string
	if companyName == "" {
		return traceIDs, nil
	}
	if err := database.DB.Model(&models.Product{}).
		Where("transport_company = ? AND is_deleted = ?", companyName, false).
		Pluck("trace_id", &traceIDs).Error; err != nil {
		return nil, err
	}
	return traceIDs, nil
}

// getConsumerTraceIDs 返回 consumer 关联商品追溯码列表（兼容历史数据：consumer_id 为空时按 consignee=username 兜底）
func getConsumerTraceIDs(userID uint, username string) ([]string, error) {
	var traceIDs []string
	if userID == 0 {
		return traceIDs, nil
	}
	name := strings.TrimSpace(username)
	q := database.DB.Model(&models.Product{}).Where("is_deleted = ?", false)
	if name != "" {
		q = q.Where("(consumer_id = ? OR ((consumer_id = 0 OR consumer_id IS NULL) AND consignee = ?))", userID, name)
	} else {
		q = q.Where("consumer_id = ?", userID)
	}
	if err := q.
		Pluck("trace_id", &traceIDs).Error; err != nil {
		return nil, err
	}
	return traceIDs, nil
}

type DashboardService struct{}

// DashboardStats 仪表盘统计
type DashboardStats struct {
	TotalProducts       int64 `json:"total_products"`
	BlockchainRecords   int64 `json:"blockchain_records"`
	AbnormalCount       int64 `json:"abnormal_count"`
	TodayQueries        int64 `json:"today_queries"`
	TemperatureCount    int64 `json:"temperature_count"`
	TransportCount      int64 `json:"transport_count"`
	BlockchainConnected bool  `json:"blockchain_connected"`
}

// AlertItem 告警项
type AlertItem struct {
	ID            uint      `json:"id"`
	RecordTime    time.Time `json:"record_time"`
	TraceID       string    `json:"trace_id"`
	Temperature   float64   `json:"temperature"`
	Humidity      float64   `json:"humidity"`
	Location      string    `json:"location"`
	WarehouseName string    `json:"warehouse_name"`
	Reason        string    `json:"reason"`
}

// GetDashboardStats 获取仪表盘统计；producer 看自己，logistics 看本公司，consumer 看本人收货商品，regulator 等看全部
func (s *DashboardService) GetDashboardStats(userID uint, role, companyName, username string) (*DashboardStats, error) {
	stats := &DashboardStats{}
	stats.BlockchainConnected = blockchain.IsConnected()

	var traceIDs []string
	var err error

	if role == "producer" {
		traceIDs, err = getProducerTraceIDs(userID)
		if err != nil {
			return nil, err
		}
		database.DB.Model(&models.Product{}).Where("producer_id = ? AND is_deleted = ?", userID, false).Count(&stats.TotalProducts)
	} else if role == "logistics" {
		traceIDs, err = getLogisticsTraceIDs(companyName)
		if err != nil {
			return nil, err
		}
		stats.TotalProducts = int64(len(traceIDs))
	} else if role == "consumer" {
		traceIDs, err = getConsumerTraceIDs(userID, username)
		if err != nil {
			return nil, err
		}
		stats.TotalProducts = int64(len(traceIDs))
	}

	if role == "producer" || role == "logistics" || role == "consumer" {
		if len(traceIDs) > 0 {
			ph, args := inClauseStrings(traceIDs)
			todayStart := time.Now().Truncate(24 * time.Hour)
			database.DB.Model(&models.ProductHistory{}).Where("trace_id IN ("+ph+")", args...).Count(&stats.BlockchainRecords)
			database.DB.Model(&models.ProductHistory{}).Where("trace_id IN ("+ph+") AND created_at >= ?", append(args, todayStart)...).Count(&stats.TodayQueries)
			database.DB.Model(&models.TemperatureRecord{}).Where("trace_id IN ("+ph+") AND is_abnormal = ?", append(args, true)...).Count(&stats.AbnormalCount)
			database.DB.Model(&models.TemperatureRecord{}).Where("trace_id IN ("+ph+")", args...).Count(&stats.TemperatureCount)
			database.DB.Model(&models.TransportNode{}).Where("trace_id IN ("+ph+")", args...).Count(&stats.TransportCount)
		}
		return stats, nil
	}

	// 监管及其他角色：全局统计
	database.DB.Model(&models.Product{}).Where("is_deleted = ?", false).Count(&stats.TotalProducts)
	database.DB.Model(&models.ProductHistory{}).Count(&stats.BlockchainRecords)
	database.DB.Model(&models.TemperatureRecord{}).Where("is_abnormal = ?", true).Count(&stats.AbnormalCount)
	todayStart := time.Now().Truncate(24 * time.Hour)
	database.DB.Model(&models.ProductHistory{}).Where("created_at >= ?", todayStart).Count(&stats.TodayQueries)
	database.DB.Model(&models.TemperatureRecord{}).Count(&stats.TemperatureCount)
	database.DB.Model(&models.TransportNode{}).Count(&stats.TransportCount)
	return stats, nil
}

// GetTemperatureTrend 温度趋势数据；producer 看自己，logistics 看本公司，consumer 看本人收货商品
func (s *DashboardService) GetTemperatureTrend(userID uint, role, companyName, username string, limit int) ([]models.TemperatureRecord, error) {
	if limit <= 0 {
		limit = 200
	}
	var records []models.TemperatureRecord
	q := database.DB.Model(&models.TemperatureRecord{}).Order("record_time ASC").Limit(limit)
	if role == "producer" {
		traceIDs, err := getProducerTraceIDs(userID)
		if err != nil {
			return nil, err
		}
		if len(traceIDs) == 0 {
			return records, nil
		}
		ph, args := inClauseStrings(traceIDs)
		q = q.Where("trace_id IN ("+ph+")", args...)
	} else if role == "logistics" {
		traceIDs, err := getLogisticsTraceIDs(companyName)
		if err != nil {
			return nil, err
		}
		if len(traceIDs) == 0 {
			return records, nil
		}
		ph, args := inClauseStrings(traceIDs)
		q = q.Where("trace_id IN ("+ph+")", args...)
	} else if role == "consumer" {
		traceIDs, err := getConsumerTraceIDs(userID, username)
		if err != nil {
			return nil, err
		}
		if len(traceIDs) == 0 {
			return records, nil
		}
		ph, args := inClauseStrings(traceIDs)
		q = q.Where("trace_id IN ("+ph+")", args...)
	}
	err := q.Find(&records).Error
	return records, err
}

// GetTransportNodesForMap 运输节点（物流图）；producer 看自己，logistics 看本公司，consumer 看本人收货商品
func (s *DashboardService) GetTransportNodesForMap(userID uint, role, companyName, username string, limit int) ([]models.TransportNode, error) {
	if limit <= 0 {
		limit = 500
	}
	var nodes []models.TransportNode
	q := database.DB.Model(&models.TransportNode{}).Order("trace_id, arrival_time ASC").Limit(limit)
	if role == "producer" {
		traceIDs, err := getProducerTraceIDs(userID)
		if err != nil {
			return nil, err
		}
		if len(traceIDs) == 0 {
			return nodes, nil
		}
		ph, args := inClauseStrings(traceIDs)
		q = q.Where("trace_id IN ("+ph+")", args...)
	} else if role == "logistics" {
		traceIDs, err := getLogisticsTraceIDs(companyName)
		if err != nil {
			return nil, err
		}
		if len(traceIDs) == 0 {
			return nodes, nil
		}
		ph, args := inClauseStrings(traceIDs)
		q = q.Where("trace_id IN ("+ph+")", args...)
	} else if role == "consumer" {
		traceIDs, err := getConsumerTraceIDs(userID, username)
		if err != nil {
			return nil, err
		}
		if len(traceIDs) == 0 {
			return nodes, nil
		}
		ph, args := inClauseStrings(traceIDs)
		q = q.Where("trace_id IN ("+ph+")", args...)
	}
	err := q.Find(&nodes).Error
	return nodes, err
}

// GetAlerts 告警列表；role=logistics 时按 companyName 过滤，role=warehouse 时按 warehouseID 过滤，regulator 看全部
func (s *DashboardService) GetAlerts(limit int, role string, warehouseID uint, companyName string) ([]AlertItem, error) {
	if limit <= 0 {
		limit = 100
	}
	var records []models.TemperatureRecord
	q := database.DB.Model(&models.TemperatureRecord{}).Where("is_abnormal = ?", true).Order("record_time DESC").Limit(limit)

	if role == "logistics" && companyName != "" {
		// 物流账号：只看 transport_company = companyName 的商品的告警
		traceIDs, err := getLogisticsTraceIDs(companyName)
		if err != nil {
			return nil, err
		}
		if len(traceIDs) == 0 {
			return []AlertItem{}, nil
		}
		ph, inArgs := inClauseStrings(traceIDs)
		args := append([]interface{}{true}, inArgs...)
		q = database.DB.Model(&models.TemperatureRecord{}).
			Where("is_abnormal = ? AND trace_id IN ("+ph+")", args...).
			Order("record_time DESC").Limit(limit)
	} else if role == "warehouse" && warehouseID > 0 {
		// 仓储账号：只看本账号记录的告警
		q = database.DB.Model(&models.TemperatureRecord{}).
			Where("is_abnormal = ? AND warehouse_id = ?", true, warehouseID).
			Order("record_time DESC").Limit(limit)
	}
	// regulator 或其他角色：全部告警（保持默认 q）

	if err := q.Find(&records).Error; err != nil {
		return nil, err
	}
	items := make([]AlertItem, 0, len(records))
	for _, r := range records {
		items = append(items, AlertItem{
			ID: r.ID, RecordTime: r.RecordTime, TraceID: r.TraceID,
			Temperature: r.Temperature, Humidity: r.Humidity, Location: r.Location,
			WarehouseName: r.WarehouseName, Reason: "温度超限",
		})
	}
	return items, nil
}
