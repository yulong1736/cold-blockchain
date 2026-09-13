package services

import (
	"cold-chain-trace/backend/blockchain"
	"cold-chain-trace/backend/database"
	"cold-chain-trace/backend/models"
	"cold-chain-trace/backend/utils"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

var (
	ErrProductNotFound   = errors.New("product not found")
	ErrProductForbidden  = errors.New("product forbidden")
	ErrProductValidation = errors.New("product validation failed")
)

func newValidationError(msg string) error {
	return fmt.Errorf("%w: %s", ErrProductValidation, msg)
}

func newForbiddenError(msg string) error {
	return fmt.Errorf("%w: %s", ErrProductForbidden, msg)
}

func validateTransportTimes(arrival, departure, prevDeparture time.Time) error {
	if arrival.IsZero() || departure.IsZero() {
		return newValidationError("到达时间和离开时间不能为空")
	}
	if departure.Before(arrival) {
		return newValidationError("离开时间不能早于到达时间")
	}
	if !prevDeparture.IsZero() && arrival.Before(prevDeparture) {
		return newValidationError("到达时间不得早于前一个节点离开时间")
	}
	return nil
}

// getFloat 从 updates 中取 key 对应的数值，若无或解析失败则返回 defaultVal（用于更新前温湿度校验）
func getFloat(updates map[string]interface{}, key string, defaultVal float64) (float64, bool) {
	v, ok := updates[key]
	if !ok {
		return defaultVal, false
	}
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	default:
		return defaultVal, false
	}
}

type ProductService struct{}

func (s *ProductService) ensureConsumerExists(username string) error {
	name := strings.TrimSpace(username)
	if name == "" {
		return newValidationError("收货人不能为空")
	}
	var count int64
	if err := database.DB.Model(&models.User{}).
		Where("username = ? AND role = ?", name, models.RoleConsumer).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return newValidationError(fmt.Sprintf("收货人不存在或不是消费者账号：%s", name))
	}
	return nil
}

func (s *ProductService) resolveConsumer(username string) (*models.User, error) {
	name := strings.TrimSpace(username)
	if name == "" {
		return nil, newValidationError("收货人不能为空")
	}
	var user models.User
	if err := database.DB.Where("username = ? AND role = ?", name, models.RoleConsumer).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, newValidationError(fmt.Sprintf("收货人不存在或不是消费者账号：%s", name))
		}
		return nil, err
	}
	return &user, nil
}

func canConsumerAccessProduct(product *models.Product, consumerID uint) bool {
	if product == nil || consumerID == 0 {
		return false
	}
	return product.ConsumerID == consumerID
}

func backfillProductConsumerID(product *models.Product, consumerID uint) {
	if product == nil || product.ID == 0 || consumerID == 0 || product.ConsumerID == consumerID {
		return
	}
	_ = database.DB.Model(&models.Product{}).
		Where("id = ? AND consumer_id = 0", product.ID).
		Update("consumer_id", consumerID).Error
}

// CreateTemperatureRecord 创建温控记录（仓储），返回新记录 id
func (s *ProductService) CreateTemperatureRecord(traceID string, temperature, humidity float64, location string, warehouseID uint, warehouseName string) (uint, error) {
	var product models.Product
	if err := database.DB.Where("trace_id = ? AND is_deleted = ?", traceID, false).First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("%w: trace_id=%s", ErrProductNotFound, traceID)
		}
		return 0, err
	}

	record := &models.TemperatureRecord{
		ProductID:     product.ID,
		TraceID:       traceID,
		Temperature:   temperature,
		Humidity:      humidity,
		Location:      location,
		RecordTime:    time.Now(),
		WarehouseID:   warehouseID,
		WarehouseName: warehouseName,
	}

	// 简单异常判定：超出商品设定的温度范围
	if product.TemperatureMin != 0 || product.TemperatureMax != 0 {
		if temperature < product.TemperatureMin || temperature > product.TemperatureMax {
			record.IsAbnormal = true
		}
	}

	if err := database.DB.Create(record).Error; err != nil {
		return 0, err
	}
	dataHash := utils.GenerateTemperatureRecordHash(
		record.ID, record.TraceID, record.Temperature, record.Humidity, record.Location,
		record.RecordTime, record.WarehouseID, record.WarehouseName, record.IsAbnormal,
	)
	if err := database.DB.Model(record).Update("blockchain_hash", dataHash).Error; err != nil {
		return 0, err
	}
	s.asyncAttachMonitorBlockchain(blockchain.EvidenceKindTemperature, record.ID, record.TraceID, dataHash, warehouseName, &models.TemperatureRecord{})
	return record.ID, nil
}

