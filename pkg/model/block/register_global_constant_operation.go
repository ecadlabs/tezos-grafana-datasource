package block

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ecadlabs/jtree"
	"github.com/ecadlabs/tezos-grafana-datasource/pkg/model"
)

type RegisterGlobalConstantOperationContents struct {
	Kind         OperationKind                            `json:"kind"`
	Source       model.Base58                             `json:"source"`
	Fee          big.Int                                  `json:"fee,string"`
	Counter      big.Int                                  `json:"counter,string"`
	GasLimit     big.Int                                  `json:"gas_limit,string"`
	StorageLimit big.Int                                  `json:"storage_limit,string"`
	Value        Expr                                     `json:"value"`
	Metadata     *RegisterGlobalConstantOperationMetadata `json:"metadata,omitempty"`
}

func (r *RegisterGlobalConstantOperationContents) OperationKind() OperationKind { return r.Kind }
func (r *RegisterGlobalConstantOperationContents) OperationMetadata() OperationMetadata {
	if r.Metadata != nil {
		return r.Metadata
	}
	return nil
}

type RegisterGlobalConstantOperationMetadata struct {
	BalanceUpdates           BalanceUpdates                        `json:"balance_updates"`
	OperationResult          RegisterGlobalConstantOperationResult `json:"operation_result"`
	InternalOperationResults InternalOperationResults              `json:"internal_operation_results,omitempty"`
}

func (r *RegisterGlobalConstantOperationMetadata) GetBalanceUpdates() BalanceUpdates {
	return r.BalanceUpdates
}

func (r *RegisterGlobalConstantOperationMetadata) GetResult() (OperationResult, InternalOperationResults) {
	return r.OperationResult, r.InternalOperationResults
}

type RegisterGlobalConstantOperationResult interface {
	OperationResult
	RegisterGlobalConstantOperationResult()
}

type RegisterGlobalConstantOperationResultBase struct {
	BalanceUpdates BalanceUpdates `json:"balance_updates"`
	ConsumedGas    big.Int        `json:"consumed_gas,string"`
	StorageSize    big.Int        `json:"storage_size,string"`
	GlobalAddress  model.Base58   `json:"global_address"`
}

func (r *RegisterGlobalConstantOperationResultBase) GetConsumedGas() (gas, milligas *big.Int) {
	return &r.ConsumedGas, nil
}

type RegisterGlobalConstantOperationResultApplied struct {
	Status OperationStatus `json:"status"`
	RegisterGlobalConstantOperationResultBase
}

func (r *RegisterGlobalConstantOperationResultApplied) RegisterGlobalConstantOperationResult() {}
func (r *RegisterGlobalConstantOperationResultApplied) GetStatus() OperationStatus             { return r.Status }
func (r *RegisterGlobalConstantOperationResultApplied) GetBalanceUpdates() BalanceUpdates {
	return r.BalanceUpdates
}

type RegisterGlobalConstantOperationResultBacktracked struct {
	Status OperationStatus `json:"status"`
	Errors []*model.Error  `json:"errors,omitempty"`
	RegisterGlobalConstantOperationResultBase
}

func (r *RegisterGlobalConstantOperationResultBacktracked) RegisterGlobalConstantOperationResult() {}
func (r *RegisterGlobalConstantOperationResultBacktracked) GetStatus() OperationStatus {
	return r.Status
}
func (r *RegisterGlobalConstantOperationResultBacktracked) GetErrors() []*model.Error {
	return r.Errors
}
func (r *RegisterGlobalConstantOperationResultBacktracked) GetBalanceUpdates() BalanceUpdates {
	return r.BalanceUpdates
}

type RegisterGlobalConstantOperationResultFailed OperationResultFailed

func (r *RegisterGlobalConstantOperationResultFailed) RegisterGlobalConstantOperationResult() {}
func (r *RegisterGlobalConstantOperationResultFailed) GetStatus() OperationStatus             { return r.Status }
func (r *RegisterGlobalConstantOperationResultFailed) GetErrors() []*model.Error              { return r.Errors }

type RegisterGlobalConstantOperationResultSkipped OperationResultSkipped

func (r *RegisterGlobalConstantOperationResultSkipped) RegisterGlobalConstantOperationResult() {}
func (r *RegisterGlobalConstantOperationResultSkipped) GetStatus() OperationStatus             { return r.Status }

func registerGlobalConstantOperationResultFunc(n jtree.Node, ctx *jtree.Context) (RegisterGlobalConstantOperationResult, error) {
	obj, ok := n.(jtree.Object)
	if !ok {
		return nil, fmt.Errorf("object expected: %t", n)
	}
	status, ok := obj.FieldByName("status").(jtree.String)
	if !ok {
		return nil, errors.New("status field is missing")
	}
	var dest RegisterGlobalConstantOperationResult
	switch status {
	case "applied":
		dest = new(RegisterGlobalConstantOperationResultApplied)
	case "failed":
		dest = new(RegisterGlobalConstantOperationResultFailed)
	case "skipped":
		dest = new(RegisterGlobalConstantOperationResultSkipped)
	case "backtracked":
		dest = new(RegisterGlobalConstantOperationResultBacktracked)
	default:
		return nil, fmt.Errorf("unknown operation result status: %s", status)
	}
	var tmp interface{} = dest // clear interface type and pass empty interface with pointer
	err := n.Decode(tmp, jtree.OpCtx(ctx))
	return dest, err
}

type RegisterGlobalConstantInternalOperationResult struct {
	Kind   OperationKind                         `json:"kind"`
	Source model.Base58                          `json:"source"`
	Nonce  uint64                                `json:"nonce"`
	Value  Expr                                  `json:"value"`
	Result RegisterGlobalConstantOperationResult `json:"result"`
}

func (r *RegisterGlobalConstantInternalOperationResult) OperationKind() OperationKind { return r.Kind }
func (r *RegisterGlobalConstantInternalOperationResult) OperationResult() OperationResult {
	return r.Result
}

var (
	_ WithBalanceUpdates          = (*RegisterGlobalConstantOperationMetadata)(nil)
	_ WithBalanceUpdates          = (*RegisterGlobalConstantOperationResultApplied)(nil)
	_ WithBalanceUpdates          = (*RegisterGlobalConstantOperationResultBacktracked)(nil)
	_ OperationMetadataWithResult = (*RegisterGlobalConstantOperationMetadata)(nil)
	_ WithConsumedGas             = (*RegisterGlobalConstantOperationResultApplied)(nil)
	_ WithConsumedGas             = (*RegisterGlobalConstantOperationResultBacktracked)(nil)
	_ WithErrors                  = (*RegisterGlobalConstantOperationResultFailed)(nil)
	_ WithErrors                  = (*RegisterGlobalConstantOperationResultBacktracked)(nil)
)

func init() {
	jtree.RegisterType(registerGlobalConstantOperationResultFunc)
}
