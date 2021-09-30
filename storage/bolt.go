package storage

import (
	"bytes"
	"context"
	"encoding/gob"
	"os"
	"path/filepath"

	bolt "go.etcd.io/bbolt"
)

const (
	bktProtocolConstants = "protocol_constants"
	bktBlockHeader       = "header"
	bktStat              = "stat"
)

type BoltStorage struct {
	DB *bolt.DB
}

func (b *BoltStorage) GetBlockHeader(ctx context.Context, blockID string) (header *BlockHeader, err error) {
	err = b.DB.View(func(tx *bolt.Tx) error {
		data := tx.Bucket([]byte(bktBlockHeader)).Get([]byte(blockID))
		if data == nil {
			return nil
		}
		header = new(BlockHeader)
		return gob.NewDecoder(bytes.NewReader(data)).Decode(header)
	})
	return
}

func (b *BoltStorage) UpdateBlockHeader(ctx context.Context, header *BlockHeader) error {
	return b.DB.Update(func(tx *bolt.Tx) error {
		var data bytes.Buffer
		if err := gob.NewEncoder(&data).Encode(header); err != nil {
			return err
		}
		return tx.Bucket([]byte(bktBlockHeader)).Put([]byte(header.Hash), data.Bytes())
	})
}

func (b *BoltStorage) GetProtocolConstants(ctx context.Context, chainID string) (constants *ProtocolConstants, err error) {
	err = b.DB.View(func(tx *bolt.Tx) error {
		data := tx.Bucket([]byte(bktProtocolConstants)).Get([]byte(chainID))
		if data == nil {
			return nil
		}
		constants = new(ProtocolConstants)
		return gob.NewDecoder(bytes.NewReader(data)).Decode(constants)
	})
	return
}

func (b *BoltStorage) UpdateProtocolConstants(ctx context.Context, chainID string, constants *ProtocolConstants) error {
	return b.DB.Update(func(tx *bolt.Tx) error {
		var data bytes.Buffer
		if err := gob.NewEncoder(&data).Encode(constants); err != nil {
			return err
		}
		return tx.Bucket([]byte(bktProtocolConstants)).Put([]byte(chainID), data.Bytes())
	})
}

func (b *BoltStorage) GetBlockStat(ctx context.Context, blockID string) (s *BlockStat, err error) {
	err = b.DB.View(func(tx *bolt.Tx) error {
		data := tx.Bucket([]byte(bktStat)).Get([]byte(blockID))
		if data == nil {
			return nil
		}
		s = new(BlockStat)
		return gob.NewDecoder(bytes.NewReader(data)).Decode(s)
	})
	return
}

func (b *BoltStorage) UpdateBlockStat(ctx context.Context, blockID string, s *BlockStat) error {
	return b.DB.Update(func(tx *bolt.Tx) error {
		var data bytes.Buffer
		if err := gob.NewEncoder(&data).Encode(s); err != nil {
			return err
		}
		return tx.Bucket([]byte(bktStat)).Put([]byte(blockID), data.Bytes())
	})
}

const dbFile = ".tezos-grafana-datasource/block_cache.db"

func NewBoltStorage() (*BoltStorage, error) {
	name := filepath.Join(os.Getenv("HOME"), dbFile)
	if err := os.MkdirAll(filepath.Dir(name), 0777); err != nil {
		return nil, err
	}
	db, err := bolt.Open(name, 0666, nil)
	if err != nil {
		return nil, err
	}

	// create buckets
	if err := db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(bktProtocolConstants)); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(bktBlockHeader)); err != nil {
			return err
		}
		_, err := tx.CreateBucketIfNotExists([]byte(bktStat))
		return err
	}); err != nil {
		return nil, err
	}

	return &BoltStorage{db}, nil
}

func (b *BoltStorage) Close() error {
	return b.DB.Close()
}

var _ BlockInfoStorage = (*BoltStorage)(nil)