// ListTemperatureRecordsByWarehouse 按仓储账号查看其记录
func (s *ProductService) ListTemperatureRecordsByWarehouse(warehouseID uint) ([]models.TemperatureRecord, error) {
	var records []models.TemperatureRecord
	if err := database.DB.
		Where("warehouse_id = ?", warehouseID).
		Order("record_time DESC").
		Limit(200).
		Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// CreateTransportNode 创建运输节点（物流）：只允许给 transport_company 与当前账号 company_name 匹配的商品添加节点；departure 可选
func (s *ProductService) CreateTransportNode(traceID, nodeName, location, status string, logisticsID uint, logisticsName, companyName string, departure *time.Time) (uint, error) {
	var product models.Product
	if err := database.DB.Where("trace_id = ? AND is_deleted = ?", traceID, false).First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("%w: trace_id=%s", ErrProductNotFound, traceID)
		}
		return 0, err
	}

	// 多租户隔离：仅允许操作 transport_company 与当前物流公司名称匹配的商品
	if companyName != "" && product.TransportCompany != companyName {
		return 0, newForbiddenError(fmt.Sprintf("该商品的运输公司为「%s」，与当前账号所属公司「%s」不匹配", product.TransportCompany, companyName))
	}

	if status == "" {
		status = "in_transit"
	}
	arrivalTime := time.Now()

	// 运输节点顺序校验：后一个节点到达时间不能早于前一个节点离开时间
	var prevNode models.TransportNode
	err := database.DB.
		Where("trace_id = ?", traceID).
		Order("arrival_time DESC, id DESC").
		First(&prevNode).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	if err == nil && !prevNode.DepartureTime.IsZero() && arrivalTime.Before(prevNode.DepartureTime) {
		return 0, newValidationError("新增节点到达时间不得早于前一个节点离开时间")
	}

	node := &models.TransportNode{
		ProductID:     product.ID,
		TraceID:       traceID,
		NodeName:      nodeName,
		Location:      location,
		ArrivalTime:   arrivalTime,
		LogisticsID:   logisticsID,
		LogisticsName: logisticsName,
		Status:        status,
	}
	if departure != nil {
		node.DepartureTime = *departure
	}

	if err := database.DB.Create(node).Error; err != nil {
		return 0, err
	}
	dataHash := utils.GenerateTransportNodeHash(
		node.ID, node.TraceID, node.NodeName, node.Location, node.ArrivalTime, node.DepartureTime,
		node.LogisticsID, node.LogisticsName, node.Status,
	)
	if err := database.DB.Model(node).Update("blockchain_hash", dataHash).Error; err != nil {
		return 0, err
	}
	s.asyncAttachMonitorBlockchain(blockchain.EvidenceKindLogistics, node.ID, node.TraceID, dataHash, logisticsName, &models.TransportNode{})
	return node.ID, nil
}

// UpdateTransportNodeTimes 更新到达/离开时间并重算存证哈希、异步再上链
func (s *ProductService) UpdateTransportNodeTimes(nodeID, logisticsID uint, companyName string, arrival, departure time.Time, operatorName string) error {
	var node models.TransportNode
	if err := database.DB.First(&node, nodeID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: 运输节点不存在", ErrProductNotFound)
		}
		return err
	}
	if node.LogisticsID != logisticsID {
		return newForbiddenError("无权操作该运输节点")
	}
	var product models.Product
	if err := database.DB.Where("trace_id = ? AND is_deleted = ?", node.TraceID, false).First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: 关联商品不存在", ErrProductNotFound)
		}
		return err
	}
	if companyName != "" && product.TransportCompany != companyName {
		return newForbiddenError("该节点所属商品不属于当前物流公司")
	}

	// 同追溯码下：当前节点到达时间不得早于上一节点离开时间
	var prevNode models.TransportNode
	prevErr := database.DB.
		Where("trace_id = ? AND id <> ?", node.TraceID, node.ID).
		Order("arrival_time DESC, id DESC").
		First(&prevNode).Error
	if prevErr != nil && !errors.Is(prevErr, gorm.ErrRecordNotFound) {
		return prevErr
	}
	if err := validateTransportTimes(arrival, departure, prevNode.DepartureTime); err != nil {
		return err
	}

	node.ArrivalTime = arrival
	node.DepartureTime = departure
	dataHash := utils.GenerateTransportNodeHash(
		node.ID, node.TraceID, node.NodeName, node.Location, node.ArrivalTime, node.DepartureTime,
		node.LogisticsID, node.LogisticsName, node.Status,
	)
	updates := map[string]interface{}{
		"arrival_time":       arrival,
		"departure_time":     departure,
		"blockchain_hash":    dataHash,
		"blockchain_tx_hash": "",
	}
	if err := database.DB.Model(&node).Updates(updates).Error; err != nil {
		return err
	}
	s.asyncAttachMonitorBlockchain(blockchain.EvidenceKindLogistics, node.ID, node.TraceID, dataHash, operatorName, &models.TransportNode{})
	return nil
}

