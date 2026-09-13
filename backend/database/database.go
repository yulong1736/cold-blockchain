package database

import (
	"cold-chain-trace/backend/config"
	"cold-chain-trace/backend/models"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

var DB *gorm.DB

type pair struct {
	productID uint
	userID    uint
}

// InitDB 初始化数据库连接
func InitDB() error {
	dsn := config.GetDSN()

	var err error
	DB, err = gorm.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 设置连接池
	sqlDB := DB.DB()
	sqlDB.SetMaxOpenConns(config.AppConfig.Database.MaxConnections)
	sqlDB.SetMaxIdleConns(10)

	// 自动迁移
	if err := AutoMigrate(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	if err := ensureProductPartialUniqueProductID(); err != nil {
		return fmt.Errorf("failed to apply product_id partial unique index: %w", err)
	}
	if err := ensureProductConsumerIDs(); err != nil {
		return fmt.Errorf("failed to backfill product consumer_id: %w", err)
	}

	log.Println("Database connected successfully")
	return nil
}

func buildProductConsumerBackfillPairs(products []models.Product, users []models.User) []pair {
	consumerByName := make(map[string]uint)
	for _, u := range users {
		if u.Role == models.RoleConsumer && u.Username != "" {
			consumerByName[u.Username] = u.ID
		}
	}
	pairs := make([]pair, 0)
	for _, p := range products {
		if p.ConsumerID != 0 || strings.TrimSpace(p.Consignee) == "" {
			continue
		}
		if uid, ok := consumerByName[p.Consignee]; ok {
			pairs = append(pairs, pair{productID: p.ID, userID: uid})
		}
	}
	return pairs
}

// ensureProductConsumerIDs 一次性增量回填 products.consumer_id，避免改名后按用户名匹配导致归属漂移。
func ensureProductConsumerIDs() error {
	var products []models.Product
	if err := DB.Model(&models.Product{}).
		Select("id, consignee, consumer_id").
		Where("(consumer_id = 0 OR consumer_id IS NULL) AND consignee <> ''").
		Find(&products).Error; err != nil {
		return err
	}
	if len(products) == 0 {
		return nil
	}
	var users []models.User
	if err := DB.Model(&models.User{}).
		Select("id, username, role").
		Where("role = ? AND username <> ''", models.RoleConsumer).
		Find(&users).Error; err != nil {
		return err
	}
	pairs := buildProductConsumerBackfillPairs(products, users)
	if len(pairs) == 0 {
		return nil
	}
	tx := DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	for _, p := range pairs {
		if err := tx.Model(&models.Product{}).
			Where("id = ? AND (consumer_id = 0 OR consumer_id IS NULL)", p.productID).
			Update("consumer_id", p.userID).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	log.Printf("Backfilled consumer_id for %d products", len(pairs))
	return nil
}

// ensureProductPartialUniqueProductID 将 product_id 的全局唯一改为「仅未删除行唯一」，
// 避免逻辑删除后无法再用同一商品编号创建。
func ensureProductPartialUniqueProductID() error {
	sqlDB := DB.DB()
	if err := dropLegacyProductIDUniqueness(sqlDB); err != nil {
		return err
	}
	_, err := sqlDB.Exec(`
CREATE UNIQUE INDEX IF NOT EXISTS idx_products_product_id_active
ON products (product_id) WHERE (is_deleted = false);
`)
	if err != nil {
		return err
	}
	return nil
}

func dropLegacyProductIDUniqueness(sqlDB *sql.DB) error {
	// 去掉 GORM 曾创建的 product_id 唯一约束 / 唯一索引（名称因版本而异，逐个尝试）
	for _, q := range []string{
		`ALTER TABLE products DROP CONSTRAINT IF EXISTS products_product_id_key`,
		`ALTER TABLE products DROP CONSTRAINT IF EXISTS uni_products_product_id`,
		`DROP INDEX IF EXISTS uix_products_product_id`,
		`DROP INDEX IF EXISTS idx_products_product_id`,
	} {
		if _, err := sqlDB.Exec(q); err != nil {
			// DROP IF EXISTS 在部分驱动上仍可能报错，忽略「不存在」类错误
			if !strings.Contains(strings.ToLower(err.Error()), "does not exist") {
				log.Printf("drop legacy product_id uniqueness (%s): %v", q, err)
			}
		}
	}
	return nil
}

// AutoMigrate 自动迁移数据库表
func AutoMigrate() error {
	return DB.AutoMigrate(
		&models.User{},
		&models.PasswordResetCode{},
		&models.Product{},
		&models.ProductHistory{},
		&models.TemperatureRecord{},
		&models.TransportNode{},
	).Error
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
