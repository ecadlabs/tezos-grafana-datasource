package block

import (
	"time"
)

type Info struct {
	Header       *Header     `json:"header"`
	Stat         *Statistics `json:"statistics"`
	MinValidTime time.Time   `json:"minimal_valid_time"`
}

type Statistics struct {
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