// SyncTemperatureRecordBlockchain 手动同步单条温控记录上链（仓储本人数据）
func (s *ProductService) SyncTemperatureRecordBlockchain(recordID, warehouseID uint, operatorName string) error {
	var r models.TemperatureRecord
	if err := database.DB.First(&r, recordID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: 温控记录不存在", ErrProductNotFound)
		}
		return err
	}
	if r.WarehouseID != warehouseID {
		return newForbiddenError("无权操作该温控记录")
	}
	dataHash := r.BlockchainHash
	if dataHash == "" {
		dataHash = utils.GenerateTemperatureRecordHash(
			r.ID, r.TraceID, r.Temperature, r.Humidity, r.Location,
			r.RecordTime, r.WarehouseID, r.WarehouseName, r.IsAbnormal,
		)
		_ = database.DB.Model(&r).Update("blockchain_hash", dataHash).Error
		r.BlockchainHash = dataHash
	}
	txHash, _, err := blockchain.StoreEvidenceWithKind(blockchain.EvidenceKindTemperature, r.TraceID, dataHash, operatorName)
	if txHash == "" {
		if err != nil {
			return fmt.Errorf("上链失败: %w", err)
		}
		return errors.New("上链失败：未获得交易哈希")
	}
	if err != nil {
		if errors.Is(err, blockchain.ErrReceiptTimeout) {
			return fmt.Errorf("%w: 上链超时（30s）：交易已发送，等待链上确认中。tx=%s", blockchain.ErrReceiptTimeout, txHash)
		}
		return fmt.Errorf("上链失败: %w", err)
	}
	return database.DB.Model(&r).Update("blockchain_tx_hash", txHash).Error
}

// SyncTransportNodeBlockchain 手动同步单条运输节点上链（本公司商品 + 本人节点）
func (s *ProductService) SyncTransportNodeBlockchain(nodeID, logisticsID uint, companyName, operatorName string) error {
	var n models.TransportNode
	if err := database.DB.First(&n, nodeID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: 运输节点不存在", ErrProductNotFound)
		}
		return err
	}
	if n.LogisticsID != logisticsID {
		return newForbiddenError("无权操作该运输节点")
	}
	var product models.Product
	if err := database.DB.Where("trace_id = ? AND is_deleted = ?", n.TraceID, false).First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: 关联商品不存在", ErrProductNotFound)
		}
		return err
	}
	if companyName != "" && product.TransportCompany != companyName {
		return newForbiddenError("该节点所属商品不属于当前物流公司")
	}
	dataHash := n.BlockchainHash
	if dataHash == "" {
		dataHash = utils.GenerateTransportNodeHash(
			n.ID, n.TraceID, n.NodeName, n.Location, n.ArrivalTime, n.DepartureTime,
			n.LogisticsID, n.LogisticsName, n.Status,
		)
		_ = database.DB.Model(&n).Update("blockchain_hash", dataHash).Error
		n.BlockchainHash = dataHash
	}
	txHash, _, err := blockchain.StoreEvidenceWithKind(blockchain.EvidenceKindLogistics, n.TraceID, dataHash, operatorName)
	if txHash == "" {
		if err != nil {
			return fmt.Errorf("上链失败: %w", err)
		}
		return errors.New("上链失败：未获得交易哈希")
	}
	if err != nil {
		if errors.Is(err, blockchain.ErrReceiptTimeout) {
			return fmt.Errorf("%w: 上链超时（30s）：交易已发送，等待链上确认中。tx=%s", blockchain.ErrReceiptTimeout, txHash)
		}
		return fmt.Errorf("上链失败: %w", err)
	}
	return database.DB.Model(&n).Update("blockchain_tx_hash", txHash).Error
}

