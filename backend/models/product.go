package models

import (
	"time"
)

// Product 商品信息模型
type Product struct {
	ID uint `gorm:"primary_key" json:"id"`
	// product_id 唯一性仅约束「未删除」记录，见 database 中部分唯一索引；已删除商品可复用同一编号
	ProductID        string    `gorm:"not null;index" json:"product_id"`
	BatchNumber      string    `gorm:"index;not null" json:"batch_number"`
	TraceID          string    `gorm:"unique_index;not null" json:"trace_id"`
	Origin           string    `gorm:"not null" json:"origin"`
	ProductionTime   time.Time `json:"production_time"`
	TemperatureMin   float64   `json:"temperature_min"`
	TemperatureMax   float64   `json:"temperature_max"`
	HumidityMin      float64   `json:"humidity_min"`
	HumidityMax      float64   `json:"humidity_max"`
	TransportCompany string    `json:"transport_company"`
	Consignee        string    `gorm:"index" json:"consignee"` // 收货人（消费者用户名）
	ConsumerID       uint      `gorm:"index" json:"consumer_id"`
	Destination      string    `json:"destination"`
	CustomsInfo      string    `gorm:"type:text" json:"customs_info"`
	ProducerID       uint      `gorm:"index" json:"producer_id"`
	ProducerName     string    `json:"producer_name"`
	BlockchainHash   string    `gorm:"index" json:"blockchain_hash"`
	BlockchainTxHash string    `json:"blockchain_tx_hash"`
	BlockNumber      uint64    `json:"block_number"`
	IsDeleted        bool      `gorm:"default:false;index" json:"is_deleted"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ProductHistory 商品修改历史记录
type ProductHistory struct {
	ID               uint      `gorm:"primary_key" json:"id"`
	ProductID        uint      `gorm:"index;not null" json:"product_id"`
	TraceID          string    `gorm:"index;not null" json:"trace_id"`
	OperationType    string    `gorm:"not null" json:"operation_type"` // create, update, delete
	OldHash          string    `json:"old_hash"`
	NewHash          string    `json:"new_hash"`
	OperatorID       uint      `json:"operator_id"`
	OperatorName     string    `json:"operator_name"`
	BlockchainHash   string    `gorm:"index" json:"blockchain_hash"`
	BlockchainTxHash string    `json:"blockchain_tx_hash"`
	ChangeDetails    string    `gorm:"type:text" json:"change_details"`
	ChangedFields    string    `json:"changed_fields"` // 修改字段，逗号分隔，如 "destination, transport_company"
	CreatedAt        time.Time `json:"created_at"`
}

// TemperatureRecord 温度监控记录
type TemperatureRecord struct {
	ID               uint      `gorm:"primary_key" json:"id"`
	ProductID        uint      `gorm:"index;not null" json:"product_id"`
	TraceID          string    `gorm:"index;not null" json:"trace_id"`
	Temperature      float64   `json:"temperature"`
	Humidity         float64   `json:"humidity"`
	Location         string    `json:"location"`
	RecordTime       time.Time `gorm:"index" json:"record_time"`
	WarehouseID      uint      `json:"warehouse_id"`
	WarehouseName    string    `json:"warehouse_name"`
	IsAbnormal       bool      `gorm:"default:false" json:"is_abnormal"`
	BlockchainHash   string    `gorm:"index" json:"blockchain_hash"`
	BlockchainTxHash string    `json:"blockchain_tx_hash"`
	CreatedAt        time.Time `json:"created_at"`
}

// TransportNode 运输节点信息
type TransportNode struct {
	ID               uint      `gorm:"primary_key" json:"id"`
	ProductID        uint      `gorm:"index;not null" json:"product_id"`
	TraceID          string    `gorm:"index;not null" json:"trace_id"`
	NodeName         string    `json:"node_name"`
	Location         string    `json:"location"`
	ArrivalTime      time.Time `json:"arrival_time"`
	DepartureTime    time.Time `json:"departure_time"`
	LogisticsID      uint      `json:"logistics_id"`
	LogisticsName    string    `json:"logistics_name"`
	Status           string    `json:"status"` // in_transit, arrived, delivered
	BlockchainHash   string    `gorm:"index" json:"blockchain_hash"`
	BlockchainTxHash string    `json:"blockchain_tx_hash"`
	CreatedAt        time.Time `json:"created_at"`
}

// TableName 指定表名
func (Product) TableName() string {
	return "products"
}

func (ProductHistory) TableName() string {
	return "product_histories"
}

func (TemperatureRecord) TableName() string {
	return "temperature_records"
}

func (TransportNode) TableName() string {
	return "transport_nodes"
}
