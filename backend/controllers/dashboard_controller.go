package controllers

import (
	"cold-chain-trace/backend/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardController struct {
	dashboardService *services.DashboardService
}

func NewDashboardController() *DashboardController {
	return &DashboardController{dashboardService: &services.DashboardService{}}
}

// GetStats 仪表盘统计（生产商仅看自家数据，物流仅看本公司，监管看全部）
func (dc *DashboardController) GetStats(c *gin.Context) {
	user, err := getRequestUser(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}
	stats, err := dc.dashboardService.GetDashboardStats(user.ID, user.Role, user.CompanyName, user.Username)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// GetTemperatureTrend 温度趋势（生产商仅看自家追溯码，物流仅看本公司，监管看全部）
func (dc *DashboardController) GetTemperatureTrend(c *gin.Context) {
	user, err := getRequestUser(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}
	limit := parsePositiveIntQuery(c, "limit", 200, 0)
	records, err := dc.dashboardService.GetTemperatureTrend(user.ID, user.Role, user.CompanyName, user.Username, limit)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": records})
}

// GetTransportMap 运输节点（生产商仅看自家追溯码的物流轨迹，物流仅看本公司，监管看全部）
func (dc *DashboardController) GetTransportMap(c *gin.Context) {
	user, err := getRequestUser(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}
	limit := parsePositiveIntQuery(c, "limit", 500, 0)
	nodes, err := dc.dashboardService.GetTransportNodesForMap(user.ID, user.Role, user.CompanyName, user.Username, limit)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": nodes})
}

// GetAlerts 告警列表（仓储只看自己的，物流只看本公司商品的，监管看全部）
func (dc *DashboardController) GetAlerts(c *gin.Context) {
	user, err := getRequestUser(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}
	limit := parsePositiveIntQuery(c, "limit", 100, 0)
	list, err := dc.dashboardService.GetAlerts(limit, user.Role, user.ID, user.CompanyName)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

// GetAllAlerts 监管专用：查看全部告警
func (dc *DashboardController) GetAllAlerts(c *gin.Context) {
	limit := parsePositiveIntQuery(c, "limit", 200, 0)
	list, err := dc.dashboardService.GetAlerts(limit, "regulator", 0, "")
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}