// ListTransportNodesByLogistics 按物流公司名称（多租户隔离）查看运输节点
// companyName 为空时退化为按 logistics_id 过滤（兼容老数据）
func (s *ProductService) ListTransportNodesByLogistics(logisticsID uint, companyName string) ([]models.TransportNode, error) {
	var nodes []models.TransportNode
	q := database.DB.Order("arrival_time DESC").Limit(200)
	if companyName != "" {
		// 通过 products 表的 transport_company 字段过滤追溯码，再查对应节点
		var traceIDs []string
		if err := database.DB.Model(&models.Product{}).
			Where("transport_company = ? AND is_deleted = ?", companyName, false).
			Pluck("trace_id", &traceIDs).Error; err != nil {
			return nil, err
		}
		if len(traceIDs) == 0 {
			return nodes, nil
		}
		ph, args := inClauseStrings(traceIDs)
		q = q.Where("trace_id IN ("+ph+")", args...)
	} else {
		q = q.Where("logistics_id = ?", logisticsID)
	}
	if err := q.Find(&nodes).Error; err != nil {
		return nil, err
	}
	return nodes, nil
}

// ListProductsByLogisticsCompany 返回 transport_company = companyName 的商品（供物流账号添加节点时选追溯码）
func (s *ProductService) ListProductsByLogisticsCompany(companyName string) ([]models.Product, error) {
	var products []models.Product
	if companyName == "" {
		return products, nil
	}
	if err := database.DB.Where("transport_company = ? AND is_deleted = ?", companyName, false).
		Order("created_at DESC").Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// ListProductsByProducer 获取生产商的商品列表
func (s *ProductService) ListProductsByProducer(producerID uint, producerName string) ([]models.Product, error) {
	var products []models.Product

	//兼容历史数据：早期创建的商品可能没有写入 producer_id / producer_name（为 0 / 空）。
	//通过 product_histories 的 create 记录把它们归属到当前生产商。
	legacyIDsQuery := database.DB.
		Table(models.ProductHistory{}.TableName()).
		Select("product_id").
		Where("operation_type = ? AND operator_id = ?", "create", producerID).
		SubQuery()

	if err := database.DB.
		Where(
			"is_deleted = ? AND (producer_id = ? OR (producer_id = 0 AND id IN (?)))",
			false,
			producerID,
			legacyIDsQuery,
		).
		Order("created_at DESC").
		Find(&products).Error; err != nil {
		return nil, err
	}

	// 回填历史数据的 producer 字段，避免下次还走兼容逻辑
	for i := range products {
		if products[i].ProducerID == 0 {
			products[i].ProducerID = producerID
			products[i].ProducerName = producerName
			_ = database.DB.Model(&models.Product{}).
				Where("id = ? AND producer_id = 0", products[i].ID).
				Updates(map[string]interface{}{
					"producer_id":   producerID,
					"producer_name": producerName,
				}).Error
		}
	}

	return products, nil
}

// asyncAttachProductBlockchain 异步上链并回写商品；historyID>0 时同时回写该条 product_histories 的 tx_hash。
// 避免 HTTP 请求被前端 axios 超时（默认 10s）打断，出现「提示失败但库里有记录」。
func (s *ProductService) asyncAttachProductBlockchain(productID, historyID uint, traceID, dataHash, operatorName string) {
	go func() {
		txHash, err := blockchain.StoreEvidenceNoWait(traceID, dataHash, operatorName)
		if txHash != "" {
			updates := map[string]interface{}{"blockchain_tx_hash": txHash}
			if uerr := database.DB.Model(&models.Product{}).Where("id = ?", productID).Updates(updates).Error; uerr != nil {
				fmt.Printf("Warning: async product blockchain fields: %v\n", uerr)
			}
			if historyID > 0 {
				_ = database.DB.Model(&models.ProductHistory{}).Where("id = ?", historyID).Update("blockchain_tx_hash", txHash).Error
			}
		}
		if err != nil {
			fmt.Printf("Warning: async blockchain evidence (product %d): %v\n", productID, err)
		}
	}()
}

// asyncAttachMonitorBlockchain 温控/物流记录异步上链并回写 blockchain_tx_hash
func (s *ProductService) asyncAttachMonitorBlockchain(chainKind string, rowID uint, traceID, dataHash, operator string, model interface{}) {
	go func() {
		txHash, err := blockchain.StoreEvidenceNoWaitWithKind(chainKind, traceID, dataHash, operator)
		if txHash != "" {
			if uerr := database.DB.Model(model).Where("id = ?", rowID).Update("blockchain_tx_hash", txHash).Error; uerr != nil {
				fmt.Printf("Warning: async monitor blockchain tx_hash: %v\n", uerr)
			}
		}
		if err != nil {
			fmt.Printf("Warning: async blockchain evidence (%s row=%d trace=%s): %v\n", chainKind, rowID, traceID, err)
		}
	}()
}

// CreateProduct 创建商品信息（数据库+链上存证）
func (s *ProductService) CreateProduct(product *models.Product, operatorID uint, operatorName string) error {
	product.Consignee = strings.TrimSpace(product.Consignee)
	consumer, err := s.resolveConsumer(product.Consignee)
	if err != nil {
		return err
	}
	product.ConsumerID = consumer.ID

	var dup models.Product
	if err := database.DB.Where("product_id = ? AND is_deleted = ?", product.ProductID, false).First(&dup).Error; err == nil {
		return fmt.Errorf("商品编号「%s」已被使用，请更换其他编号", product.ProductID)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// 0. 温湿度参数校验：最高温度必须大于最低温度，最高湿度必须大于最低湿度
	if product.TemperatureMax <= product.TemperatureMin {
		return newValidationError("商品温度参数错误：最高温度必须大于最低温度")
	}
	if product.HumidityMax <= product.HumidityMin {
		return newValidationError("商品湿度参数错误：最高湿度必须大于最低湿度")
	}

	// 1. 生成追溯码
	product.TraceID = utils.GenerateTraceID(product.ProductID, product.BatchNumber, "signature")

	// 2. 计算数据哈希
	dataHash := utils.GenerateProductHash(
		product.ProductID,
		product.BatchNumber,
		product.Origin,
		product.ProductionTime.Format(time.RFC3339),
	)
	product.BlockchainHash = dataHash

	// 3. 写入数据库
	if err := database.DB.Create(product).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate key") ||
			strings.Contains(strings.ToLower(err.Error()), "unique constraint") ||
			strings.Contains(strings.ToLower(err.Error()), "idx_products_product_id_active") {
			return newValidationError(fmt.Sprintf("商品编号「%s」与已有记录冲突（或追溯码重复），请稍后再试或更换编号", product.ProductID))
		}
		return fmt.Errorf("failed to create product in database: %w", err)
	}

	// 4. 记录操作历史（链上存证异步完成，避免阻塞 HTTP 超过前端超时时间）
	history := &models.ProductHistory{
		ProductID:        product.ID,
		TraceID:          product.TraceID,
		OperationType:    "create",
		NewHash:          dataHash,
		OperatorID:       operatorID,
		OperatorName:     operatorName,
		BlockchainHash:   dataHash,
		BlockchainTxHash: "",
		ChangeDetails:    "商品信息创建",
		ChangedFields:    "—",
	}
	if err := database.DB.Create(history).Error; err != nil {
		return fmt.Errorf("记录操作历史失败: %w", err)
	}
	s.asyncAttachProductBlockchain(product.ID, history.ID, product.TraceID, dataHash, operatorName)

	return nil
}

// SyncProductBlockchain 对已有商品重新发起链上存证（修复「待上链」或历史上链失败）
func (s *ProductService) SyncProductBlockchain(productID uint, producerID uint, operatorName string) error {
	var product models.Product
	if err := database.DB.First(&product, productID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: 商品不存在", ErrProductNotFound)
		}
		return err
	}
	if product.IsDeleted {
		return newValidationError("已删除的商品无法同步上链")
	}
	if product.ProducerID != producerID {
		return newForbiddenError("无权操作该商品")
	}
	dataHash := product.BlockchainHash
	if dataHash == "" {
		dataHash = utils.GenerateProductHash(
			product.ProductID,
			product.BatchNumber,
			product.Origin,
			product.ProductionTime.Format(time.RFC3339),
		)
		_ = database.DB.Model(&product).Update("blockchain_hash", dataHash).Error
		product.BlockchainHash = dataHash
	}
	txHash, blockNum, err := blockchain.StoreEvidence(product.TraceID, dataHash, operatorName)
	if txHash != "" {
		product.BlockchainTxHash = txHash
		if blockNum != nil {
			product.BlockNumber = blockNum.Uint64()
		}
		if saveErr := database.DB.Save(&product).Error; saveErr != nil {
			return fmt.Errorf("链上交易已发送(tx=%s)，但写入数据库失败: %w", txHash, saveErr)
		}
	}
	if txHash == "" {
		if err != nil {
			return fmt.Errorf("上链失败: %w", err)
		}
		return errors.New("上链失败：未获得交易哈希")
	}
	if err != nil {
		if errors.Is(err, blockchain.ErrReceiptTimeout) {
			return fmt.Errorf("%w: 上链超时（30s）：交易已发送，等待链上确认中。tx=%s", blockchain.ErrReceiptTimeout, txHash)
		}
		return fmt.Errorf("上链失败: %w", err)
	}
	return nil
}

