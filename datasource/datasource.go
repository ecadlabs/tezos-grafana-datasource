package datasource

import (
	"context"
	"time"

	"github.com/ecadlabs/tezos-grafana-datasource/client"
	"github.com/ecadlabs/tezos-grafana-datasource/model"
	"github.com/ecadlabs/tezos-grafana-datasource/storage"
)

type Datasource struct {
	DB     storage.BlockInfoStorage
	Client *client.Client
}

type BlockTimeInfo struct {
	*model.BlockInfo
	PredecessorTimestamp time.Time     `json:"predecessor_timestamp"`
	MinDelay             time.Duration `json:"minimal_delay"`
	Delay                time.Duration `json:"delay"`
}

func (d *Datasource) getBlockInfo(ctx context.Context, blockID model.Base58) (*model.BlockInfo, error) {
	info, err := d.DB.GetBlockInfo(ctx, blockID)
	if err != nil {
		return nil, err
	}
	if info != nil {
		return info, nil
	}
	block, err := d.Client.GetBlock(ctx, blockID.String())
	if err != nil {
		return nil, err
	}
	stat := block.Stat()
	ts, err := d.Client.GetMinimalValidTime(ctx, block.Header.Predecessor.String(), int(block.Header.Priority), int(stat.Slots))
	if err != nil {
		return nil, err
	}
	info = &model.BlockInfo{
		Header:       block.GetHeader(),
		Stat:         stat,
		MinValidTime: ts,
	}
	if err = d.DB.UpdateBlockInfo(ctx, info); err != nil {
		return nil, err
	}
	return info, nil
}

func (d *Datasource) GetBlockTimes(ctx context.Context, start, end time.Time) ([]*BlockTimeInfo, error) {
	// get head first
	h, err := d.Client.GetBlockHeader(ctx, "head")
	if err != nil {
		return nil, err
	}
	nextBlock := h.Hash

	var (
		blocks    []*BlockTimeInfo
		prevBlock *BlockTimeInfo
	)
	for {
		i, err := d.getBlockInfo(ctx, nextBlock)
		if err != nil {
			return nil, err
		}
		info := &BlockTimeInfo{
			BlockInfo: i,
		}

		if prevBlock != nil {
			prevBlock.PredecessorTimestamp = info.Header.Timestamp
			prevBlock.Delay = prevBlock.Header.Timestamp.Sub(info.Header.Timestamp)
			prevBlock.MinDelay = prevBlock.MinValidTime.Sub(info.Header.Timestamp)
		}

		if info.Header.Timestamp.Before(start) {
			break
		}

		prevBlock = info
		nextBlock = info.Header.Predecessor

		if !info.Header.Timestamp.Before(end) {
			continue
		}
		blocks = append(blocks, info)
	}

	// reverse
	res := make([]*BlockTimeInfo, len(blocks))
	for i, b := range blocks {
		res[len(blocks)-i-1] = b
	}
	return res, nil
}
