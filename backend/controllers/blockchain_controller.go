package controllers

import (
	"cold-chain-trace/backend/blockchain"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type BlockchainController struct{}

func NewBlockchainController() *BlockchainController {
	return &BlockchainController{}
}

// GetStatus 链是否可达及当前块高（供前端轮询）
func (bc *BlockchainController) GetStatus(c *gin.Context) {
	ok := blockchain.IsConnected()
	out := gin.H{"connected": ok}
	if ok {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if bn, err := blockchain.GetBlockNumber(ctx); err == nil {
			out["block_number"] = fmt.Sprintf("0x%x", bn)
		}
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}