// SyncProductBlockchainByTraceID 按追溯码同步上链（与按数字 id 同步等价，避免搞错路径参数）
func (s *ProductService) SyncProductBlockchainByTraceID(traceID string, producerID uint, operatorName string) error {
	var product models.Product
	if err := database.DB.Where("trace_id = ? AND is_deleted = ?", traceID, false).First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("未找到该追溯码对应的商品")
		}
		return err
	}
	return s.SyncProductBlockchain(product.ID, producerID, operatorName)
}

// GetProductByTraceID 根据追溯码查询商品信息
func (s *ProductService) GetProductByTraceID(traceID string) (*models.Product, error) {
	var product models.Product
	if err := database.DB.Where("trace_id = ? AND is_deleted = ?", traceID, false).First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: trace_id=%s", ErrProductNotFound, traceID)
		}
		return nil, err
	}

	// 验证链上数据
	chainData, err := blockchain.QueryEvidence(traceID)
	if err == nil {
		// 验证哈希是否一致
		if chainHash, ok := chainData["hash"].(string); ok {
			if chainHash != product.BlockchainHash {
				return nil, errors.New("data integrity check failed: hash mismatch")
			}
		}
	}

	return &product, nil
}

// GetProductByTraceIDForConsumer 消费者查询：仅允许查询收货人为自己的商品
func (s *ProductService) GetProductByTraceIDForConsumer(traceID string, consumerID uint, consumerUsername string) (*models.Product, error) {
	product, err := s.GetProductByTraceID(traceID)
	if err != nil {
		return nil, err
	}

	if canConsumerAccessProduct(product, consumerID) {
		return product, nil
	}

	// 兼容历史数据：consumer_id 为空时，允许按 consignee=username 访问并自动回填 consumer_id。
	if product.ConsumerID == 0 && strings.TrimSpace(consumerUsername) != "" &&
		strings.TrimSpace(product.Consignee) == strings.TrimSpace(consumerUsername) {
		backfillProductConsumerID(product, consumerID)
		product.ConsumerID = consumerID
		return product, nil
	}

	if !canConsumerAccessProduct(product, consumerID) {
		return nil, newForbiddenError("无权查询该商品")
	}
	return product, nil
}

