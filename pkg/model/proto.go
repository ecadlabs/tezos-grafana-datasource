package model

import "math/big"

type ProtocolConstants struct {
	ProofOfWorkNonceSize              uint64     `json:"proof_of_work_nonce_size"`
	NonceLength                       uint64     `json:"nonce_length"`
	MaxAnonOpsPerBlock                uint64     `json:"max_anon_ops_per_block"`
	MaxOperationDataLength            int64      `json:"max_operation_data_length"`
	MaxProposalsPerDelegate           uint64     `json:"max_proposals_per_delegate"`
	PreservedCycles                   uint64     `json:"preserved_cycles"`
	BlocksPerCycle                    int64      `json:"blocks_per_cycle"`
	BlocksPerCommitment               int64      `json:"blocks_per_commitment"`
	BlocksPerRollSnapshot             int64      `json:"blocks_per_roll_snapshot"`
	BlocksPerVotingPeriod             int64      `json:"blocks_per_voting_period"`
	TimeBetweenBlocks                 []int64    `json:"time_between_blocks,[string]"`
	EndorsersPerBlock                 uint64     `json:"endorsers_per_block"`
	HardGasLimitPerOperation          big.Int    `json:"hard_gas_limit_per_operation,string"`
	HardGasLimitPerBlock              big.Int    `json:"hard_gas_limit_per_block,string"`
	ProofOfWorkThreshold              int64      `json:"proof_of_work_threshold,string"`
	TokensPerRoll                     big.Int    `json:"tokens_per_roll,string"`
	MichelsonMaximumTypeSize          uint64     `json:"michelson_maximum_type_size"`
	SeedNonceRevelationTip            big.Int    `json:"seed_nonce_revelation_tip,string"`
	OriginationSize                   int64      `json:"origination_size"`
	BlockSecurityDeposit              big.Int    `json:"block_security_deposit,string"`
	EndorsementSecurityDeposit        big.Int    `json:"endorsement_security_deposit,string"`
	BakingRewardPerEndorsement        []*big.Int `json:"baking_reward_per_endorsement,[string]"`
	EndorsementReward                 []*big.Int `json:"endorsement_reward,[string]"`
	CostPerByte                       big.Int    `json:"cost_per_byte,string"`
	HardStorageLimitPerOperation      big.Int    `json:"hard_storage_limit_per_operation,string"`
	QuorumMin                         int64      `json:"quorum_min"`
	QuorumMax                         int64      `json:"quorum_max"`
	MinProposalQuorum                 int64      `json:"min_proposal_quorum"`
	InitialEndorsers                  uint64     `json:"initial_endorsers"`
	DelayPerMissingEndorsement        int64      `json:"delay_per_missing_endorsement,string"`
	MinimalBlockDelay                 int64      `json:"minimal_block_delay,string"`
	LiquidityBakingSubsidy            big.Int    `json:"liquidity_baking_subsidy,string"`
	LiquidityBakingSunsetLevel        int64      `json:"liquidity_baking_sunset_level"`
	LiquidityBakingEscapeEMAThreshold int64      `json:"liquidity_baking_escape_ema_threshold"`
}

type Error struct {
	King string `json:"kind"`
	ID   string `json:"id"`
}
