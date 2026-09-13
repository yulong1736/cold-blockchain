package services

import (
	"cold-chain-trace/backend/blockchain"
	"cold-chain-trace/backend/database"
	"cold-chain-trace/backend/models"
	"cold-chain-trace/backend/utils"
	"log"
	"time"
)

const reconcileBatchSize = 25

// StartPendingProductBlockchainReconcile 周期性将「待上链」商品补发到链上（创建时链未通会留下空 blockchain_tx_hash）
func StartPendingProductBlockchainReconcile(interval time.Duration) {
	if interval <= 0 {
		interval = 15 * time.Second
	}
	go func() {
		reconcileAllPendingEvidence()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			reconcileAllPendingEvidence()
		}
	}()
	log.Printf("blockchain reconcile: started (interval=%s)", interval)
}

func reconcileAllPendingEvidence() {
	reconcilePendingProductsToBlockchain()
	reconcilePendingTemperatureRecordsToBlockchain()
	reconcilePendingTransportNodesToBlockchain()
}

func reconcilePendingProductsToBlockchain() {
	if !blockchain.IsConnected() {
		return
	}
	var pending []models.Product
	if err := database.DB.Where("is_deleted = ?", false).
		Where("(blockchain_tx_hash IS NULL OR blockchain_tx_hash = '')").
		Order("id ASC").
		Limit(reconcileBatchSize).
		Find(&pending).Error; err != nil {
		log.Printf("blockchain reconcile: query pending products: %v", err)
		return
	}
	for i := range pending {
		p := &pending[i]
		dataHash := p.BlockchainHash
		if dataHash == "" {
			dataHash = utils.GenerateProductHash(
				p.ProductID,
				p.BatchNumber,
				p.Origin,
				p.ProductionTime.Format(time.RFC3339),
			)
			if uerr := database.DB.Model(&models.Product{}).Where("id = ?", p.ID).Update("blockchain_hash", dataHash).Error; uerr != nil {
				log.Printf("blockchain reconcile: product %d set hash: %v", p.ID, uerr)
				continue
			}
		}
		op := p.ProducerName
		if op == "" {
			op = "系统补链"
		}
		txHash, err := blockchain.StoreEvidenceNoWait(p.TraceID, dataHash, op)
		if txHash == "" {
			if err != nil {
				log.Printf("blockchain reconcile: product %d trace=%s send failed: %v", p.ID, p.TraceID, err)
			}
			continue
		}
		res := database.DB.Model(&models.Product{}).
			Where("id = ? AND (blockchain_tx_hash IS NULL OR blockchain_tx_hash = '')", p.ID).
			Updates(map[string]interface{}{"blockchain_tx_hash": txHash})
		if res.Error != nil {
			log.Printf("blockchain reconcile: product %d update tx: %v", p.ID, res.Error)
			continue
		}
		if res.RowsAffected > 0 {
			log.Printf("blockchain reconcile: product id=%d trace=%s tx=%s", p.ID, p.TraceID, txHash)
		}
	}
}

func reconcilePendingTemperatureRecordsToBlockchain() {
	if !blockchain.IsConnected() {
		return
	}
	var pending []models.TemperatureRecord
	if err := database.DB.
		Where("(blockchain_tx_hash IS NULL OR blockchain_tx_hash = '')").
		Order("id ASC").
		Limit(reconcileBatchSize).
		Find(&pending).Error; err != nil {
		log.Printf("blockchain reconcile: query pending temperature_records: %v", err)
		return
	}
	for i := range pending {
		r := &pending[i]
		dataHash := r.BlockchainHash
		if dataHash == "" {
			dataHash = utils.GenerateTemperatureRecordHash(
				r.ID, r.TraceID, r.Temperature, r.Humidity, r.Location,
				r.RecordTime, r.WarehouseID, r.WarehouseName, r.IsAbnormal,
			)
			if uerr := database.DB.Model(&models.TemperatureRecord{}).Where("id = ?", r.ID).Update("blockchain_hash", dataHash).Error; uerr != nil {
				log.Printf("blockchain reconcile: temperature_record %d set hash: %v", r.ID, uerr)
				continue
			}
		}
		op := r.WarehouseName
		if op == "" {
			op = "系统补链"
		}
		txHash, err := blockchain.StoreEvidenceNoWaitWithKind(blockchain.EvidenceKindTemperature, r.TraceID, dataHash, op)
		if txHash == "" {
			if err != nil {
				log.Printf("blockchain reconcile: temperature_record %d trace=%s send failed: %v", r.ID, r.TraceID, err)
			}
			continue
		}
		res := database.DB.Model(&models.TemperatureRecord{}).
			Where("id = ? AND (blockchain_tx_hash IS NULL OR blockchain_tx_hash = '')", r.ID).
			Updates(map[string]interface{}{"blockchain_tx_hash": txHash})
		if res.Error != nil {
			log.Printf("blockchain reconcile: temperature_record %d update tx: %v", r.ID, res.Error)
			continue
		}
		if res.RowsAffected > 0 {
			log.Printf("blockchain reconcile: temperature_record id=%d trace=%s tx=%s", r.ID, r.TraceID, txHash)
		}
	}
}

func reconcilePendingTransportNodesToBlockchain() {
	if !blockchain.IsConnected() {
		return
	}
	var pending []models.TransportNode
	if err := database.DB.
		Where("(blockchain_tx_hash IS NULL OR blockchain_tx_hash = '')").
		Order("id ASC").
		Limit(reconcileBatchSize).
		Find(&pending).Error; err != nil {
		log.Printf("blockchain reconcile: query pending transport_nodes: %v", err)
		return
	}
	for i := range pending {
		n := &pending[i]
		dataHash := n.BlockchainHash
		if dataHash == "" {
			dataHash = utils.GenerateTransportNodeHash(
				n.ID, n.TraceID, n.NodeName, n.Location, n.ArrivalTime, n.DepartureTime,
				n.LogisticsID, n.LogisticsName, n.Status,
			)
			if uerr := database.DB.Model(&models.TransportNode{}).Where("id = ?", n.ID).Update("blockchain_hash", dataHash).Error; uerr != nil {
				log.Printf("blockchain reconcile: transport_node %d set hash: %v", n.ID, uerr)
				continue
			}
		}
		op := n.LogisticsName
		if op == "" {
			op = "系统补链"
		}
		txHash, err := blockchain.StoreEvidenceNoWaitWithKind(blockchain.EvidenceKindLogistics, n.TraceID, dataHash, op)
		if txHash == "" {
			if err != nil {
				log.Printf("blockchain reconcile: transport_node %d trace=%s send failed: %v", n.ID, n.TraceID, err)
			}
			continue
		}
		res := database.DB.Model(&models.TransportNode{}).
			Where("id = ? AND (blockchain_tx_hash IS NULL OR blockchain_tx_hash = '')", n.ID).
			Updates(map[string]interface{}{"blockchain_tx_hash": txHash})
		if res.Error != nil {
			log.Printf("blockchain reconcile: transport_node %d update tx: %v", n.ID, res.Error)
			continue
		}
		if res.RowsAffected > 0 {
			log.Printf("blockchain reconcile: transport_node id=%d trace=%s tx=%s", n.ID, n.TraceID, txHash)
		}
	}
}