// ListProductsByConsumerID 查询消费者名下商品
func (s *ProductService) ListProductsByConsumerID(consumerID uint, consumerUsername string) ([]models.Product, error) {
	var products []models.Product
	if consumerID == 0 {
		return products, nil
	}
	normalizedName := strings.TrimSpace(consumerUsername)
	q := database.DB.Where("is_deleted = ?", false)
	if normalizedName != "" {
		// 优先稳定归属 consumer_id，同时兼容老数据（consumer_id 为 0 或 NULL 时仅有 consignee）。
		q = q.Where("(consumer_id = ? OR ((consumer_id = 0 OR consumer_id IS NULL) AND consignee = ?))", consumerID, normalizedName)
	} else {
		q = q.Where("consumer_id = ?", consumerID)
	}
	if err := q.
		Order("created_at DESC").
		Find(&products).Error; err != nil {
		return nil, err
	}

	// 命中历史数据时自动回填 consumer_id，后续查询走稳定主键。
	for i := range products {
		if products[i].ConsumerID == 0 && normalizedName != "" && strings.TrimSpace(products[i].Consignee) == normalizedName {
			backfillProductConsumerID(&products[i], consumerID)
			products[i].ConsumerID = consumerID
		}
	}
	return products, nil
}

// 业务字段白名单：仅这些字段参与审计的「变更字段」与键值对展示
var auditAllowedKeys = map[string]bool{
	"product_id": true, "batch_number": true, "origin": true, "production_time": true,
	"temperature_min": true, "temperature_max": true, "humidity_min": true, "humidity_max": true,
	"transport_company": true, "consignee": true, "destination": true, "customs_info": true,
}

