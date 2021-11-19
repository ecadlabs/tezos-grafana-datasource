// Package block implements block related types
package block

import (
	"math/big"
	"time"

	"github.com/ecadlabs/tezos-grafana-datasource/pkg/model"
)

type Block struct {
	Protocol   model.Base58    `json:"protocol"`
	ChainID    model.Base58    `json:"chain_id"`
	Hash       model.Base58    `json:"hash"`
	Header     RawHeader       `json:"header"`
	Metadata   *HeaderMetadata `json:"metadata,omitempty"`
	Operations Operations      `json:"operations"`
}

func (b *Block) GetHeader() *Header {
	return &Header{
		RawHeader: b.Header,
		Protocol:  b.Protocol,
		ChainID:   b.ChainID,
		Hash:      b.Hash,
	}
}

func (b *Block) GetMetadataInfo() *MetadataInfo {
	m := b.Metadata
	if m == nil {
		return nil
	}
	return &MetadataInfo{
		Protocol:                 m.Protocol,
		NextProtocol:             m.NextProtocol,
		MaxOperationsTTL:         m.MaxOperationsTTL,
		MaxOperationDataLength:   m.MaxOperationDataLength,
		MaxBlockHeaderLength:     m.MaxBlockHeaderLength,
		Baker:                    m.Baker,
		LevelInfo:                m.LevelInfo,
		VotingPeriodInfo:         m.VotingPeriodInfo,
		NonceHash:                m.NonceHash,
		ConsumedGas:              m.ConsumedGas,
		LiquidityBakingEscapeEMA: m.LiquidityBakingEscapeEMA,
	}
}

func (b *Block) Stat() *Statistics {
	stat := newStatistics()
	for _, tmp := range b.Operations {
		for _, operation := range tmp {
			for _, contents := range operation.Contents {
				stat.updateFromContents(contents)
			}
		}
	}

	if b.Metadata != nil {
		for _, ir := range b.Metadata.ImplicitOperationsResults {
			stat.updateFromContents(ir)
		}
	}

	return stat
}

type ShellHeader struct {
	Hash           model.Base58  `json:"hash"`
	Level          int64         `json:"level"`
	Proto          uint64        `json:"proto"`
	Predecessor    model.Base58  `json:"predecessor"`
	Timestamp      time.Time     `json:"timestamp"`
	ValidationPass uint64        `json:"validation_pass"`
	OperationsHash model.Base58  `json:"operations_hash"`
	Fitness        []model.Bytes `json:"fitness"`
	Context        model.Base58  `json:"context"`
	ProtocolData   model.Bytes   `json:"protocol_data"`
}

type Header struct {
	Protocol model.Base58 `json:"protocol"`
	ChainID  model.Base58 `json:"chain_id"`
	Hash     model.Base58 `json:"hash"`
	RawHeader
}

type RawHeader struct {
	Level                     int64         `json:"level"`
	Proto                     uint64        `json:"proto"`
	Predecessor               model.Base58  `json:"predecessor"`
	Timestamp                 time.Time     `json:"timestamp"`
	ValidationPass            uint64        `json:"validation_pass"`
	OperationsHash            model.Base58  `json:"operations_hash"`
	Fitness                   []model.Bytes `json:"fitness"`
	Context                   model.Base58  `json:"context"`
	Priority                  uint64        `json:"priority"`
	ProofOfWorkNonce          model.Bytes   `json:"proof_of_work_nonce"`
	SeedNonceHash             model.Base58  `json:"seed_nonce_hash,omitempty"`
	LiquidityBakingEscapeVote bool          `json:"liquidity_baking_escape_vote"`
	Signature                 model.Base58  `json:"signature"`
}

type HeaderMetadata struct {
	Protocol               model.Base58    `json:"protocol"`
	NextProtocol           model.Base58    `json:"next_protocol"`
	TestChainStatus        TestChainStatus `json:"test_chain_status"`
	MaxOperationsTTL       int64           `json:"max_operations_ttl"`
	MaxOperationDataLength int64           `json:"max_operation_data_length"`
	MaxBlockHeaderLength   int64           `json:"max_block_header_length"`
	MaxOperationListLength []struct {
		MaxSize int64  `json:"max_size"`
		MaxOp   *int64 `json:"max_op"`
	} `json:"max_operation_list_length"`
	Baker     model.Base58 `json:"baker"`
	LevelInfo struct {
		Level              int64 `json:"level"`
		LevelPosition      int64 `json:"level_position"`
		Cycle              int64 `json:"cycle"`
		CyclePosition      int64 `json:"cycle_position"`
		ExpectedCommitment bool  `json:"expected_commitment"`
	} `json:"level_info"`
	VotingPeriodInfo struct {
		VotingPeriod struct {
			Index         int64  `json:"index"`
			Kind          string `json:"kind"`
			StartPosition int64  `json:"start_position"`
		} `json:"voting_period"`
		Position  int64 `json:"position"`
		Remaining int64 `json:"remaining"`
	} `json:"voting_period_info"`
	NonceHash                 model.Base58              `json:"nonce_hash"`
	ConsumedGas               big.Int                   `json:"consumed_gas,string"`
	Deactivated               []model.Base58            `json:"deactivated"`
	BalanceUpdates            BalanceUpdates            `json:"balance_updates"`
	LiquidityBakingEscapeEMA  int64                     `json:"liquidity_baking_escape_ema"`
	ImplicitOperationsResults []ImplicitOperationResult `json:"implicit_operations_results"`
}

type TestChainStatus struct {
	Status     string       `json:"status"`
	Protocol   model.Base58 `json:"protocol,omitempty"`
	Expiration *time.Time   `json:"expiration,omitempty"`
	ChainID    model.Base58 `json:"chain_id,omitempty"`
	Genesis    model.Base58 `json:"genesis,omitempty"`
}
