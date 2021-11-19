package block

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ecadlabs/jtree"
	"github.com/ecadlabs/tezos-grafana-datasource/pkg/model"
)

type OriginationOperationContents struct {
	Kind         OperationKind                 `json:"kind"`
	Source       model.Base58                  `json:"source"`
	Fee          big.Int                       `json:"fee,string"`
	Counter      big.Int                       `json:"counter,string"`
	GasLimit     big.Int                       `json:"gas_limit,string"`
	StorageLimit big.Int                       `json:"storage_limit,string"`
	Balance      big.Int                       `json:"balance,string"`
	Delegate     model.Base58                  `json:"delegate,omitempty"`
	Script       Script                        `json:"script"`
	Metadata     *OriginationOperationMetadata `json:"metadata,omitempty"`
}

func (o *OriginationOperationContents) OperationKind() OperationKind { return o.Kind }
func (o *OriginationOperationContents) OperationMetadata() OperationMetadata {
	if o.Metadata != nil {
		return o.Metadata
	}
	return nil
}

type OriginationOperationMetadata struct {
	BalanceUpdates           BalanceUpdates             `json:"balance_updates"`
	OperationResult          OriginationOperationResult `json:"operation_result"`
	InternalOperationResults InternalOperationResults   `json:"internal_operation_results,omitempty"`
}

func (o *OriginationOperationMetadata) GetBalanceUpdates() BalanceUpdates { return o.BalanceUpdates }
func (o *OriginationOperationMetadata) GetResult() OperationResult        { return o.OperationResult }
func (o *OriginationOperationMetadata) GetInternalOperationResults() InternalOperationResults {
	return o.InternalOperationResults
}

type OriginationOperationResult interface {
	OperationResult
	OriginationOperationResult()
}

type OriginationOperationResultBase struct {
	BigMapDiff          jtree.Node     `json:"big_map_diff,omitempty"`
	BalanceUpdates      BalanceUpdates `json:"balance_updates,omitempty"`
	OriginatedContracts []model.Base58 `json:"originated_contracts,omitempty"`
	ConsumedGas         *big.Int       `json:"consumed_gas,string,omitempty"`
	ConsumedMilligas    *big.Int       `json:"consumed_milligas,string,omitempty"`
	StorageSize         *big.Int       `json:"storage_size,string,omitempty"`
	PaidStorageSizeDiff *big.Int       `json:"paid_storage_size_diff,string,omitempty"`
	LazyStorageDiff     jtree.Node     `json:"lazy_storage_diff,omitempty"`
}

func (r *OriginationOperationResultBase) GetConsumedMilligas() *big.Int {
	return getConsumedMilligas(r.ConsumedGas, r.ConsumedMilligas)
}

func (r *OriginationOperationResultBase) GetStorageSize() *big.Int {
	return r.StorageSize
}

type OriginationOperationResultApplied struct {
	Status OperationStatus `json:"status"`
	OriginationOperationResultBase
}

func (r *OriginationOperationResultApplied) OriginationOperationResult() {}
func (r *OriginationOperationResultApplied) GetStatus() OperationStatus  { return r.Status }
func (r *OriginationOperationResultApplied) GetBalanceUpdates() BalanceUpdates {
	return r.BalanceUpdates
}

type OriginationOperationResultBacktracked struct {
	Status OperationStatus `json:"status"`
	Errors []*model.Error  `json:"errors,omitempty"`
	OriginationOperationResultBase
}

func (r *OriginationOperationResultBacktracked) OriginationOperationResult() {}
func (r *OriginationOperationResultBacktracked) GetStatus() OperationStatus  { return r.Status }
func (r *OriginationOperationResultBacktracked) GetErrors() []*model.Error   { return r.Errors }

type OriginationOperationResultFailed OperationResultFailed

func (r *OriginationOperationResultFailed) OriginationOperationResult() {}
func (r *OriginationOperationResultFailed) GetStatus() OperationStatus  { return r.Status }
func (r *OriginationOperationResultFailed) GetErrors() []*model.Error   { return r.Errors }

type OriginationOperationResultSkipped OperationResultSkipped

func (r *OriginationOperationResultSkipped) OriginationOperationResult() {}
func (r *OriginationOperationResultSkipped) GetStatus() OperationStatus  { return r.Status }

func originationOperationResultFunc(n jtree.Node, ctx *jtree.Context) (OriginationOperationResult, error) {
	obj, ok := n.(jtree.Object)
	if !ok {
		return nil, fmt.Errorf("object expected: %t", n)
	}
	status, ok := obj.FieldByName("status").(jtree.String)
	if !ok {
		return nil, errors.New("status field is missing")
	}
	var dest OriginationOperationResult
	switch OperationStatus(status) {
	case StatusApplied:
		dest = new(OriginationOperationResultApplied)
	case StatusFailed:
		dest = new(OriginationOperationResultFailed)
	case StatusSkipped:
		dest = new(OriginationOperationResultSkipped)
	case StatusBacktracked:
		dest = new(OriginationOperationResultBacktracked)
	default:
		return nil, fmt.Errorf("unknown operation result status: %s", status)
	}
	var tmp interface{} = dest // clear interface type and pass empty interface with pointer
	err := n.Decode(tmp, jtree.OpCtx(ctx))
	return dest, err
}

type OriginationInternalOperationResult struct {
	Kind     OperationKind              `json:"kind"`
	Source   model.Base58               `json:"source"`
	Nonce    uint64                     `json:"nonce"`
	Balance  big.Int                    `json:"balance,string"`
	Delegate model.Base58               `json:"delegate,omitempty"`
	Script   Script                     `json:"script"`
	Result   OriginationOperationResult `json:"result"`
}

func (o *OriginationInternalOperationResult) OperationKind() OperationKind { return o.Kind }
func (o *OriginationInternalOperationResult) GetResult() OperationResult   { return o.Result }

type OriginationImplicitOperationResult struct {
	Kind OperationKind `json:"kind"`
	OriginationOperationResultBase
}

func (r *OriginationImplicitOperationResult) OperationKind() OperationKind { return r.Kind }

var (
	_ WithBalanceUpdates           = (*OriginationOperationMetadata)(nil)
	_ WithInternalOperationResults = (*OriginationOperationMetadata)(nil)
	_ WithConsumedMilligas         = (*OriginationOperationResultApplied)(nil)
	_ WithConsumedMilligas         = (*OriginationOperationResultBacktracked)(nil)
	_ WithConsumedMilligas         = (*OriginationImplicitOperationResult)(nil)
	_ WithStorage                  = (*OriginationOperationResultApplied)(nil)
	_ WithStorage                  = (*OriginationOperationResultBacktracked)(nil)
	_ WithStorage                  = (*OriginationImplicitOperationResult)(nil)
	_ WithErrors                   = (*OriginationOperationResultFailed)(nil)
	_ WithErrors                   = (*OriginationOperationResultBacktracked)(nil)
	_ WithBalanceUpdates           = (*OriginationOperationResultApplied)(nil)
)

func init() {
	jtree.RegisterType(originationOperationResultFunc)
}