func getProductOldValueString(product *models.Product, key string) string {
	empty := func(s string) string {
		if s == "" {
			return "空"
		}
		return s
	}
	switch key {
	case "product_id":
		return empty(product.ProductID)
	case "batch_number":
		return empty(product.BatchNumber)
	case "origin":
		return empty(product.Origin)
	case "production_time":
		return product.ProductionTime.Format(time.RFC3339)
	case "temperature_min":
		return strconv.FormatFloat(product.TemperatureMin, 'f', -1, 64)
	case "temperature_max":
		return strconv.FormatFloat(product.TemperatureMax, 'f', -1, 64)
	case "humidity_min":
		return strconv.FormatFloat(product.HumidityMin, 'f', -1, 64)
	case "humidity_max":
		return strconv.FormatFloat(product.HumidityMax, 'f', -1, 64)
	case "transport_company":
		return empty(product.TransportCompany)
	case "consignee":
		return empty(product.Consignee)
	case "destination":
		return empty(product.Destination)
	case "customs_info":
		return empty(product.CustomsInfo)
	default:
		return ""
	}
}

func formatUpdateValue(v interface{}) string {
	if v == nil {
		return "空"
	}
	switch val := v.(type) {
	case string:
		if val == "" {
			return "空"
		}
		// 尝试解析为时间并统一为 RFC3339，便于与数据库旧值比较
		if t, err := time.Parse(time.RFC3339, val); err == nil {
			return t.Format(time.RFC3339)
		}
		if t, err := time.Parse("2006-01-02T15:04:05Z07:00", val); err == nil {
			return t.Format(time.RFC3339)
		}
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case uint:
		return strconv.FormatUint(uint64(val), 10)
	case bool:
		if val {
			return "是"
		}
		return "否"
	case time.Time:
		return val.Format(time.RFC3339)
	default:
		return fmt.Sprint(v)
	}
}

// buildUpdateAuditText 根据更新前的商品与 updates 生成键值对描述（字段: 旧值 -> 新值）
// 只处理业务白名单字段；旧值取自数据库当前值，新值取自 updates
func buildUpdateAuditText(product *models.Product, updates map[string]interface{}) (fieldNames string, kvPairs string) {
	var names []string
	var pairs []string
	for k, newVal := range updates {
		if !auditAllowedKeys[k] {
			continue
		}
		oldStr := getProductOldValueString(product, k)
		newStr := formatUpdateValue(newVal)
		// 即使相等也记录（前端已做 diff，此处兜底）
		names = append(names, k)
		if oldStr != newStr {
			pairs = append(pairs, fmt.Sprintf("%s: %s -> %s", k, oldStr, newStr))
		} else {
			pairs = append(pairs, fmt.Sprintf("%s: %s（未变更）", k, oldStr))
		}
	}
	if len(names) == 0 {
		return "—", "—"
	}
	return strings.Join(names, ", "), strings.Join(pairs, "; ")
}

