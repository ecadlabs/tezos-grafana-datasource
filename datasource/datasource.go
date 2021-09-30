package datasource

import (
	"context"
	"time"

	"github.com/ecadlabs/tezos-grafana-datasource/client"
	"github.com/ecadlabs/tezos-grafana-datasource/storage"
)

type Datasource struct {
	DB     storage.BlockInfoStorage
	Client *client.Client
}

type BlockInfo struct {
	*storage.BlockHeader
	*storage.BlockStat
	PredecessorTimestamp time.Time
}

func (d *Datasource) GetBlockTimes(ctx context.Context, start, end time.Time) ([]*BlockInfo, error) {
	var (
		blocks    []*BlockInfo
		prevBlock *BlockInfo
	)
	nextBlock := "head"
	for {
		var (
			header *storage.BlockHeader
			err    error
		)
		if nextBlock != "head" {
			header, err = d.DB.GetBlockHeader(ctx, nextBlock)
			if err != nil {
				return nil, err
			}
		}
		if header == nil {
			header, err = d.Client.GetBlockHeader(ctx, nextBlock)
			if err != nil {
				return nil, err
			}
			if err = d.DB.UpdateBlockHeader(ctx, header); err != nil {
				return nil, err
			}
		}

		info := &BlockInfo{
			BlockHeader: header,
		}

		if prevBlock != nil {
			prevBlock.PredecessorTimestamp = header.Timestamp
		}

		if header.Timestamp.Before(start) {
			break
		}

		prevBlock = info
		nextBlock = header.Predecessor

		if !header.Timestamp.Before(end) {
			continue
		}

		stat, err := d.DB.GetBlockStat(ctx, header.Hash)
		if err != nil {
			return nil, err
		}

		if stat == nil {
			ops, err := d.Client.GetBlockOperations(ctx, header.Hash)
			if err != nil {
				return nil, err
			}

			// calculate endorsing power
			var slots int
			for _, tmp := range ops {
				for _, operation := range tmp {
					for _, contents := range operation.Contents {
						switch e := contents.(type) {
						case *storage.EndorsementWithSlot:
							if e.Metadata != nil {
								slots += len(e.Metadata.Slots)
							}
						case *storage.Endorsement:
							if e.Metadata != nil {
								slots += len(e.Metadata.Slots)
							}
						}
					}
				}
			}

			ts, err := d.Client.GetMinimalValidTime(ctx, header.Predecessor, int(header.Priority), slots)
			if err != nil {
				return nil, err
			}

			stat = &storage.BlockStat{
				MinimalValidTime: ts,
				EndorsementSlots: uint(slots),
			}

			if err = d.DB.UpdateBlockStat(ctx, header.Hash, stat); err != nil {
				return nil, err
			}
		}

		info.BlockStat = stat
		blocks = append(blocks, info)
	}

	// reverse
	res := make([]*BlockInfo, len(blocks))
	for i, b := range blocks {
		res[len(blocks)-i-1] = b
	}
	return res, nil
}
