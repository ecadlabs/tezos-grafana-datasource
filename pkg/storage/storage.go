package storage

import (
	"context"

	"github.com/ecadlabs/tezos-grafana-datasource/pkg/model"
	"github.com/ecadlabs/tezos-grafana-datasource/pkg/model/block"
)

type BlockInfoStorage interface {
	GetBlockInfo(ctx context.Context, blockID model.Base58) (s *block.Info, err error)
	UpdateBlockInfo(ctx context.Context, s *block.Info) error
}
