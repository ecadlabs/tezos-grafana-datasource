package block

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ecadlabs/jtree"
	"github.com/ecadlabs/tezos-grafana-datasource/pkg/model"
)

type TransactionOperationContents struct {
	Kind         OperationKind                 `json:"kind"`
	Source       model.Base58                  `json:"source"`
	Fee          big.Int                       `json:"fee,string"`
	Counter      big.Int                       `json:"counter,string"`
	GasLimit     big.Int                       `json:"gas_limit,string"`
	StorageLimit big.Int                       `json:"storage_limit,string"`
	Amount       big.Int                       `json:"amount,string"`
	Destination  model.Base58                  `json:"destination"`
	Parameters   *Parameters                   `json:"parameters,omitempty"`
	Metadata     *TransactionOperationMetadata `json:"metadata,omitempty"`
}

func (t *TransactionOperationContents) OperationKind() OperationKind { return t.Kind }
func (t *TransactionOperationContents) OperationMetadata() OperationMetadata {
	if t.Metadata != nil {
		return t.Metadata
	}
	return nil
}

type Parameters struct {
	Entrypoint string `json:"entrypoint"`
	Value      Expr   `json:"value"`
}

type TransactionOperationMetadata struct {
	BalanceUpdates           BalanceUpdates             `json:"balance_updates"`
	OperationResult          TransactionOperationResult `json:"operation_result"`
	InternalOperationResults InternalOperationResults   `json:"internal_operation_results,omitempty"`
}

func (t *TransactionOperationMetadata) GetBalanceUpdates() BalanceUpdates { return t.BalanceUpdates }
func (t *TransactionOperationMetadata) GetResult() OperationResult        { return t.OperationResult }
func (t *TransactionOperationMetadata) GetInternalOperationResults() InternalOperationResults {
	return t.InternalOperationResults
}

type TransactionOperationResult interface {
	OperationResult
	TransactionOperationResult()
}

type TransactionOperationResultBase struct {
	Storage                      Expr           `json:"storage,omitempty"`
	BigMapDiff                   jtree.Node     `json:"big_map_diff,omitempty"`
	BalanceUpdates               BalanceUpdates `json:"balance_updates,omitempty"`
	OriginatedContracts          []model.Base58 `json:"originated_contracts,omitempty"`
	ConsumedGas                  *big.Int       `json:"consumed_gas,string,omitempty"`
	ConsumedMilligas             *big.Int       `json:"consumed_milligas,string,omitempty"`
	StorageSize                  *big.Int       `json:"storage_size,string,omitempty"`
	PaidStorageSizeDiff          *big.Int       `json:"paid_storage_size_diff,string,omitempty"`
	AllocatedDestinationContract *bool          `json:"allocated_destination_contract,omitempty"`
	LazyStorageDiff              jtree.Node     `json:"lazy_storage_diff,omitempty"`
}

func (r *TransactionOperationResultBase) GetConsumedMilligas() *big.Int {
	return getConsumedMilligas(r.ConsumedGas, r.ConsumedMilligas)
}

func (r *TransactionOperationResultBase) GetStorageSize() *big.Int {
	return r.StorageSize
}

type TransactionOperationResultApplied struct {
	Status OperationStatus `json:"status"`
	TransactionOperationResultBase
}

func (t *TransactionOperationResultApplied) TransactionOperationResult() {}
func (t *TransactionOperationResultApplied) GetStatus() OperationStatus  { return t.Status }

type TransactionOperationResultBacktracked struct {
	Status OperationStatus `json:"status"`
	Errors []*model.Error  `json:"errors,omitempty"`
	TransactionOperationResultBase
}

func (t *TransactionOperationResultBacktracked) TransactionOperationResult() {}
func (t *TransactionOperationResultBacktracked) GetStatus() OperationStatus  { return t.Status }
func (t *TransactionOperationResultBacktracked) GetErrors() []*model.Error   { return t.Errors }

type TransactionOperationResultFailed OperationResultFailed

func (r *TransactionOperationResultFailed) TransactionOperationResult() {}
func (r *TransactionOperationResultFailed) GetStatus() OperationStatus  { return r.Status }
func (r *TransactionOperationResultFailed) GetErrors() []*model.Error   { return r.Errors }

type TransactionOperationResultSkipped OperationResultSkipped

func (r *TransactionOperationResultSkipped) TransactionOperationResult() {}
func (r *TransactionOperationResultSkipped) GetStatus() OperationStatus  { return r.Status }

func transactionOperationResultFunc(n jtree.Node, ctx *jtree.Context) (TransactionOperationResult, error) {
	obj, ok := n.(jtree.Object)
	if !ok {
		return nil, fmt.Errorf("object expected: %t", n)
	}
	status, ok := obj.FieldByName("status").(jtree.String)
	if !ok {
		return nil, errors.New("status field is missing")
	}
	var dest TransactionOperationResult
	switch status {
	case "applied":
		dest = new(TransactionOperationResultApplied)
	case "failed":
		dest = new(TransactionOperationResultFailed)
	case "skipped":
		dest = new(TransactionOperationResultSkipped)
	case "backtracked":
		dest = new(TransactionOperationResultBacktracked)
	default:
		return nil, fmt.Errorf("unknown operation result status: %s", status)
	}
	var tmp interface{} = dest // clear interface type and pass empty interface with pointer
	err := n.Decode(tmp, jtree.OpCtx(ctx))
	return dest, err
}

type TransactionInternalOperationResult struct {
	Kind        OperationKind              `json:"kind"`
	Source      model.Base58               `json:"source"`
	Nonce       uint64                     `json:"nonce"`
	Amount      big.Int                    `json:"amount,string"`
	Destination model.Base58               `json:"destination"`
	Parameters  *Parameters                `json:"parameters,omitempty"`
	Result      TransactionOperationResult `json:"result"`
}

func (t *TransactionInternalOperationResult) OperationKind() OperationKind { return t.Kind }
func (t *TransactionInternalOperationResult) GetResult() OperationResult   { return t.Result }

type TransactionImplicitOperationResult struct {
	Kind OperationKind `json:"kind"`
	TransactionOperationResultBase
}

func (r *TransactionImplicitOperationResult) OperationKind() OperationKind { return r.Kind }

var (
	_ WithBalanceUpdates   = (*TransactionOperationMetadata)(nil)
	_ WithResult           = (*TransactionOperationMetadata)(nil)
	_ WithConsumedMilligas = (*TransactionOperationResultApplied)(nil)
	_ WithConsumedMilligas = (*TransactionOperationResultBacktracked)(nil)
	_ WithConsumedMilligas = (*TransactionImplicitOperationResult)(nil)
	_ WithStorage          = (*TransactionOperationResultApplied)(nil)
	_ WithStorage          = (*TransactionOperationResultBacktracked)(nil)
	_ WithStorage          = (*TransactionImplicitOperationResult)(nil)
	_ WithErrors           = (*TransactionOperationResultFailed)(nil)
	_ WithErrors           = (*TransactionOperationResultBacktracked)(nil)
)

func init() {
	jtree.RegisterType(transactionOperationResultFunc)
}
