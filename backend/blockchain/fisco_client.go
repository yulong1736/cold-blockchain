package blockchain

import (
	"cold-chain-trace/backend/config"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// 链上交易 data 字段中的业务类型（与合约 appendAuxEvidence 的 kind 语义对齐）
const (
	EvidenceKindProduct     = "product"
	EvidenceKindTemperature = "temperature"
	EvidenceKindLogistics   = "logistics"
)

var (
	FISCOClient       *ethclient.Client
	senderKey         *ecdsa.PrivateKey
	senderAddr        common.Address
	traceEvidenceAddr common.Address
	traceEvidenceABI  abi.ABI
	contractMode      bool
	// effectiveChainID 来自 eth_chainId，与节点一致；签名错误时易出现重复/回放类错误
	effectiveChainID *big.Int
	storeEvidenceMu  sync.Mutex
	initMu           sync.Mutex
)

// ErrReceiptTimeout 交易已发送但在阈值内未拿到回执
var ErrReceiptTimeout = errors.New("blockchain receipt timeout")

// IsConnected 区块链是否已连接
func IsConnected() bool {
	if err := ensureConnected(2 * time.Second); err != nil {
		return false
	}
	return true
}

func normalizeHexPrivateKey(k string) string {
	k = strings.TrimSpace(k)
	k = strings.TrimPrefix(k, "0x")
	return k
}

// InitFISCOClient 初始化FISCO BCOS客户端（FISCO 3.x Web3 RPC）
func InitFISCOClient() error {
	initMu.Lock()
	defer initMu.Unlock()

	rpcURL := strings.TrimSpace(config.AppConfig.Blockchain.FISCOBCOS.NodeURL)
	if rpcURL == "" {
		return fmt.Errorf("blockchain node_url is empty")
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return fmt.Errorf("failed to connect to FISCO BCOS: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := client.BlockNumber(ctx); err != nil {
		return fmt.Errorf("connected but failed to query block number: %w", err)
	}

	cid, err := client.ChainID(ctx)
	if err != nil || cid == nil || cid.Sign() == 0 {
		effectiveChainID = big.NewInt(int64(config.AppConfig.Blockchain.FISCOBCOS.ChainID))
		if effectiveChainID.Sign() <= 0 {
			effectiveChainID = big.NewInt(1)
		}
	} else {
		effectiveChainID = new(big.Int).Set(cid)
	}

	privHex := normalizeHexPrivateKey(config.AppConfig.Blockchain.FISCOBCOS.PrivateKey)
	if privHex == "" {
		return fmt.Errorf("blockchain private_key is empty")
	}
	decodedKey, err := hex.DecodeString(privHex)
	if err != nil {
		return fmt.Errorf("invalid private_key format: %w", err)
	}
	key, err := crypto.ToECDSA(decodedKey)
	if err != nil {
		return fmt.Errorf("invalid private_key: %w", err)
	}

	senderKey = key
	senderAddr = crypto.PubkeyToAddress(key.PublicKey)
	FISCOClient = client
	if err := initTraceEvidenceContract(); err != nil {
		return err
	}
	return nil
}

func pingClient(timeout time.Duration) error {
	if FISCOClient == nil || senderKey == nil {
		return fmt.Errorf("blockchain client not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_, err := FISCOClient.BlockNumber(ctx)
	return err
}

func ensureConnected(timeout time.Duration) error {
	if err := pingClient(timeout); err == nil {
		return nil
	}

	if err := InitFISCOClient(); err != nil {
		return err
	}
	return pingClient(timeout)
}

func sendTx(to common.Address, txData []byte) (common.Hash, error) {
	if err := ensureConnected(3 * time.Second); err != nil {
		return common.Hash{}, err
	}

	storeEvidenceMu.Lock()
	defer storeEvidenceMu.Unlock()

	ctxSend, cancelSend := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelSend()

	nonce, err := FISCOClient.PendingNonceAt(ctxSend, senderAddr)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get nonce: %w", err)
	}

	gasPrice, err := FISCOClient.SuggestGasPrice(ctxSend)
	if err != nil || gasPrice == nil || gasPrice.Sign() <= 0 {
		gasPrice = big.NewInt(1_000_000_000) // 1 gwei fallback
	}

	callMsg := ethereum.CallMsg{
		From: senderAddr,
		To:   &to,
		Data: txData,
	}
	gasLimit, err := FISCOClient.EstimateGas(ctxSend, callMsg)
	if err != nil || gasLimit == 0 {
		gasLimit = 300000
	}
	const gasFloor = 250000
	if gasLimit < gasFloor {
		gasLimit = gasFloor
	}

	chainID := effectiveChainID
	if chainID == nil || chainID.Sign() == 0 {
		chainID = big.NewInt(int64(config.AppConfig.Blockchain.FISCOBCOS.ChainID))
		if chainID.Sign() <= 0 {
			chainID = big.NewInt(1)
		}
	}

	tx := types.NewTransaction(nonce, to, big.NewInt(0), gasLimit, gasPrice, txData)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), senderKey)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to sign tx: %w", err)
	}

	sendErr := FISCOClient.SendTransaction(ctxSend, signedTx)
	if sendErr != nil {
		// FISCO 3.x 部分节点对 EIP-155 交易校验与 geth 不完全一致，回退为 Homestead 签名再发一笔
		tx2 := types.NewTransaction(nonce, to, big.NewInt(0), gasLimit, gasPrice, txData)
		signedTx2, hsErr := types.SignTx(tx2, types.HomesteadSigner{}, senderKey)
		if hsErr != nil {
			return common.Hash{}, fmt.Errorf("failed to send tx: %w (homestead sign: %v)", sendErr, hsErr)
		}
		if err2 := FISCOClient.SendTransaction(ctxSend, signedTx2); err2 != nil {
			return common.Hash{}, fmt.Errorf("failed to send tx (eip155: %v; homestead: %w)", sendErr, err2)
		}
		signedTx = signedTx2
	}
	return signedTx.Hash(), nil
}

