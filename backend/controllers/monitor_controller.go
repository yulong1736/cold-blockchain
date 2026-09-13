package controllers

import (
	"cold-chain-trace/backend/blockchain"
	"cold-chain-trace/backend/services"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type MonitorController struct {
	productService *services.ProductService
}

func NewMonitorController() *MonitorController {
	return &MonitorController{
		productService: &services.ProductService{},
	}
}

// CreateTemperatureRecord 仓储添加温控记录
func (mc *MonitorController) CreateTemperatureRecord(c *gin.Context) {
	var req struct {
		TraceID     string  `json:"trace_id" binding:"required"`
		Temperature float64 `json:"temperature" binding:"required"`
		Humidity    float64 `json:"humidity"`
		Location    string  `json:"location"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	user, err := getRequestUser(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}

	id, err := mc.productService.CreateTemperatureRecord(req.TraceID, req.Temperature, req.Humidity, req.Location, user.ID, user.Username)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "温控记录创建成功", "data": gin.H{"id": id}})
}

// ListTemperatureRecords 仓储查看自己的温控记录
func (mc *MonitorController) ListTemperatureRecords(c *gin.Context) {
	user, err := getRequestUser(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}

	records, err := mc.productService.ListTemperatureRecordsByWarehouse(user.ID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": records})
}

// CreateTransportNode 物流添加运输节点
func (mc *MonitorController) CreateTransportNode(c *gin.Context) {
	var req struct {
		TraceID       string `json:"trace_id" binding:"required"`
		NodeName      string `json:"node_name" binding:"required"`
		Location      string `json:"location"`
		Status        string `json:"status"`
		DepartureTime string `json:"departure_time"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	var departurePtr *time.Time
	if s := strings.TrimSpace(req.DepartureTime); s != "" {
		var parsed time.Time
		var perr error
		for _, layout := range []string{time.RFC3339, time.RFC3339Nano, "2006-01-02 15:04:05", "2006-01-02T15:04:05"} {
			parsed, perr = time.Parse(layout, s)
			if perr == nil {
				departurePtr = &parsed
				break
			}
		}
		if departurePtr == nil {
			respondErrorMessage(c, http.StatusBadRequest, "departure_time 格式无效，请使用 RFC3339 或 yyyy-MM-dd HH:mm:ss")
			return
		}
	}

	user, err := getRequestUser(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}

	id, err := mc.productService.CreateTransportNode(req.TraceID, req.NodeName, req.Location, req.Status, user.ID, user.Username, user.CompanyName, departurePtr)
	if err != nil {
		if errors.Is(err, services.ErrProductNotFound) {
			respondError(c, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, services.ErrProductForbidden) {
			respondError(c, http.StatusForbidden, err)
			return
		}
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "运输节点创建成功", "data": gin.H{"id": id}})
}

// SyncTemperatureBlockchain 手动同步温控记录上链
func (mc *MonitorController) SyncTemperatureBlockchain(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondErrorMessage(c, http.StatusBadRequest, "无效的记录 id")
		return
	}
	user, err := getRequestUser(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}
	if err := mc.productService.SyncTemperatureRecordBlockchain(uint(id64), user.ID, user.Username); err != nil {
		if errors.Is(err, blockchain.ErrReceiptTimeout) {
			respondError(c, http.StatusRequestTimeout, err)
			return
		}
		if errors.Is(err, services.ErrProductNotFound) {
			respondError(c, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, services.ErrProductForbidden) {
			respondError(c, http.StatusForbidden, err)
			return
		}
		if errors.Is(err, services.ErrProductValidation) {
			respondError(c, http.StatusBadRequest, err)
			return
		}
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "链上同步已提交"})
}

// SyncTransportBlockchain 手动同步运输节点上链
func (mc *MonitorController) SyncTransportBlockchain(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondErrorMessage(c, http.StatusBadRequest, "无效的节点 id")
		return
	}
	user, err := getRequestUser(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}
	if err := mc.productService.SyncTransportNodeBlockchain(uint(id64), user.ID, user.CompanyName, user.Username); err != nil {
		if errors.Is(err, blockchain.ErrReceiptTimeout) {
			respondError(c, http.StatusRequestTimeout, err)
			return
		}
		if errors.Is(err, services.ErrProductNotFound) {
			respondError(c, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, services.ErrProductForbidden) {
			respondError(c, http.StatusForbidden, err)
			return
		}
		if errors.Is(err, services.ErrProductValidation) {
			respondError(c, http.StatusBadRequest, err)
			return
		}
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "链上同步已提交"})
}

// UpdateTransportNodeTimes 更新运输节点到达/离开时间
func (mc *MonitorController) UpdateTransportNodeTimes(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondErrorMessage(c, http.StatusBadRequest, "无效的节点 id")
		return
	}
	var req struct {
		ArrivalTime   string `json:"arrival_time" binding:"required"`
		DepartureTime string `json:"departure_time" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}
	parseTime := func(raw string) (time.Time, bool) {
		s := strings.TrimSpace(raw)
		var t time.Time
		var ok bool
		for _, layout := range []string{time.RFC3339, time.RFC3339Nano, "2006-01-02 15:04:05", "2006-01-02T15:04:05"} {
			tt, perr := time.Parse(layout, s)
			if perr == nil {
				t = tt
				ok = true
				break
			}
		}
		return t, ok
	}

	arr, ok := parseTime(req.ArrivalTime)
	if !ok {
		respondErrorMessage(c, http.StatusBadRequest, "arrival_time 格式无效")
		return
	}
	dep, ok := parseTime(req.DepartureTime)
	if !ok {
		respondErrorMessage(c, http.StatusBadRequest, "departure_time 格式无效")
		return
	}
	if dep.Before(arr) {
		respondErrorMessage(c, http.StatusBadRequest, "离开时间不能早于到达时间")
		return
	}

	user, err := getRequestUser(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}
	if err := mc.productService.UpdateTransportNodeTimes(uint(id64), user.ID, user.CompanyName, arr, dep, user.Username); err != nil {
		if errors.Is(err, services.ErrProductNotFound) {
			respondError(c, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, services.ErrProductForbidden) {
			respondError(c, http.StatusForbidden, err)
			return
		}
		if errors.Is(err, services.ErrProductValidation) {
			respondError(c, http.StatusBadRequest, err)
			return
		}
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "到达/离开时间已更新"})
}

// ListTransportNodes 物流查看自己公司的运输节点
func (mc *MonitorController) ListTransportNodes(c *gin.Context) {
	user, err := getRequestUser(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}

	nodes, err := mc.productService.ListTransportNodesByLogistics(user.ID, user.CompanyName)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": nodes})
}

// ListMyProducts 物流账号获取本公司关联的商品列表（供添加节点时选追溯码）
func (mc *MonitorController) ListMyProducts(c *gin.Context) {
	user, err := getRequestUser(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}

	products, err := mc.productService.ListProductsByLogisticsCompany(user.CompanyName)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": products})
}
