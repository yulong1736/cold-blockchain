package controllers

import (
	"cold-chain-trace/backend/blockchain"
	"cold-chain-trace/backend/models"
	"cold-chain-trace/backend/services"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductController struct {
	productService *services.ProductService
}

func NewProductController() *ProductController {
	return &ProductController{
		productService: &services.ProductService{},
	}
}

// CreateProduct 创建商品信息
// @Summary 创建商品信息
// @Description 上传商品信息并上链存证
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param product body models.Product true "商品信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/products [post]
func (pc *ProductController) CreateProduct(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	user, err := getRequestUser(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}

	// Attach producer info so the producer can later list/manage their own products.
	product.ProducerID = user.ID
	product.ProducerName = user.Username

	if err := pc.productService.CreateProduct(&product, user.ID, user.Username); err != nil {
		if errors.Is(err, services.ErrProductValidation) {
			respondError(c, http.StatusBadRequest, err)
			return
		}
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "商品创建成功",
		"data":    product,
	})
}

// SyncProductBlockchain 将当前商品数据再次写入区块链并更新 tx_hash
func (pc *ProductController) SyncProductBlockchain(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondErrorMessage(c, http.StatusBadRequest, "无效的商品 id")
		return
	}
	user, err := getRequestUser(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}
	if err := pc.productService.SyncProductBlockchain(uint(id64), user.ID, user.Username); err != nil {
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

// SyncProductBlockchainByTraceID 按追溯码同步上链
func (pc *ProductController) SyncProductBlockchainByTraceID(c *gin.Context) {
	traceID := c.Param("trace_id")
	if traceID == "" {
		respondErrorMessage(c, http.StatusBadRequest, "追溯码不能为空")
		return
	}
	user, err := getRequestUser(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}
	if err := pc.productService.SyncProductBlockchainByTraceID(traceID, user.ID, user.Username); err != nil {
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

// GetMyProducts 获取当前登录生产商的商品列表
// @Summary 获取商品列表
// @Description 获取当前登录生产商创建的商品列表
// @Tags 商品管理
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/products [get]
func (pc *ProductController) GetMyProducts(c *gin.Context) {
	user, err := getRequestUser(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}

	products, err := pc.productService.ListProductsByProducer(user.ID, user.Username)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": products})
}

// GetProductByTraceID 根据追溯码查询商品
// @Summary 查询商品信息
// @Description 通过追溯码查询商品详细信息（链上+链下）
// @Tags 商品管理
// @Produce json
// @Param trace_id path string true "追溯码"
// @Success 200 {object} map[string]interface{}
// @Router /api/products/trace/{trace_id} [get]
func (pc *ProductController) GetProductByTraceID(c *gin.Context) {
	traceID := c.Param("trace_id")

	product, err := pc.productService.GetProductByTraceID(traceID)
	if err != nil {
		if errors.Is(err, services.ErrProductNotFound) {
			respondError(c, http.StatusNotFound, err)
			return
		}
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	// 获取温度记录
	temperatureRecords, _ := pc.productService.GetTemperatureRecords(traceID)
	// 获取运输节点记录
	transportNodes, _ := pc.productService.GetTransportNodesByTraceID(traceID)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"product":             product,
			"temperature_records": temperatureRecords,
			"transport_nodes":     transportNodes,
		},
	})
}

// UpdateProduct 更新商品信息
// @Summary 更新商品信息
// @Description 更新商品信息并记录到链上
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param id path int true "商品ID"
// @Param updates body map[string]interface{} true "更新字段"
// @Success 200 {object} map[string]interface{}
// @Router /api/products/{id} [put]
func (pc *ProductController) UpdateProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondErrorMessage(c, http.StatusBadRequest, "invalid product id")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	user, err := getRequestUser(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}

	if err := pc.productService.UpdateProduct(uint(id), updates, user.ID, user.Username); err != nil {
		if errors.Is(err, services.ErrProductValidation) {
			respondError(c, http.StatusBadRequest, err)
			return
		}
		if errors.Is(err, services.ErrProductForbidden) {
			respondError(c, http.StatusForbidden, err)
			return
		}
		if errors.Is(err, services.ErrProductNotFound) {
			respondError(c, http.StatusNotFound, err)
			return
		}
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "商品更新成功"})
}

// DeleteProduct 删除商品（逻辑删除）
// @Summary 删除商品
// @Description 逻辑删除商品并记录到链上
// @Tags 商品管理
// @Produce json
// @Param id path int true "商品ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/products/{id} [delete]
func (pc *ProductController) DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respondErrorMessage(c, http.StatusBadRequest, "invalid product id")
		return
	}

	user, err := getRequestUser(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}

	if err := pc.productService.DeleteProduct(uint(id), user.ID, user.Username); err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "商品删除成功"})
}

// GetProductHistory 获取商品操作历史
// @Summary 获取操作历史
// @Description 获取商品的所有操作历史记录
// @Tags 商品管理
// @Produce json
// @Param trace_id path string true "追溯码"
// @Success 200 {object} map[string]interface{}
// @Router /api/products/history/{trace_id} [get]
func (pc *ProductController) GetProductHistory(c *gin.Context) {
	traceID := c.Param("trace_id")

	histories, err := pc.productService.GetProductHistory(traceID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": histories})
}

// GetMyPurchasedProducts 消费者：查询自己购买的商品
func (pc *ProductController) GetMyPurchasedProducts(c *gin.Context) {
	user, err := getRequestUser(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}
	products, err := pc.productService.ListProductsByConsumerID(user.ID, user.Username)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": products})
}

// GetMyProductByTraceID 消费者：仅可查询自己购买商品的追溯信息
func (pc *ProductController) GetMyProductByTraceID(c *gin.Context) {
	traceID := c.Param("trace_id")
	user, err := getRequestUser(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}

	product, err := pc.productService.GetProductByTraceIDForConsumer(traceID, user.ID, user.Username)
	if err != nil {
		if errors.Is(err, services.ErrProductForbidden) {
			respondError(c, http.StatusForbidden, err)
			return
		}
		if errors.Is(err, services.ErrProductNotFound) {
			respondError(c, http.StatusNotFound, err)
			return
		}
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	temperatureRecords, _ := pc.productService.GetTemperatureRecords(traceID)
	transportNodes, _ := pc.productService.GetTransportNodesByTraceID(traceID)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"product":             product,
			"temperature_records": temperatureRecords,
			"transport_nodes":     transportNodes,
		},
	})
}

// GetMyProductHistory 消费者：仅可查看自己购买商品的操作历史
func (pc *ProductController) GetMyProductHistory(c *gin.Context) {
	traceID := c.Param("trace_id")
	user, err := getRequestUser(c)
	if err != nil {
		respondError(c, http.StatusUnauthorized, err)
		return
	}

	if _, err := pc.productService.GetProductByTraceIDForConsumer(traceID, user.ID, user.Username); err != nil {
		if errors.Is(err, services.ErrProductForbidden) {
			respondError(c, http.StatusForbidden, err)
			return
		}
		if errors.Is(err, services.ErrProductNotFound) {
			respondError(c, http.StatusNotFound, err)
			return
		}
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	histories, err := pc.productService.GetProductHistory(traceID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": histories})
}

// GetAuditLogs 监管端：审计日志（谁在什么时候做了什么操作）
func (pc *ProductController) GetAuditLogs(c *gin.Context) {
	traceID := c.Query("trace_id")
	limit := parsePositiveIntQuery(c, "limit", 500, 1000)
	histories, err := pc.productService.GetAuditLogs(traceID, limit)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": histories})
}