func sendEvidenceTx(kind, traceID, dataHash, operator string) (common.Hash, error) {
	if kind == "" {
		kind = EvidenceKindProduct
	}

	// 通过普通交易 data 字段做轻量存证；rnd + 纳秒时间戳避免相同内容被节点判为重复交易
	var rnd [8]byte
	_, _ = rand.Read(rnd[:])
	payload := fmt.Sprintf("kind=%s;trace_id=%s;hash=%s;operator=%s;ts=%d;rnd=%x",
		kind, traceID, dataHash, operator, time.Now().UnixNano(), rnd[:])
	txData := []byte(payload)
	return sendTx(senderAddr, txData)
}

// StoreEvidence 存储商品主存证（kind=product）
func StoreEvidence(traceID, dataHash, operator string) (string, *big.Int, error) {
	return StoreEvidenceWithKind(EvidenceKindProduct, traceID, dataHash, operator)
}

// StoreEvidenceWithKind 按类型存证（product / temperature / logistics）
func StoreEvidenceWithKind(kind, traceID, dataHash, operator string) (string, *big.Int, error) {
	var txHash common.Hash
	var err error
	if useContractMode() {
		switch kind {
		case EvidenceKindProduct:
			txHash, err = sendContractTx("storeEvidence", traceID, dataHash, operator)
		case EvidenceKindTemperature, EvidenceKindLogistics:
			txHash, err = sendContractTx("appendAuxEvidence", traceID, kind, dataHash, operator)
		default:
			txHash, err = sendContractTx("appendAuxEvidence", traceID, kind, dataHash, operator)
		}
	} else {
		txHash, err = sendEvidenceTx(kind, traceID, dataHash, operator)
	}
	if err != nil {
		return "", nil, err
	}

	ctxWait, cancelWait := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelWait()
	blockNumber, err := waitTxMined(ctxWait, txHash)
	if err != nil {
		return txHash.Hex(), nil, err
	}
	return txHash.Hex(), blockNumber, nil
}

// StoreEvidenceNoWait 异步商品主存证
func StoreEvidenceNoWait(traceID, dataHash, operator string) (string, error) {
	return StoreEvidenceNoWaitWithKind(EvidenceKindProduct, traceID, dataHash, operator)
}

