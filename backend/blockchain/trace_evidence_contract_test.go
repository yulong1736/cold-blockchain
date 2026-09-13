package blockchain

import (
	"cold-chain-trace/backend/config"
	"testing"
)

func TestInitTraceEvidenceContract_ModeSwitch(t *testing.T) {
	config.AppConfig = &config.Config{}

	if err := initTraceEvidenceContract(); err != nil {
		t.Fatalf("empty address should not error: %v", err)
	}
	if useContractMode() {
		t.Fatal("contract mode should be disabled when contract_address is empty")
	}

	config.AppConfig.Blockchain.FISCOBCOS.ContractAddress = "0x1111111111111111111111111111111111111111"
	if err := initTraceEvidenceContract(); err != nil {
		t.Fatalf("valid address should enable contract mode: %v", err)
	}
	if !useContractMode() {
		t.Fatal("contract mode should be enabled when contract_address is set")
	}
	if _, ok := traceEvidenceABI.Methods["storeEvidence"]; !ok {
		t.Fatal("trace evidence ABI should include storeEvidence")
	}
}
