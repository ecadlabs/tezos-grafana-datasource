package datasource

import (
	"context"
	"time"

	"github.com/ecadlabs/tezos-grafana-datasource/pkg/client"
	"github.com/ecadlabs/tezos-grafana-datasource/pkg/model"
	"github.com/ecadlabs/tezos-grafana-datasource/pkg/storage"
)

type Datasource struct {
	DB     storage.BlockInfoStorage
	Client *client.Client
}

type BlockInfo struct {
	*model.BlockInfo
	PredecessorTimestamp time.Time `json:"predecessor_timestamp"`
	MinDelay             int64     `json:"minimal_delay"`
	Delay                int64     `json:"delay"`
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

func (d *Datasource) GetBlocksInfo(ctx context.Context, start, end time.Time) ([]*BlockInfo, error) {
	// get head first
	h, err := d.Client.GetBlockHeader(ctx, "head")
	if err != nil {
		return nil, err
	}
	nextBlock := h.Hash

	var (
		blocks    []*BlockInfo
		prevBlock *BlockInfo
	)
	for {
		i, err := d.getBlockInfo(ctx, nextBlock)
		if err != nil {
			return nil, err
		}
		info := &BlockInfo{
			BlockInfo: i,
		}

		if prevBlock != nil {
			prevBlock.PredecessorTimestamp = info.Header.Timestamp
			prevBlock.Delay = int64(prevBlock.Header.Timestamp.Sub(info.Header.Timestamp))
			prevBlock.MinDelay = int64(prevBlock.MinValidTime.Sub(info.Header.Timestamp))
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
	res := make([]*BlockInfo, len(blocks))
	for i, b := range blocks {
		res[len(blocks)-i-1] = b
	}
	return res, nil
}

func (d *Datasource) MonitorBlockInfo(ctx context.Context) (blockInfo <-chan *BlockInfo, errors <-chan error, err error) {
	var cancelFunc context.CancelFunc
	ctx, cancelFunc = context.WithCancel(ctx)
	headerCh, clientErrCh, err := d.Client.GetMonitorHeads(ctx)
	if err != nil {
		cancelFunc()
		return nil, nil, err
	}
	blockinfoCh := make(chan *BlockInfo, 100)
	errorsCh := make(chan error, 1)
	go (func() {
		defer (func() {
			close(blockinfoCh)
			close(errorsCh)
			cancelFunc()
		})()

		var err error
	headerLoop:
		for h := range headerCh {
			var bi, pred *model.BlockInfo
			if bi, err = d.getBlockInfo(ctx, h.Hash); err != nil {
				break
			}
			if pred, err = d.getBlockInfo(ctx, h.Predecessor); err != nil {
				break
			}

			blockinfo := &BlockInfo{
				BlockInfo:            bi,
				PredecessorTimestamp: pred.Header.Timestamp,
				Delay:                int64(bi.Header.Timestamp.Sub(pred.Header.Timestamp)),
				MinDelay:             int64(bi.MinValidTime.Sub(pred.Header.Timestamp)),
			}

			select {
			case blockinfoCh <- blockinfo:
			case <-ctx.Done():
				err = ctx.Err()
				break headerLoop
			}
		}

		if err != nil {
			errorsCh <- err
		} else if err, ok := <-clientErrCh; ok {
			errorsCh <- err
		}
	})()
	return blockinfoCh, errorsCh, nil
}
