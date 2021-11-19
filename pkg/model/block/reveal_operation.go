package block

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ecadlabs/jtree"
	"github.com/ecadlabs/tezos-grafana-datasource/pkg/model"
)

type RevealOperationContents struct {
	Kind         OperationKind            `json:"kind"`
	Source       model.Base58             `json:"source"`
	Fee          big.Int                  `json:"fee,string"`
	Counter      big.Int                  `json:"counter,string"`
	GasLimit     big.Int                  `json:"gas_limit,string"`
	StorageLimit big.Int                  `json:"storage_limit,string"`
	PublicKey    model.Base58             `json:"public_key"`
	Metadata     *RevealOperationMetadata `json:"metadata,omitempty"`
}

func (r *RevealOperationContents) OperationKind() OperationKind { return r.Kind }
func (r *RevealOperationContents) OperationMetadata() OperationMetadata {
	if r.Metadata != nil {
		return r.Metadata
	}
	return nil
}

type RevealOperationMetadata struct {
	BalanceUpdates           BalanceUpdates           `json:"balance_updates"`
	OperationResult          RevealOperationResult    `json:"operation_result"`
	InternalOperationResults InternalOperationResults `json:"internal_operation_results,omitempty"`
}

func (r *RevealOperationMetadata) GetBalanceUpdates() BalanceUpdates { return r.BalanceUpdates }
func (r *RevealOperationMetadata) GetResult() OperationResult        { return r.OperationResult }
func (r *RevealOperationMetadata) GetInternalOperationResults() InternalOperationResults {
	return r.InternalOperationResults
}

type RevealOperationResult interface {
	OperationResult
	RevealOperationResult()
}

type RevealOperationResultBase struct {
	ConsumedGas      *big.Int `json:"consumed_gas,string,omitempty"`
	ConsumedMilligas *big.Int `json:"consumed_milligas,string,omitempty"`
}

func (r *RevealOperationResultBase) GetConsumedMilligas() *big.Int {
	return getConsumedMilligas(r.ConsumedGas, r.ConsumedMilligas)
}

type RevealOperationResultApplied struct {
	Status OperationStatus `json:"status"`
	RevealOperationResultBase
}

func (r *RevealOperationResultApplied) RevealOperationResult()     {}
func (r *RevealOperationResultApplied) GetStatus() OperationStatus { return r.Status }

type RevealOperationResultBacktracked struct {
	Status OperationStatus `json:"status"`
	Errors []*model.Error  `json:"errors,omitempty"`
	RevealOperationResultBase
}

func (r *RevealOperationResultBacktracked) RevealOperationResult()     {}
func (r *RevealOperationResultBacktracked) GetStatus() OperationStatus { return r.Status }
func (r *RevealOperationResultBacktracked) GetErrors() []*model.Error  { return r.Errors }

type RevealOperationResultFailed OperationResultFailed

func (r *RevealOperationResultFailed) RevealOperationResult()     {}
func (r *RevealOperationResultFailed) GetStatus() OperationStatus { return r.Status }
func (r *RevealOperationResultFailed) GetErrors() []*model.Error  { return r.Errors }

type RevealOperationResultSkipped OperationResultSkipped

func (r *RevealOperationResultSkipped) RevealOperationResult()     {}
func (r *RevealOperationResultSkipped) GetStatus() OperationStatus { return r.Status }

func revealOperationResultFunc(n jtree.Node, ctx *jtree.Context) (RevealOperationResult, error) {
	obj, ok := n.(jtree.Object)
	if !ok {
		return nil, fmt.Errorf("object expected: %t", n)
	}
	status, ok := obj.FieldByName("status").(jtree.String)
	if !ok {
		return nil, errors.New("status field is missing")
	}
	var dest RevealOperationResult
	switch OperationStatus(status) {
	case StatusApplied:
		dest = new(RevealOperationResultApplied)
	case StatusFailed:
		dest = new(RevealOperationResultFailed)
	case StatusSkipped:
		dest = new(RevealOperationResultSkipped)
	case StatusBacktracked:
		dest = new(RevealOperationResultBacktracked)
	default:
		return nil, fmt.Errorf("unknown operation result status: %s", status)
	}
	var tmp interface{} = dest // clear interface type and pass empty interface with pointer
	err := n.Decode(tmp, jtree.OpCtx(ctx))
	return dest, err
}

type RevealInternalOperationResult struct {
	Kind      OperationKind         `json:"kind"`
	Source    model.Base58          `json:"source"`
	Nonce     uint64                `json:"nonce"`
	PublicKey model.Base58          `json:"public_key"`
	Result    RevealOperationResult `json:"result"`
}

func (r *RevealInternalOperationResult) OperationKind() OperationKind { return r.Kind }
func (r *RevealInternalOperationResult) GetResult() OperationResult   { return r.Result }

type RevealImplicitOperationResult struct {
	Kind OperationKind `json:"kind"`
	RevealOperationResultBase
}

func (r *RevealImplicitOperationResult) OperationKind() OperationKind { return r.Kind }

var (
	_ WithBalanceUpdates           = (*RevealOperationMetadata)(nil)
	_ WithInternalOperationResults = (*RevealOperationMetadata)(nil)
	_ WithConsumedMilligas         = (*RevealOperationResultApplied)(nil)
	_ WithConsumedMilligas         = (*RevealOperationResultBacktracked)(nil)
	_ WithConsumedMilligas         = (*RevealImplicitOperationResult)(nil)
	_ WithErrors                   = (*RevealOperationResultFailed)(nil)
	_ WithErrors                   = (*RevealOperationResultBacktracked)(nil)
)

func init() {
	jtree.RegisterType(revealOperationResultFunc)
}
