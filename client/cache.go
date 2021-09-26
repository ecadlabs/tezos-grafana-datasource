package client

import (
	"bytes"
	"encoding/gob"
	"os"
	"path/filepath"

	bolt "go.etcd.io/bbolt"
)

const dbFile = ".tezos-grafana-datasource/block_header_cache.db"

type Cache struct {
	DB     *bolt.DB
	Client *Client
}

func NewDB() (*bolt.DB, error) {
	name := filepath.Join(os.Getenv("HOME"), dbFile)
	if err := os.MkdirAll(filepath.Dir(name), 0777); err != nil {
		return nil, err
	}
	return bolt.Open(name, 0666, nil)
}

func (c *Cache) GetBlockHeader(chainID, blockID string) (header *BlockHeader, err error) {
	if chainID != "" && blockID != "head" {
		err = c.DB.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte("block_header_" + chainID))
			if err != nil {
				return err
			}

			if data := b.Get([]byte(blockID)); data != nil {
				header = new(BlockHeader)
				return gob.NewDecoder(bytes.NewReader(data)).Decode(header)
			} else {
				var err error
				header, err = c.Client.GetBlockHeader(blockID)
				if err != nil {
					return err
				}
				var data bytes.Buffer
				if err := gob.NewEncoder(&data).Encode(header); err != nil {
					return err
				}
				return b.Put([]byte(blockID), data.Bytes())
			}
		})
		return
	}

	header, err = c.Client.GetBlockHeader(blockID)
	if err != nil {
		return
	}

	chainID = header.ChainID.String()
	blockID = header.Hash.String()

	var data bytes.Buffer
	if err = gob.NewEncoder(&data).Encode(header); err != nil {
		return
	}

	err = c.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("block_header_" + chainID))
		if err != nil {
			return err
		}
		return b.Put([]byte(blockID), data.Bytes())
	})
	return
}
