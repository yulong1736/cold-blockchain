package blockchain

import (
	"cold-chain-trace/backend/config"
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

const traceEvidenceABIJSON = `[
  {"inputs":[{"internalType":"string","name":"_traceID","type":"string"},{"internalType":"string","name":"_dataHash","type":"string"},{"internalType":"string","name":"_operator","type":"string"}],"name":"storeEvidence","outputs":[],"stateMutability":"nonpayable","type":"function"},
  {"inputs":[{"internalType":"string","name":"_traceID","type":"string"},{"internalType":"string","name":"_kind","type":"string"},{"internalType":"string","name":"_dataHash","type":"string"},{"internalType":"string","name":"_operator","type":"string"}],"name":"appendAuxEvidence","outputs":[],"stateMutability":"nonpayable","type":"function"},
  {"inputs":[{"internalType":"string","name":"_traceID","type":"string"}],"name":"queryEvidence","outputs":[{"internalType":"string","name":"traceID","type":"string"},{"internalType":"string","name":"dataHash","type":"string"},{"internalType":"string","name":"operator","type":"string"},{"internalType":"uint256","name":"timestamp","type":"uint256"},{"internalType":"uint256","name":"blockNumber","type":"uint256"},{"internalType":"bool","name":"isValid","type":"bool"}],"stateMutability":"view","type":"function"},
  {"inputs":[{"internalType":"string","name":"_traceID","type":"string"}],"name":"verifyTraceID","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},
  {"inputs":[{"internalType":"string","name":"_traceID","type":"string"}],"name":"getHistoryCount","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},
  {"inputs":[{"internalType":"string","name":"_traceID","type":"string"},{"internalType":"uint256","name":"_index","type":"uint256"}],"name":"getHistoryRecord","outputs":[{"internalType":"string","name":"traceID","type":"string"},{"internalType":"string","name":"operationType","type":"string"},{"internalType":"string","name":"oldHash","type":"string"},{"internalType":"string","name":"newHash","type":"string"},{"internalType":"string","name":"operator","type":"string"},{"internalType":"uint256","name":"timestamp","type":"uint256"},{"internalType":"string","name":"changeDetails","type":"string"}],"stateMutability":"view","type":"function"}
]`

func initTraceEvidenceContract() error {
	contractMode = false
	traceEvidenceAddr = common.Address{}
	traceEvidenceABI = abi.ABI{}

	addrText := strings.TrimSpace(configAddress())
	if addrText == "" {
		return nil
	}
	if !common.IsHexAddress(addrText) {
		return fmt.Errorf("invalid contract_address: %s", addrText)
	}
	parsedABI, err := abi.JSON(strings.NewReader(traceEvidenceABIJSON))
	if err != nil {
		return fmt.Errorf("parse TraceEvidence ABI: %w", err)
	}
	traceEvidenceAddr = common.HexToAddress(addrText)
	traceEvidenceABI = parsedABI
	contractMode = true
	return nil
}

func configAddress() string {
	if config.AppConfig == nil {
		return ""
	}
	return config.AppConfig.Blockchain.FISCOBCOS.ContractAddress
}

func useContractMode() bool {
	return contractMode && traceEvidenceAddr != (common.Address{})
}

func sendContractTx(method string, args ...interface{}) (common.Hash, error) {
	if !useContractMode() {
		return common.Hash{}, fmt.Errorf("contract mode disabled")
	}
	data, err := traceEvidenceABI.Pack(method, args...)
	if err != nil {
		return common.Hash{}, fmt.Errorf("pack %s: %w", method, err)
	}
	return sendTx(traceEvidenceAddr, data)
}

func callContract(method string, args ...interface{}) ([]interface{}, error) {
	if !useContractMode() {
		return nil, fmt.Errorf("contract mode disabled")
	}
	if err := ensureConnected(3 * time.Second); err != nil {
		return nil, err
	}
	data, err := traceEvidenceABI.Pack(method, args...)
	if err != nil {
		return nil, fmt.Errorf("pack %s: %w", method, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := FISCOClient.CallContract(ctx, ethereum.CallMsg{
		From: senderAddr,
		To:   &traceEvidenceAddr,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("call %s: %w", method, err)
	}
	values, err := traceEvidenceABI.Unpack(method, out)
	if err != nil {
		return nil, fmt.Errorf("unpack %s: %w", method, err)
	}
	return values, nil
}

func callContractBool(method string, args ...interface{}) (bool, error) {
	values, err := callContract(method, args...)
	if err != nil {
		return false, err
	}
	if len(values) != 1 {
		return false, fmt.Errorf("%s returned %d values", method, len(values))
	}
	v, ok := values[0].(bool)
	if !ok {
		return false, fmt.Errorf("%s returned non-bool", method)
	}
	return v, nil
}

func callContractBigInt(method string, args ...interface{}) (*big.Int, error) {
	values, err := callContract(method, args...)
	if err != nil {
		return nil, err
	}
	if len(values) != 1 {
		return nil, fmt.Errorf("%s returned %d values", method, len(values))
	}
	v, ok := values[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("%s returned non-bigint", method)
	}
	return v, nil
}

