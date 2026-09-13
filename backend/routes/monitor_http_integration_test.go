package routes

import (
	"cold-chain-trace/backend/config"
	"cold-chain-trace/backend/database"
	"cold-chain-trace/backend/models"
	"cold-chain-trace/backend/utils"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func repoRootConfigPath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	routesDir := filepath.Dir(file)
	root := filepath.Join(routesDir, "..", "..")
	return filepath.Join(root, "backend", "config", "config.yaml")
}

// TestIntegrationMonitorAPIsBlockchainFields 需 PostgreSQL 与已有 warehouse / logistics 用户；与 services 集成测试共用环境变量。
func TestIntegrationMonitorAPIsBlockchainFields(t *testing.T) {
	if os.Getenv("COLD_CHAIN_MONITOR_BLOCKCHAIN_TEST") == "" {
		t.Skip("set COLD_CHAIN_MONITOR_BLOCKCHAIN_TEST=1")
	}
	gin.SetMode(gin.TestMode)
	if err := config.LoadConfig(repoRootConfigPath(t)); err != nil {
		t.Fatalf("config: %v", err)
	}
	if err := database.InitDB(); err != nil {
		t.Fatalf("db: %v", err)
	}
	t.Cleanup(func() { _ = database.CloseDB() })

	var wh models.User
	if err := database.DB.Where("role = ?", models.RoleWarehouse).First(&wh).Error; err != nil {
		t.Fatalf("warehouse user: %v", err)
	}
	var lg models.User
	if err := database.DB.Where("role = ?", models.RoleLogistics).First(&lg).Error; err != nil {
		t.Fatalf("logistics user: %v", err)
	}

	whTok, err := utils.GenerateToken(wh.ID, wh.Username, string(models.RoleWarehouse), wh.CompanyName)
	if err != nil {
		t.Fatal(err)
	}
	lgTok, err := utils.GenerateToken(lg.ID, lg.Username, string(models.RoleLogistics), lg.CompanyName)
	if err != nil {
		t.Fatal(err)
	}

	engine := SetupRoutes()

	t.Run("GET_temperature_data_has_blockchain_keys", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/temperature", nil)
		req.Header.Set("Authorization", "Bearer "+whTok)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("status %d body %s", w.Code, w.Body.String())
		}
		var wrap struct {
			Data []json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &wrap); err != nil {
			t.Fatal(err)
		}
		if len(wrap.Data) == 0 {
			t.Skip("no temperature rows to assert JSON shape")
		}
		var row map[string]interface{}
		if err := json.Unmarshal(wrap.Data[0], &row); err != nil {
			t.Fatal(err)
		}
		for _, k := range []string{"blockchain_hash", "blockchain_tx_hash"} {
			if _, ok := row[k]; !ok {
				t.Fatalf("missing key %q in row: %v", k, row)
			}
		}
	})

	t.Run("GET_transport_data_has_blockchain_keys", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/transport", nil)
		req.Header.Set("Authorization", "Bearer "+lgTok)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("status %d body %s", w.Code, w.Body.String())
		}
		var wrap struct {
			Data []json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &wrap); err != nil {
			t.Fatal(err)
		}
		if len(wrap.Data) == 0 {
			t.Skip("no transport rows to assert JSON shape")
		}
		var row map[string]interface{}
		if err := json.Unmarshal(wrap.Data[0], &row); err != nil {
			t.Fatal(err)
		}
		for _, k := range []string{"blockchain_hash", "blockchain_tx_hash"} {
			if _, ok := row[k]; !ok {
				t.Fatalf("missing key %q in row: %v", k, row)
			}
		}
	})

	t.Run("GET_blockchain_status_warehouse_allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/blockchain/status", nil)
		req.Header.Set("Authorization", "Bearer "+whTok)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("warehouse GET /blockchain/status want 200, got %d: %s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), `"connected"`) {
			t.Fatalf("unexpected body: %s", w.Body.String())
		}
	})

	t.Run("POST_temperature_sync_not_found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/temperature/999999999/sync-blockchain", nil)
		req.Header.Set("Authorization", "Bearer "+whTok)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404 for missing record, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("POST_transport_sync_not_found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/transport/999999999/sync-blockchain", nil)
		req.Header.Set("Authorization", "Bearer "+lgTok)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("want 404 for missing node, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("PUT_transport_times", func(t *testing.T) {
		var tn models.TransportNode
		if err := database.DB.Where("logistics_id = ?", lg.ID).Order("id DESC").First(&tn).Error; err != nil {
			t.Skip("no transport node for logistics user")
		}
		body := `{"arrival_time":"2026-04-08T13:30:00+08:00","departure_time":"2026-04-08T14:30:00+08:00"}`
		req := httptest.NewRequest(http.MethodPut, "/api/transport/"+strconv.FormatUint(uint64(tn.ID), 10), strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+lgTok)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("PUT times want 200, got %d: %s", w.Code, w.Body.String())
		}
		var reloaded models.TransportNode
		if err := database.DB.First(&reloaded, tn.ID).Error; err != nil {
			t.Fatal(err)
		}
		if reloaded.ArrivalTime.IsZero() {
			t.Fatal("expected arrival_time persisted")
		}
		if reloaded.DepartureTime.IsZero() {
			t.Fatal("expected departure_time persisted")
		}
	})
}
