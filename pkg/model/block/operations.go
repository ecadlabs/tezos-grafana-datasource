package block

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ecadlabs/jtree"
	"github.com/ecadlabs/tezos-grafana-datasource/pkg/model"
)

type Operations [][]*Operation

type Operation struct {
	Protocol  model.Base58                 `json:"protocol"`
	ChainID   model.Base58                 `json:"chain_id"`
	Hash      model.Base58                 `json:"hash"`
	Branch    model.Base58                 `json:"branch"`
	Contents  []OperationContentsAndResult `json:"contents"`
	Signature model.Base58                 `json:"signature,omitempty"`
}

type OperationKind string

const (
	OpEndorsement               OperationKind = "endorsement"
	OpSeedNonceRevelation       OperationKind = "seed_nonce_revelation"
	OpEndorsementWithSlot       OperationKind = "endorsement_with_slot"
	OpDoubleEndorsementEvidence OperationKind = "double_endorsement_evidence"
	OpDoubleBakingEvidence      OperationKind = "double_baking_evidence"
	OpActivateAccount           OperationKind = "activate_account"
	OpProposals                 OperationKind = "proposals"
	OpBallot                    OperationKind = "ballot"
	OpReveal                    OperationKind = "reveal"
	OpTransaction               OperationKind = "transaction"
	OpOrigination               OperationKind = "origination"
	OpDelegation                OperationKind = "delegation"
	OpRegisterGlobalConstant    OperationKind = "register_global_constant"
)

type OperationStatus string

const (
	StatusApplied     OperationStatus = "applied"
	StatusFailed      OperationStatus = "failed"
	StatusSkipped     OperationStatus = "skipped"
	StatusBacktracked OperationStatus = "backtracked"
)

type OperationContents interface {
	OperationKind() OperationKind
}

type OperationContentsAndResult interface {
	OperationContents
	OperationMetadata() OperationMetadata // can be nil
}

func operationContentsAndResultFunc(n jtree.Node, ctx *jtree.Context) (OperationContentsAndResult, error) {
	obj, ok := n.(jtree.Object)
	if !ok {
		return nil, fmt.Errorf("object expected: %t", n)
	}
	kind, ok := obj.FieldByName("kind").(jtree.String)
	if !ok {
		return nil, errors.New("kind field is missing")
	}

	var dest OperationContentsAndResult
	switch OperationKind(kind) {
	case OpEndorsement:
		dest = new(EndorsementOperationContents)
	case OpSeedNonceRevelation:
		dest = new(SeedNonceRevelationOperationContents)
	case OpEndorsementWithSlot:
		dest = new(EndorsementWithSlotOperationContents)
	case OpDoubleEndorsementEvidence:
		dest = new(DoubleEndorsementEvidenceOperationContents)
	case OpDoubleBakingEvidence:
		dest = new(DoubleBakingEvidenceOperationContents)
	case OpActivateAccount:
		dest = new(ActivateAccountOperationContents)
	case OpProposals:
		dest = new(ProposalsOperationContents)
	case OpBallot:
		dest = new(BallotOperationContents)
	case OpReveal:
		dest = new(RevealOperationContents)
	case OpTransaction:
		dest = new(TransactionOperationContents)
	case OpOrigination:
		dest = new(OriginationOperationContents)
	case OpDelegation:
		dest = new(DelegationOperationContents)
	case OpRegisterGlobalConstant:
		dest = new(RegisterGlobalConstantOperationContents)
	default:
		return nil, fmt.Errorf("unknown operation: %s", kind)
	}
	var tmp interface{} = dest // clear interface type and pass empty interface with pointer
	err := n.Decode(tmp, jtree.OpCtx(ctx))
	return dest, err
}

type OperationMetadata interface{}

type OperationMetadataWithResult interface {
	GetResult() (OperationResult, InternalOperationResults)
}

type WithBalanceUpdates interface {
	GetBalanceUpdates() BalanceUpdates
}

type WithConsumedGas interface {
	GetConsumedGas() (gas, milligas *big.Int) // can be nil
}

type WithErrors interface {
	GetErrors() []*model.Error // can be nil
}

type OperationResult interface {
	GetStatus() OperationStatus
}

type OperationResultFailed struct {
	Status OperationStatus `json:"status"`
	Errors []*model.Error  `json:"errors"`
}

type OperationResultSkipped struct {
	Status OperationStatus `json:"status"`
}

type InternalOperationResult interface {
	OperationContents
	OperationResult() OperationResult
}