// UpdateProduct 更新商品信息（可追踪）
func (s *ProductService) UpdateProduct(productID uint, updates map[string]interface{}, operatorID uint, operatorName string) error {
	var product models.Product
	if err := database.DB.First(&product, productID).Error; err != nil {
		return err
	}

	if product.IsDeleted {
		return errors.New("cannot update deleted product")
	}

	if raw, ok := updates["consignee"]; ok {
		newConsignee := strings.TrimSpace(fmt.Sprint(raw))
		consumer, err := s.resolveConsumer(newConsignee)
		if err != nil {
			return err
		}
		updates["consignee"] = newConsignee
		updates["consumer_id"] = consumer.ID
	}

	// 更新前校验温湿度：合并 updates 与当前值后校验，避免写入非法数据
	tempMin, _ := getFloat(updates, "temperature_min", product.TemperatureMin)
	tempMax, _ := getFloat(updates, "temperature_max", product.TemperatureMax)
	humMin, _ := getFloat(updates, "humidity_min", float64(product.HumidityMin))
	humMax, _ := getFloat(updates, "humidity_max", float64(product.HumidityMax))
	if tempMax <= tempMin {
		return newValidationError("商品温度参数错误：最高温度必须大于最低温度")
	}
	if humMax <= humMin {
		return newValidationError("商品湿度参数错误：最高湿度必须大于最低湿度")
	}

	oldHash := product.BlockchainHash
	// 在更新前生成审计用键值对（基于旧值）
	_, changedKV := buildUpdateAuditText(&product, updates)

	// 更新数据库
	if err := database.DB.Model(&product).Updates(updates).Error; err != nil {
		return err
	}

	// 重新读取
	database.DB.First(&product, productID)

	// 重新计算哈希
	newHash := utils.GenerateProductHash(
		product.ProductID,
		product.BatchNumber,
		product.Origin,
		product.ProductionTime.Format(time.RFC3339),
	)
	product.BlockchainHash = newHash
	database.DB.Save(&product)

	// 记录历史；链上异步写入
	history := &models.ProductHistory{
		ProductID:        product.ID,
		TraceID:          product.TraceID,
		OperationType:    "update",
		OldHash:          oldHash,
		NewHash:          newHash,
		OperatorID:       operatorID,
		OperatorName:     operatorName,
		BlockchainHash:   newHash,
		BlockchainTxHash: "",
		ChangeDetails:    changedKV,
		ChangedFields:    changedKV,
	}
	if err := database.DB.Create(history).Error; err != nil {
		return err
	}
	s.asyncAttachProductBlockchain(product.ID, history.ID, product.TraceID, newHash, operatorName)

	return nil
}

// DeleteProduct 逻辑删除商品
func (s *ProductService) DeleteProduct(productID uint, operatorID uint, operatorName string) error {
	var product models.Product
	if err := database.DB.First(&product, productID).Error; err != nil {
		return err
	}

	// 逻辑删除
	product.IsDeleted = true
	if err := database.DB.Save(&product).Error; err != nil {
		return err
	}

	deleteHash := utils.CalculateHash(fmt.Sprintf("delete:%s:%d", product.TraceID, time.Now().Unix()))

	// 记录历史；链上异步写入
	history := &models.ProductHistory{
		ProductID:        product.ID,
		TraceID:          product.TraceID,
		OperationType:    "delete",
		OldHash:          product.BlockchainHash,
		NewHash:          deleteHash,
		OperatorID:       operatorID,
		OperatorName:     operatorName,
		BlockchainHash:   deleteHash,
		BlockchainTxHash: "",
		ChangeDetails:    "商品信息删除",
		ChangedFields:    "—",
	}
	if err := database.DB.Create(history).Error; err != nil {
		return err
	}
	s.asyncAttachProductBlockchain(product.ID, history.ID, product.TraceID, deleteHash, operatorName)

	return nil
}

// GetProductHistory 获取商品操作历史
func (s *ProductService) GetProductHistory(traceID string) ([]models.ProductHistory, error) {
	var histories []models.ProductHistory
	if err := database.DB.Where("trace_id = ?", traceID).Order("created_at DESC").Find(&histories).Error; err != nil {
		return nil, err
	}
	return histories, nil
}

// GetAuditLogs 监管端：审计日志（谁在什么时候做了什么操作），可按追溯码筛选
func (s *ProductService) GetAuditLogs(traceID string, limit int) ([]models.ProductHistory, error) {
	if limit <= 0 {
		limit = 500
	}
	var histories []models.ProductHistory
	q := database.DB.Model(&models.ProductHistory{}).Order("created_at DESC").Limit(limit)
	if traceID != "" {
		q = q.Where("trace_id = ?", traceID)
	}
	if err := q.Find(&histories).Error; err != nil {
		return nil, err
	}
	return histories, nil
}

// GetTemperatureRecords 获取温度监控记录
func (s *ProductService) GetTemperatureRecords(traceID string) ([]models.TemperatureRecord, error) {
	var records []models.TemperatureRecord
	if err := database.DB.Where("trace_id = ?", traceID).Order("record_time ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// GetTransportNodesByTraceID 按追溯码查询运输节点（公开溯源用）
func (s *ProductService) GetTransportNodesByTraceID(traceID string) ([]models.TransportNode, error) {
	var nodes []models.TransportNode
	if err := database.DB.Where("trace_id = ?", traceID).Order("arrival_time ASC").Find(&nodes).Error; err != nil {
		return nil, err
	}
	return nodes, nil
}
