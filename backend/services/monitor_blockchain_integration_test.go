package services

import (
	"cold-chain-trace/backend/blockchain"
	"cold-chain-trace/backend/config"
	"cold-chain-trace/backend/database"
	"cold-chain-trace/backend/models"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func testRepoConfigPath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	svcDir := filepath.Dir(file)
	root := filepath.Join(svcDir, "..", "..")
	return filepath.Join(root, "backend", "config", "config.yaml")
}

// TestIntegrationMonitorRecordsWithWarehouseAndLogisticsUsers 依赖现有 PostgreSQL 中已注册的 warehouse / logistics 用户及商品数据。
// 运行: COLD_CHAIN_MONITOR_BLOCKCHAIN_TEST=1 go test ./backend/services/ -run TestIntegrationMonitorRecordsWithWarehouseAndLogisticsUsers -count=1
// 链节点可选：未连接时仍校验 blockchain_hash 写入；连接成功时会短暂等待 tx 回写。
func TestIntegrationMonitorRecordsWithWarehouseAndLogisticsUsers(t *testing.T) {
	if os.Getenv("COLD_CHAIN_MONITOR_BLOCKCHAIN_TEST") == "" {
		t.Skip("set COLD_CHAIN_MONITOR_BLOCKCHAIN_TEST=1 to run against your database")
	}
	cfgPath := testRepoConfigPath(t)
	if err := config.LoadConfig(cfgPath); err != nil {
		t.Fatalf("load config %s: %v", cfgPath, err)
	}
	if err := database.InitDB(); err != nil {
		t.Fatalf("init db: %v", err)
	}
	t.Cleanup(func() { _ = database.CloseDB() })

	_ = blockchain.InitFISCOClient()

	var wh models.User
	if err := database.DB.Where("role = ?", models.RoleWarehouse).First(&wh).Error; err != nil {
		t.Fatalf("数据库中需要至少一个 role=warehouse 的用户: %v", err)
	}
	var lg models.User
	if err := database.DB.Where("role = ?", models.RoleLogistics).First(&lg).Error; err != nil {
		t.Fatalf("数据库中需要至少一个 role=logistics 的用户: %v", err)
	}

	var anyProduct models.Product
	if err := database.DB.Where("is_deleted = ?", false).First(&anyProduct).Error; err != nil {
		t.Fatalf("数据库中需要至少一个未删除的商品: %v", err)
	}

	svc := &ProductService{}
	if _, err := svc.CreateTemperatureRecord(anyProduct.TraceID, 2.0, 50, "集成测试冷库", wh.ID, wh.Username); err != nil {
		t.Fatalf("CreateTemperatureRecord: %v", err)
	}
	var lastTemp models.TemperatureRecord
	if err := database.DB.Where("trace_id = ? AND warehouse_id = ?", anyProduct.TraceID, wh.ID).
		Order("id DESC").First(&lastTemp).Error; err != nil {
		t.Fatalf("读取温控记录: %v", err)
	}
	if lastTemp.BlockchainHash == "" {
		t.Fatal("温控记录应写入 blockchain_hash")
	}

	errBadWh := svc.SyncTemperatureRecordBlockchain(lastTemp.ID, wh.ID+99999, wh.Username)
	if errBadWh == nil {
		t.Fatal("SyncTemperatureRecordBlockchain 应对非本人 warehouse_id 返回错误")
	}
	if !errors.Is(errBadWh, ErrProductForbidden) {
		t.Fatalf("SyncTemperatureRecordBlockchain 越权应返回 ErrProductForbidden: %v", errBadWh)
	}

	if lg.CompanyName == "" {
		t.Log("物流账号 company_name 为空，跳过运输节点测试（需与商品 transport_company 匹配）")
	} else {
		var shipProduct models.Product
		err := database.DB.Where("is_deleted = ? AND transport_company = ?", false, lg.CompanyName).First(&shipProduct).Error
		if err != nil {
			t.Logf("无 transport_company=%q 的商品，跳过物流节点测试: %v", lg.CompanyName, err)
		} else {
			if _, err := svc.CreateTransportNode(shipProduct.TraceID, "集成测试节点", "测试港", "in_transit", lg.ID, lg.Username, lg.CompanyName, nil); err != nil {
				if errors.Is(err, ErrProductValidation) {
					t.Skipf("当前追溯码最新节点离开时间晚于当前时刻，触发顺序校验，跳过本次新增节点断言: %v", err)
				}
				t.Fatalf("CreateTransportNode: %v", err)
			}
			var lastNode models.TransportNode
			if err := database.DB.Where("trace_id = ?", shipProduct.TraceID).Order("id DESC").First(&lastNode).Error; err != nil {
				t.Fatalf("读取运输节点: %v", err)
			}
			if lastNode.BlockchainHash == "" {
				t.Fatal("运输节点应写入 blockchain_hash")
			}
			arriveAt := time.Now().Add(90 * time.Minute)
			leaveAt := time.Now().Add(2 * time.Hour)
			if err := svc.UpdateTransportNodeTimes(lastNode.ID, lg.ID, lg.CompanyName, arriveAt, leaveAt, lg.Username); err != nil {
				t.Fatalf("UpdateTransportNodeTimes: %v", err)
			}
			if err := database.DB.First(&lastNode, lastNode.ID).Error; err != nil {
				t.Fatal(err)
			}
			if lastNode.ArrivalTime.IsZero() {
				t.Fatal("到达时间应已写入")
			}
			if lastNode.DepartureTime.IsZero() {
				t.Fatal("离开时间应已写入")
			}
			errBadSync := svc.SyncTransportNodeBlockchain(lastNode.ID, lg.ID+99999, lg.CompanyName, lg.Username)
			if errBadSync == nil {
				t.Fatal("SyncTransportNodeBlockchain 应对错误 logistics_id 返回错误")
			}
			if !errors.Is(errBadSync, ErrProductForbidden) {
				t.Fatalf("SyncTransportNodeBlockchain 越权应返回 ErrProductForbidden: %v", errBadSync)
			}
		}
	}

	if blockchain.IsConnected() {
		deadline := time.Now().Add(8 * time.Second)
		for time.Now().Before(deadline) {
			database.DB.First(&lastTemp, lastTemp.ID)
			if lastTemp.BlockchainTxHash != "" {
				break
			}
			time.Sleep(400 * time.Millisecond)
		}
		if lastTemp.BlockchainTxHash == "" {
			t.Log("链已连接但尚未回写 blockchain_tx_hash，可稍后查库或检查节点；存证哈希已落库")
		}
	}
}