func internalOperationResultFunc(n jtree.Node, ctx *jtree.Context) (InternalOperationResult, error) {
	obj, ok := n.(jtree.Object)
	if !ok {
		return nil, fmt.Errorf("object expected: %t", n)
	}
	kind, ok := obj.FieldByName("kind").(jtree.String)
	if !ok {
		return nil, errors.New("kind field is missing")
	}
	var dest InternalOperationResult
	switch OperationKind(kind) {
	case OpReveal:
		dest = new(RevealInternalOperationResult)
	case OpTransaction:
		dest = new(TransactionInternalOperationResult)
	case OpOrigination:
		dest = new(OriginationInternalOperationResult)
	case OpDelegation:
		dest = new(DelegationInternalOperationResult)
	case OpRegisterGlobalConstant:
		dest = new(RegisterGlobalConstantInternalOperationResult)
	default:
		return nil, fmt.Errorf("unknown internal operation: %s", kind)
	}
	var tmp interface{} = dest // clear interface type and pass empty interface with pointer
	err := n.Decode(tmp, jtree.OpCtx(ctx))
	return dest, err
}

type ImplicitOperationResult interface {
	OperationKind() OperationKind
}

func implicitOperationResultFunc(n jtree.Node, ctx *jtree.Context) (ImplicitOperationResult, error) {
	obj, ok := n.(jtree.Object)
	if !ok {
		return nil, fmt.Errorf("object expected: %t", n)
	}
	kind, ok := obj.FieldByName("kind").(jtree.String)
	if !ok {
		return nil, errors.New("kind field is missing")
	}
	var dest ImplicitOperationResult
	switch OperationKind(kind) {
	case OpReveal:
		dest = new(RevealImplicitOperationResult)
	case OpTransaction:
		dest = new(TransactionImplicitOperationResult)
	case OpOrigination:
		dest = new(OriginationImplicitOperationResult)
	case OpDelegation:
		dest = new(DelegationImplicitOperationResult)
	default:
		return nil, fmt.Errorf("unknown internal operation: %s", kind)
	}
	var tmp interface{} = dest // clear interface type and pass empty interface with pointer
	err := n.Decode(tmp, jtree.OpCtx(ctx))
	return dest, err
}

type InternalOperationResults []InternalOperationResult

type BalanceUpdatesMetadata struct {
	BalanceUpdates BalanceUpdates `json:"balance_updates"`
}

func (b *BalanceUpdatesMetadata) GetBalanceUpdates() BalanceUpdates { return b.BalanceUpdates }

// Balance updates

type BalanceUpdate interface {
	BalanceUpdateKind() string
	BalanceUpdateChange() int64
	BalanceUpdateOrigin() string
}

func balanceUpdateFunc(n jtree.Node, ctx *jtree.Context) (BalanceUpdate, error) {
	obj, ok := n.(jtree.Object)
	if !ok {
		return nil, fmt.Errorf("object expected: %t", n)
	}
	kind, ok := obj.FieldByName("kind").(jtree.String)
	if !ok {
		return nil, errors.New("kind field is missing")
	}
	var dest BalanceUpdate
	switch kind {
	case "contract":
		dest = new(ContractBalanceUpdate)
	case "freezer":
		dest = new(FreezerBalanceUpdate)
	default:
		return nil, fmt.Errorf("unknown balance update kind: %s", kind)
	}
	var tmp interface{} = dest // clear interface type and pass empty interface with pointer
	err := n.Decode(tmp, jtree.OpCtx(ctx))
	return dest, err
}

type ContractBalanceUpdate struct {
	Kind     string       `json:"kind"`
	Contract model.Base58 `json:"contract"`
	Change   int64        `json:"change,string"`
	Origin   string       `json:"origin"`
}

func (u *ContractBalanceUpdate) BalanceUpdateKind() string   { return u.Kind }
func (u *ContractBalanceUpdate) BalanceUpdateChange() int64  { return u.Change }
func (u *ContractBalanceUpdate) BalanceUpdateOrigin() string { return u.Origin }

type FreezerBalanceUpdate struct {
	Kind     string       `json:"kind"`
	Category string       `json:"category"`
	Delegate model.Base58 `json:"delegate"`
	Cycle    *int64       `json:"cycle"`
	Change   int64        `json:"change,string"`
	Origin   string       `json:"origin"`
}

func (u *FreezerBalanceUpdate) BalanceUpdateKind() string   { return u.Kind }
func (u *FreezerBalanceUpdate) BalanceUpdateChange() int64  { return u.Change }
func (u *FreezerBalanceUpdate) BalanceUpdateOrigin() string { return u.Origin }

type BalanceUpdates []BalanceUpdate

// Operations

type EndorsementWithSlotOperationContents struct {
	Kind        OperationKind                 `json:"kind"`
	Endorsement InlinedEndorsementOperation   `json:"endorsement"`
	Slot        uint64                        `json:"slot"`
	Metadata    *EndorsementOperationMetadata `json:"metadata,omitempty"`
}

type InlinedEndorsementOperation struct {
	Branch     model.Base58                        `json:"branch"`
	Operations InlinedEndorsementOperationContents `json:"operations"`
	Signature  model.Base58                        `json:"signature,omitempty"`
}

type InlinedEndorsementOperationContents struct {
	Kind  OperationKind `json:"kind"`
	Level int64         `json:"level"`
}

