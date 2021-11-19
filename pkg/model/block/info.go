package block

import (
	"math/big"
	"time"

	"github.com/ecadlabs/tezos-grafana-datasource/pkg/model"
)

type Info struct {
	Header       *Header       `json:"header"`
	Stat         *Statistics   `json:"stat"`
	Metadata     *MetadataInfo `json:"metadata,omitempty"`
	MinValidTime time.Time     `json:"minimal_valid_time"`
}

type MetadataInfo struct {
	Protocol               model.Base58 `json:"protocol"`
	NextProtocol           model.Base58 `json:"next_protocol"`
	MaxOperationsTTL       int64        `json:"max_operations_ttl"`
	MaxOperationDataLength int64        `json:"max_operation_data_length"`
	MaxBlockHeaderLength   int64        `json:"max_block_header_length"`
	Baker                  model.Base58 `json:"baker"`
	LevelInfo              struct {
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
	NonceHash                model.Base58 `json:"nonce_hash"`
	ConsumedGas              big.Int      `json:"consumed_gas,string"`
	LiquidityBakingEscapeEMA int64        `json:"liquidity_baking_escape_ema"`
}
type NumOps struct {
	Total                     uint64 `json:"total"`
	Endorsement               uint64 `json:"endorsement"`
	SeedNonceRevelation       uint64 `json:"seed_nonce_revelation"`
	DoubleEndorsementEvidence uint64 `json:"double_endorsement_evidence"`
	DoubleBakingEvidence      uint64 `json:"double_baking_evidence"`
	ActivateAccount           uint64 `json:"activate_account"`
	Proposals                 uint64 `json:"proposals"`
	Ballot                    uint64 `json:"ballot"`
	Reveal                    uint64 `json:"reveal"`
	Transaction               uint64 `json:"transaction"`
	Origination               uint64 `json:"origination"`
	Delegation                uint64 `json:"delegation"`
	RegisterGlobalConstant    uint64 `json:"register_global_constant"`
}

type ConsumedMilligas struct {
	Total                  *big.Int `json:"total,string"`
	Reveal                 *big.Int `json:"reveal,string"`
	Transaction            *big.Int `json:"transaction,string"`
	Origination            *big.Int `json:"origination,string"`
	Delegation             *big.Int `json:"delegation,string"`
	RegisterGlobalConstant *big.Int `json:"register_global_constant,string"`
}

type StorageSize struct {
	Total                  *big.Int `json:"total,string"`
	Transaction            *big.Int `json:"transaction,string"`
	Origination            *big.Int `json:"origination,string"`
	RegisterGlobalConstant *big.Int `json:"register_global_constant,string"`
}

type Statistics struct {
	NumOps           NumOps           `json:"n_ops"`
	Slots            uint64           `json:"n_slots"`
	ConsumedMilligas ConsumedMilligas `json:"consumed_milligas"`
	StorageSize      StorageSize      `json:"storage_size"`
}

func newStatistics() *Statistics {
	return &Statistics{
		ConsumedMilligas: ConsumedMilligas{
			Total:                  new(big.Int),
			Reveal:                 new(big.Int),
			Transaction:            new(big.Int),
			Origination:            new(big.Int),
			Delegation:             new(big.Int),
			RegisterGlobalConstant: new(big.Int),
		},
		StorageSize: StorageSize{
			Total:                  new(big.Int),
			Transaction:            new(big.Int),
			Origination:            new(big.Int),
			RegisterGlobalConstant: new(big.Int),
		},
	}
}

func (s *Statistics) updateFromContents(op OperationContents) {
	var (
		consumedMilligas *big.Int
		storageSize      *big.Int
		numOps           *uint64
	)

	switch op.OperationKind() {
	case OpEndorsement, OpEndorsementWithSlot:
		numOps = &s.NumOps.Endorsement
	case OpSeedNonceRevelation:
		numOps = &s.NumOps.SeedNonceRevelation
	case OpDoubleEndorsementEvidence:
		numOps = &s.NumOps.DoubleEndorsementEvidence
	case OpDoubleBakingEvidence:
		numOps = &s.NumOps.DoubleBakingEvidence
	case OpActivateAccount:
		numOps = &s.NumOps.ActivateAccount
	case OpProposals:
		numOps = &s.NumOps.Proposals
	case OpBallot:
		numOps = &s.NumOps.Ballot
	case OpReveal:
		numOps = &s.NumOps.Reveal
		consumedMilligas = s.ConsumedMilligas.Reveal
	case OpTransaction:
		numOps = &s.NumOps.Transaction
		consumedMilligas = s.ConsumedMilligas.Transaction
		storageSize = s.StorageSize.Transaction
	case OpOrigination:
		numOps = &s.NumOps.Origination
		consumedMilligas = s.ConsumedMilligas.Origination
		storageSize = s.StorageSize.Origination
	case OpDelegation:
		numOps = &s.NumOps.Delegation
		consumedMilligas = s.ConsumedMilligas.Delegation
	case OpRegisterGlobalConstant:
		numOps = &s.NumOps.RegisterGlobalConstant
		consumedMilligas = s.ConsumedMilligas.RegisterGlobalConstant
		storageSize = s.StorageSize.RegisterGlobalConstant
	default:
		return
	}

	s.NumOps.Total++
	(*numOps)++

	var target interface{} = op
	if cr, ok := op.(OperationContentsAndResult); ok {
		meta := cr.OperationMetadata()
		if meta == nil {
			return
		}
		if em, ok := meta.(*EndorsementOperationMetadata); ok {
			s.Slots += uint64(len(em.Slots))
		}
		target = meta
	}

	if wr, ok := target.(WithResult); ok {
		target = wr.GetResult()
	}

	if mg, ok := target.(WithConsumedMilligas); ok {
		v := mg.GetConsumedMilligas()
		if v != nil {
			if consumedMilligas != nil {
				consumedMilligas.Add(consumedMilligas, v)
			}
			s.ConsumedMilligas.Total.Add(s.ConsumedMilligas.Total, v)
		}
	}
	if st, ok := target.(WithStorage); ok {
		v := st.GetStorageSize()
		if v != nil {
			if storageSize != nil {
				storageSize.Add(storageSize, v)
			}
			s.StorageSize.Total.Add(s.StorageSize.Total, v)
		}
	}

	if wir, ok := target.(WithInternalOperationResults); ok {
		for _, r := range wir.GetInternalOperationResults() {
			s.updateFromContents(r)
		}
	}
}
