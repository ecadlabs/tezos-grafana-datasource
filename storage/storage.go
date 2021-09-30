package storage

import (
	"context"
	"time"
)

type BlockStat struct {
	MinimalValidTime time.Time
	EndorsementSlots uint
}

type BlockInfoStorage interface {
	GetBlockHeader(ctx context.Context, blockID string) (header *BlockHeader, err error)
	UpdateBlockHeader(ctx context.Context, header *BlockHeader) error
	GetProtocolConstants(ctx context.Context, chainID string) (constants *ProtocolConstants, err error)
	UpdateProtocolConstants(ctx context.Context, chainID string, constants *ProtocolConstants) error
	GetBlockStat(ctx context.Context, blockID string) (s *BlockStat, err error)
	UpdateBlockStat(ctx context.Context, blockID string, s *BlockStat) error
}
