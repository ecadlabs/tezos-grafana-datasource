package bolt

import (
	"context"
	"os"
	"path/filepath"

	"github.com/ecadlabs/tezos-grafana-datasource/pkg/model"
	"github.com/ecadlabs/tezos-grafana-datasource/pkg/storage"
)

const (
	bktBlockInfo = "block_info"
)

type BoltStorage struct {
	DB *DB
}

func (b *BoltStorage) GetBlockInfo(ctx context.Context, blockID model.Base58) (info *model.BlockInfo, err error) {
	err = b.DB.View(func(tx *Tx) error {
		var ok bool
		i := new(model.BlockInfo)
		ok, err = tx.Bucket([]byte(bktBlockInfo)).Get(blockID, i)
		if !ok {
			return nil
		}
		if err != nil {
			return err
		}
		info = i
		return nil
	})
	return
}

func (b *BoltStorage) UpdateBlockInfo(ctx context.Context, info *model.BlockInfo) error {
	return b.DB.Update(func(tx *Tx) error {
		return tx.Bucket([]byte(bktBlockInfo)).Put(info.Header.Hash, info)
	})
}

const defaultDBFile = ".tezos-grafana-datasource/block_cache.db"

func NewBoltStorage(path string) (*BoltStorage, error) {
	if path == "" {
		path = filepath.Join(os.Getenv("HOME"), defaultDBFile)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return nil, err
	}
	db, err := Open(path, 0666, nil, nil)
	if err != nil {
		return nil, err
	}

	// create buckets
	if err := db.Update(func(tx *Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bktBlockInfo))
		return err
	}); err != nil {
		return nil, err
	}

	return &BoltStorage{db}, nil
}

func (b *BoltStorage) Close() error {
	return b.DB.Close()
}

var _ storage.BlockInfoStorage = (*BoltStorage)(nil)
