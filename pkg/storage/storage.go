package storage

import (
	"context"

	"github.com/ecadlabs/tezos-grafana-datasource/pkg/model"
)

type BlockInfoStorage interface {
	GetBlockInfo(ctx context.Context, blockID model.Base58) (s *model.BlockInfo, err error)
	UpdateBlockInfo(ctx context.Context, s *model.BlockInfo) error
}
