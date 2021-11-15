package block

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ecadlabs/jtree"
	"github.com/ecadlabs/tezos-grafana-datasource/pkg/model"
)

type DelegationOperationContents struct {
	Kind         OperationKind                `json:"kind"`
	Source       model.Base58                 `json:"source"`
	Fee          big.Int                      `json:"fee,string"`
	Counter      big.Int                      `json:"counter,string"`
	GasLimit     big.Int                      `json:"gas_limit,string"`
	StorageLimit big.Int                      `json:"storage_limit,string"`
	Delegate     model.Base58                 `json:"delegate,omitempty"`
	Metadata     *DelegationOperationMetadata `json:"metadata,omitempty"`
}

func (d *DelegationOperationContents) OperationKind() OperationKind { return d.Kind }
func (d *DelegationOperationContents) OperationMetadata() OperationMetadata {
	if d.Metadata != nil {
		return d.Metadata
	}
	return nil
}

type DelegationOperationMetadata struct {
	BalanceUpdates           BalanceUpdates            `json:"balance_updates"`
	OperationResult          DelegationOperationResult `json:"operation_result"`
	InternalOperationResults InternalOperationResults  `json:"internal_operation_results,omitempty"`
}

func (d *DelegationOperationMetadata) GetBalanceUpdates() BalanceUpdates { return d.BalanceUpdates }
func (d *DelegationOperationMetadata) GetResult() (OperationResult, InternalOperationResults) {
	return d.OperationResult, d.InternalOperationResults
}

type DelegationOperationResult interface {
	OperationResult
	DelegationOperationResult()
}

type DelegationOperationResultBase RevealOperationResultBase

func (r *DelegationOperationResultBase) GetConsumedGas() (gas, milligas *big.Int) {
	return r.ConsumedGas, r.ConsumedMilligas
}

type DelegationOperationResultApplied struct {
	Status OperationStatus `json:"status"`
	DelegationOperationResultBase
}

func (r *DelegationOperationResultApplied) DelegationOperationResult() {}
func (r *DelegationOperationResultApplied) GetStatus() OperationStatus { return r.Status }

type DelegationOperationResultBacktracked struct {
	Status OperationStatus `json:"status"`
	Errors []*model.Error  `json:"errors,omitempty"`
	DelegationOperationResultBase
}

func (r *DelegationOperationResultBacktracked) DelegationOperationResult() {}
func (r *DelegationOperationResultBacktracked) GetStatus() OperationStatus { return r.Status }
func (r *DelegationOperationResultBacktracked) GetErrors() []*model.Error  { return r.Errors }

type DelegationOperationResultFailed OperationResultFailed

func (r *DelegationOperationResultFailed) DelegationOperationResult() {}
func (r *DelegationOperationResultFailed) GetStatus() OperationStatus { return r.Status }
func (r *DelegationOperationResultFailed) GetErrors() []*model.Error  { return r.Errors }

type DelegationOperationResultSkipped OperationResultSkipped

func (r *DelegationOperationResultSkipped) DelegationOperationResult() {}
func (r *DelegationOperationResultSkipped) GetStatus() OperationStatus { return r.Status }

func delegationOperationResultFunc(n jtree.Node, ctx *jtree.Context) (DelegationOperationResult, error) {
	obj, ok := n.(jtree.Object)
	if !ok {
		return nil, fmt.Errorf("object expected: %t", n)
	}
	status, ok := obj.FieldByName("status").(jtree.String)
	if !ok {
		return nil, errors.New("status field is missing")
	}
	var dest DelegationOperationResult
	switch status {
	case "applied":
		dest = new(DelegationOperationResultApplied)
	case "failed":
		dest = new(DelegationOperationResultFailed)
	case "skipped":
		dest = new(DelegationOperationResultSkipped)
	case "backtracked":
		dest = new(DelegationOperationResultBacktracked)
	default:
		return nil, fmt.Errorf("unknown operation result status: %s", status)
	}
	var tmp interface{} = dest // clear interface type and pass empty interface with pointer
	err := n.Decode(tmp, jtree.OpCtx(ctx))
	return dest, err
}

type DelegationInternalOperationResult struct {
	Kind     OperationKind             `json:"kind"`
	Source   model.Base58              `json:"source"`
	Nonce    uint64                    `json:"nonce"`
	Delegate model.Base58              `json:"delegate,omitempty"`
	Result   DelegationOperationResult `json:"result"`
}

func (d *DelegationInternalOperationResult) OperationKind() OperationKind     { return d.Kind }
func (d *DelegationInternalOperationResult) OperationResult() OperationResult { return d.Result }

type DelegationImplicitOperationResult struct {
	Kind OperationKind `json:"kind"`
	OriginationOperationResultBase
}

func (r *DelegationImplicitOperationResult) OperationKind() OperationKind { return r.Kind }

var (
	_ WithBalanceUpdates          = (*DelegationOperationMetadata)(nil)
	_ OperationMetadataWithResult = (*DelegationOperationMetadata)(nil)
	_ WithConsumedGas             = (*DelegationOperationResultApplied)(nil)
	_ WithConsumedGas             = (*DelegationOperationResultBacktracked)(nil)
	_ WithConsumedGas             = (*DelegationImplicitOperationResult)(nil)
	_ WithErrors                  = (*DelegationOperationResultFailed)(nil)
	_ WithErrors                  = (*DelegationOperationResultBacktracked)(nil)
)

func init() {
	jtree.RegisterType(delegationOperationResultFunc)
}