type EndorsementOperationMetadata struct {
	BalanceUpdates BalanceUpdates `json:"balance_updates"`
	Delegate       model.Base58   `json:"delegate"`
	Slots          []uint64       `json:"slots"`
}

func (e *EndorsementOperationMetadata) GetBalanceUpdates() BalanceUpdates    { return e.BalanceUpdates }
func (e *EndorsementWithSlotOperationContents) OperationKind() OperationKind { return e.Kind }

func (e *EndorsementWithSlotOperationContents) OperationMetadata() OperationMetadata {
	if e.Metadata != nil {
		return e.Metadata
	}
	return nil
}

type EndorsementOperationContents struct {
	Kind     OperationKind                 `json:"kind"`
	Level    int64                         `json:"level"`
	Metadata *EndorsementOperationMetadata `json:"metadata,omitempty"`
}

func (e *EndorsementOperationContents) OperationKind() OperationKind { return e.Kind }

func (e *EndorsementOperationContents) OperationMetadata() OperationMetadata {
	if e.Metadata != nil {
		return e.Metadata
	}
	return nil
}

type SeedNonceRevelationOperationContents struct {
	Kind     OperationKind           `json:"kind"`
	Level    int64                   `json:"level"`
	Nonce    []byte                  `json:"nonce,hex"`
	Metadata *BalanceUpdatesMetadata `json:"metadata,omitempty"`
}

func (s *SeedNonceRevelationOperationContents) OperationKind() OperationKind { return s.Kind }
func (s *SeedNonceRevelationOperationContents) OperationMetadata() OperationMetadata {
	if s.Metadata != nil {
		return s.Metadata
	}
	return nil
}

type DoubleEndorsementEvidenceOperationContents struct {
	Kind     OperationKind               `json:"kind"`
	Slot     uint64                      `json:"slot"`
	Op1      InlinedEndorsementOperation `json:"op1"`
	Op2      InlinedEndorsementOperation `json:"op2"`
	Metadata *BalanceUpdatesMetadata     `json:"metadata,omitempty"`
}

func (d *DoubleEndorsementEvidenceOperationContents) OperationKind() OperationKind { return d.Kind }
func (d *DoubleEndorsementEvidenceOperationContents) OperationMetadata() OperationMetadata {
	if d.Metadata != nil {
		return d.Metadata
	}
	return nil
}

type DoubleBakingEvidenceOperationContents struct {
	Kind     OperationKind           `json:"kind"`
	Bh1      Header                  `json:"bh1"`
	Bh2      Header                  `json:"bh2"`
	Metadata *BalanceUpdatesMetadata `json:"metadata,omitempty"`
}

func (d *DoubleBakingEvidenceOperationContents) OperationKind() OperationKind { return d.Kind }
func (d *DoubleBakingEvidenceOperationContents) OperationMetadata() OperationMetadata {
	if d.Metadata != nil {
		return d.Metadata
	}
	return nil
}

type ActivateAccountOperationContents struct {
	Kind     OperationKind           `json:"kind"`
	PKH      model.Base58            `json:"pkh"`
	Metadata *BalanceUpdatesMetadata `json:"metadata,omitempty"`
}

func (a *ActivateAccountOperationContents) OperationKind() OperationKind { return a.Kind }
func (a *ActivateAccountOperationContents) OperationMetadata() OperationMetadata {
	if a.Metadata != nil {
		return a.Metadata
	}
	return nil
}

type ProposalsOperationContents struct {
	Kind      OperationKind  `json:"kind"`
	Source    model.Base58   `json:"source"`
	Period    int64          `json:"period"`
	Proposals []model.Base58 `json:"proposals"`
	Metadata  *struct{}      `json:"metadata,omitempty"`
}

func (p *ProposalsOperationContents) OperationKind() OperationKind { return p.Kind }
func (p *ProposalsOperationContents) OperationMetadata() OperationMetadata {
	if p.Metadata != nil {
		return p.Metadata
	}
	return nil
}

type BallotOperationContents struct {
	Kind     OperationKind `json:"kind"`
	Source   model.Base58  `json:"source"`
	Period   int64         `json:"period"`
	Proposal model.Base58  `json:"proposal"`
	Ballot   string        `json:"ballot"`
	Metadata *struct{}     `json:"metadata,omitempty"`
}

func (b *BallotOperationContents) OperationKind() OperationKind { return b.Kind }
func (b *BallotOperationContents) OperationMetadata() OperationMetadata {
	if b.Metadata != nil {
		return b.Metadata
	}
	return nil
}

type Script struct {
	Code    Expr `json:"code"`
	Storage Expr `json:"storage"`
}

func init() {
	jtree.RegisterType(operationContentsAndResultFunc)
	jtree.RegisterType(internalOperationResultFunc)
	jtree.RegisterType(implicitOperationResultFunc)
	jtree.RegisterType(balanceUpdateFunc)
}

var (
	_ WithBalanceUpdates = (*BalanceUpdatesMetadata)(nil)
	_ WithBalanceUpdates = (*EndorsementOperationMetadata)(nil)
)