// StoreEvidenceNoWaitWithKind 异步按类型存证
func StoreEvidenceNoWaitWithKind(kind, traceID, dataHash, operator string) (string, error) {
	var txHash common.Hash
	var err error
	if useContractMode() {
		switch kind {
		case EvidenceKindProduct:
			txHash, err = sendContractTx("storeEvidence", traceID, dataHash, operator)
		case EvidenceKindTemperature, EvidenceKindLogistics:
			txHash, err = sendContractTx("appendAuxEvidence", traceID, kind, dataHash, operator)
		default:
			txHash, err = sendContractTx("appendAuxEvidence", traceID, kind, dataHash, operator)
		}
	} else {
		txHash, err = sendEvidenceTx(kind, traceID, dataHash, operator)
	}
	if err != nil {
		return "", err
	}
	return txHash.Hex(), nil
}

// QueryEvidence 查询存证信息
func QueryEvidence(traceID string) (map[string]interface{}, error) {
	if err := ensureConnected(2 * time.Second); err != nil {
		return nil, err
	}
	result := make(map[string]interface{})
	result["trace_id"] = traceID
	if !useContractMode() {
		result["exists"] = true
		return result, nil
	}

	exists, err := callContractBool("verifyTraceID", traceID)
	if err != nil {
		return nil, err
	}
	result["exists"] = exists
	if !exists {
		return result, nil
	}
	values, err := callContract("queryEvidence", traceID)
	if err != nil {
		return nil, err
	}
	if len(values) == 6 {
		result["trace_id"] = values[0].(string)
		result["hash"] = values[1].(string)
		result["data_hash"] = values[1].(string)
		result["operator"] = values[2].(string)
		if ts, ok := values[3].(*big.Int); ok {
			result["timestamp"] = ts.Uint64()
		}
		if bn, ok := values[4].(*big.Int); ok {
			result["block_number"] = bn.Uint64()
		}
		result["is_valid"] = values[5].(bool)
	}
	return result, nil
}

// QueryHistory 查询操作历史
func QueryHistory(traceID string) ([]map[string]interface{}, error) {
	if err := ensureConnected(2 * time.Second); err != nil {
		return nil, err
	}
	history := make([]map[string]interface{}, 0)
	if !useContractMode() {
		return history, nil
	}
	count, err := callContractBigInt("getHistoryCount", traceID)
	if err != nil {
		return nil, err
	}
	for i := int64(0); i < count.Int64(); i++ {
		values, err := callContract("getHistoryRecord", traceID, big.NewInt(i))
		if err != nil {
			return nil, err
		}
		if len(values) != 7 {
			continue
		}
		row := map[string]interface{}{
			"trace_id":       values[0].(string),
			"operation_type": values[1].(string),
			"old_hash":       values[2].(string),
			"new_hash":       values[3].(string),
			"operator":       values[4].(string),
			"change_details": values[6].(string),
		}
		if ts, ok := values[5].(*big.Int); ok {
			row["timestamp"] = ts.Uint64()
		}
		history = append(history, row)
	}
	return history, nil
}

// GetBlockNumber 获取当前区块号
func GetBlockNumber(ctx context.Context) (*big.Int, error) {
	if err := ensureConnected(2 * time.Second); err != nil {
		return nil, err
	}

	blockNumber, err := FISCOClient.BlockNumber(ctx)
	if err != nil {
		return nil, err
	}

	return new(big.Int).SetUint64(blockNumber), nil
}

func waitTxMined(ctx context.Context, txHash common.Hash) (*big.Int, error) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		receipt, err := FISCOClient.TransactionReceipt(ctx, txHash)
		if err == nil && receipt != nil {
			if receipt.Status == types.ReceiptStatusFailed {
				return nil, fmt.Errorf("transaction reverted on chain")
			}
			return new(big.Int).SetUint64(receipt.BlockNumber.Uint64()), nil
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: transaction sent but receipt wait timeout: %v", ErrReceiptTimeout, ctx.Err())
		case <-ticker.C:
		}
	}
}
