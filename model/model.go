package model

import (
	"bytes"
	"encoding/json"
	"time"
)

type Block struct {
	Protocol   Base58          `json:"protocol"`
	ChainID    Base58          `json:"chain_id"`
	Hash       Base58          `json:"hash"`
	Header     RawBlockHeader  `json:"header"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
	Operations BlockOperations `json:"operations"`
}

func (b *Block) GetHeader() *BlockHeader {
	return &BlockHeader{
		RawBlockHeader: b.Header,
		Protocol:       b.Protocol,
		ChainID:        b.ChainID,
		Hash:           b.Hash,
	}
}

func (b *Block) Stat() *BlockStatistics {
	var (
		slots int
		opCnt int
		ops   NumOps
	)
	for _, tmp := range b.Operations {
		for _, operation := range tmp {
			for _, contents := range operation.Contents {
				opCnt++
				switch op := contents.(type) {
				case *EndorsementWithSlot:
					ops.Endorsement++
					if op.Metadata != nil {
						slots += len(op.Metadata.Slots)
					}
				case *Endorsement:
					ops.Endorsement++
					if op.Metadata != nil {
						slots += len(op.Metadata.Slots)
					}
				case OpaqueOperation:
					switch op.OperationKind() {
					case "seed_nonce_revelation":
						ops.SeedNonceRevelation++
					case "double_endorsement_evidence":
						ops.DoubleEndorsementEvidence++
					case "double_baking_evidence":
						ops.DoubleBakingEvidence++
					case "activate_account":
						ops.ActivateAccount++
					case "proposals":
						ops.Proposals++
					case "ballot":
						ops.Ballot++
					case "reveal":
						ops.Reveal++
					case "transaction":
						ops.Transaction++
					case "origination":
						ops.Origination++
					case "delegation":
						ops.Delegation++
					case "failing_noop":
						ops.FailingNoop++
					}
				}
			}
		}
	}
	return &BlockStatistics{
		NumOps: uint64(opCnt),
		Ops:    &ops,
		Slots:  uint64(slots),
	}
}

type BlockHeader struct {
	Protocol Base58 `json:"protocol"`
	ChainID  Base58 `json:"chain_id"`
	Hash     Base58 `json:"hash"`
	RawBlockHeader
}

type RawBlockHeader struct {
	Level                     int64     `json:"level"`
	Proto                     uint64    `json:"proto"`
	Predecessor               Base58    `json:"predecessor"`
	Timestamp                 time.Time `json:"timestamp"`
	ValidationPass            uint64    `json:"validation_pass"`
	OperationsHash            Base58    `json:"operations_hash"`
	Fitness                   []Bytes   `json:"fitness"`
	Context                   Base58    `json:"context"`
	Priority                  uint64    `json:"priority"`
	ProofOfWorkNonce          Bytes     `json:"proof_of_work_nonce"`
	SeedNonceHash             Base58    `json:"seed_nonce_hash,omitempty"`
	LiquidityBakingEscapeVote bool      `json:"liquidity_baking_escape_vote"`
	Signature                 Base58    `json:"signature"`
}

type ProtocolConstants struct {
	ProofOfWorkNonceSize              uint64    `json:"proof_of_work_nonce_size"`
	NonceLength                       uint64    `json:"nonce_length"`
	MaxAnonOpsPerBlock                uint64    `json:"max_anon_ops_per_block"`
	MaxOperationDataLength            int64     `json:"max_operation_data_length"`
	MaxProposalsPerDelegate           uint64    `json:"max_proposals_per_delegate"`
	PreservedCycles                   uint64    `json:"preserved_cycles"`
	BlocksPerCycle                    int64     `json:"blocks_per_cycle"`
	BlocksPerCommitment               int64     `json:"blocks_per_commitment"`
	BlocksPerRollSnapshot             int64     `json:"blocks_per_roll_snapshot"`
	BlocksPerVotingPeriod             int64     `json:"blocks_per_voting_period"`
	TimeBetweenBlocks                 []Int64   `json:"time_between_blocks"`
	EndorsersPerBlock                 uint64    `json:"endorsers_per_block"`
	HardGasLimitPerOperation          *BigInt   `json:"hard_gas_limit_per_operation"`
	HardGasLimitPerBlock              *BigInt   `json:"hard_gas_limit_per_block"`
	ProofOfWorkThreshold              int64     `json:"proof_of_work_threshold,string"`
	TokensPerRoll                     *BigInt   `json:"tokens_per_roll"`
	MichelsonMaximumTypeSize          uint64    `json:"michelson_maximum_type_size"`
	SeedNonceRevelationTip            *BigInt   `json:"seed_nonce_revelation_tip"`
	OriginationSize                   int64     `json:"origination_size"`
	BlockSecurityDeposit              *BigInt   `json:"block_security_deposit"`
	EndorsementSecurityDeposit        *BigInt   `json:"endorsement_security_deposit"`
	BakingRewardPerEndorsement        []*BigInt `json:"baking_reward_per_endorsement"`
	EndorsementReward                 []*BigInt `json:"endorsement_reward"`
	CostPerByte                       *BigInt   `json:"cost_per_byte"`
	HardStorageLimitPerOperation      *BigInt   `json:"hard_storage_limit_per_operation"`
	QuorumMin                         int64     `json:"quorum_min"`
	QuorumMax                         int64     `json:"quorum_max"`
	MinProposalQuorum                 int64     `json:"min_proposal_quorum"`
	InitialEndorsers                  uint64    `json:"initial_endorsers"`
	DelayPerMissingEndorsement        int64     `json:"delay_per_missing_endorsement,string"`
	MinimalBlockDelay                 int64     `json:"minimal_block_delay,string"`
	LiquidityBakingSubsidy            *BigInt   `json:"liquidity_baking_subsidy"`
	LiquidityBakingSunsetLevel        int64     `json:"liquidity_baking_sunset_level"`
	LiquidityBakingEscapeEMAThreshold int64     `json:"liquidity_baking_escape_ema_threshold"`
}

type BlockOperations [][]*BlockOperation

type BlockOperation struct {
	Protocol  Base58                 `json:"protocol"`
	ChainID   Base58                 `json:"chain_id"`
	Hash      Base58                 `json:"hash"`
	Branch    Base58                 `json:"branch"`
	Contents  BlockOperationContents `json:"contents"`
	Signature Base58                 `json:"signature,omitempty"`
}

type BlockOperationContents []Operation

func (ops *BlockOperationContents) UnmarshalJSON(text []byte) error {
	var tmp []json.RawMessage
	if err := json.Unmarshal(text, &tmp); err != nil {
		return err
	}
	res := make([]Operation, len(tmp))
	for i, rawOp := range tmp {
		var kind struct {
			Kind string `json:"kind"`
		}
		if err := json.Unmarshal(rawOp, &kind); err != nil {
			return err
		}

		var target Operation
		switch kind.Kind {
		case "endorsement_with_slot":
			target = new(EndorsementWithSlot)
		case "endorsement":
			target = new(Endorsement)
		default:
			target = new(OpaqueOperation)
		}

		dec := json.NewDecoder(bytes.NewReader(rawOp))
		dec.DisallowUnknownFields()
		if err := dec.Decode(target); err != nil {
			return err
		}
		res[i] = target
	}
	*ops = res
	return nil
}

type Operation interface {
	OperationKind() string
}

type EndorsementWithSlot struct {
	Kind        string               `json:"kind"`
	Endorsement InlinedEndorsement   `json:"endorsement"`
	Slot        uint64               `json:"slot"`
	Metadata    *EndorsementMetadata `json:"metadata,omitempty"`
}

type InlinedEndorsement struct {
	Branch     Base58                     `json:"branch"`
	Operations InlinedEndorsementContents `json:"operations"`
	Signature  Base58                     `json:"signature,omitempty"`
}

type InlinedEndorsementContents struct {
	Kind  string `json:"kind"`
	Level int64  `json:"level"`
}

type EndorsementMetadata struct {
	BalanceUpdates BalanceUpdates `json:"balance_updates"`
	Delegate       Base58         `json:"delegate"`
	Slots          []uint64       `json:"slots"`
}

func (*EndorsementWithSlot) OperationKind() string {
	return "endorsement_with_slot"
}

type Endorsement struct {
	Kind     string               `json:"kind"`
	Level    int64                `json:"level"`
	Metadata *EndorsementMetadata `json:"metadata,omitempty"`
}

func (*Endorsement) OperationKind() string {
	return "endorsement"
}

type OpaqueOperation map[string]interface{}

func (o OpaqueOperation) OperationKind() string {
	if k, ok := o["kind"].(string); ok {
		return k
	}
	return ""
}

type BalanceUpdate interface {
	BalanceUpdateKind() string
}

type ContractBalanceUpdate struct {
	Kind     string `json:"kind"`
	Contract Base58 `json:"contract"`
	Change   Int64  `json:"change"`
	Origin   string `json:"origin"`
}

func (*ContractBalanceUpdate) BalanceUpdateKind() string {
	return "contract"
}

type NonContractBalanceUpdate struct {
	Kind     string `json:"kind"`
	Category string `json:"category"`
	Delegate Base58 `json:"delegate"`
	Cycle    int64  `json:"cycle"`
	Change   Int64  `json:"change"`
	Origin   string `json:"origin"`
}

func (u *NonContractBalanceUpdate) BalanceUpdateKind() string {
	return u.Kind
}

type BalanceUpdates []BalanceUpdate

func (u *BalanceUpdates) UnmarshalJSON(text []byte) error {
	var tmp []json.RawMessage
	if err := json.Unmarshal(text, &tmp); err != nil {
		return err
	}
	res := make([]BalanceUpdate, len(tmp))
	for i, rawOp := range tmp {
		var kind struct {
			Kind string `json:"kind"`
		}
		if err := json.Unmarshal(rawOp, &kind); err != nil {
			return err
		}

		var target BalanceUpdate
		switch kind.Kind {
		case "contract":
			target = new(ContractBalanceUpdate)
		default:
			target = new(NonContractBalanceUpdate)
		}

		dec := json.NewDecoder(bytes.NewReader(rawOp))
		dec.DisallowUnknownFields()
		if err := dec.Decode(target); err != nil {
			return err
		}
		res[i] = target
	}
	*u = res
	return nil
}

type BlockInfo struct {
	Header       *BlockHeader     `json:"header"`
	Stat         *BlockStatistics `json:"statistics"`
	MinValidTime time.Time        `json:"minimal_valid_time"`
}

type BlockStatistics struct {
	NumOps uint64  `json:"n_ops_total"`
	Ops    *NumOps `json:"n_ops"`
	Slots  uint64  `json:"endorsement_slots"`
}

type NumOps struct {
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
	FailingNoop               uint64 `json:"failing_noop"`
}
