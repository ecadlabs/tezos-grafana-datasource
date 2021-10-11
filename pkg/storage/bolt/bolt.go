package bolt

import (
	"fmt"
	"os"
	"reflect"

	bolt "go.etcd.io/bbolt"
)

type Codecs struct {
	Key   Codec
	Value Codec
}

type DB struct {
	codec *Codecs
	*bolt.DB
}

func (db *DB) Begin(writable bool) (*Tx, error) {
	tx, err := db.DB.Begin(writable)
	if err != nil {
		return nil, err
	}
	return &Tx{codec: db.codec, Tx: tx}, nil
}

func (db *DB) View(fn func(*Tx) error) error {
	return db.DB.View(func(tx *bolt.Tx) error {
		return fn(&Tx{codec: db.codec, Tx: tx})
	})
}

func (db *DB) Update(fn func(*Tx) error) error {
	return db.DB.Update(func(tx *bolt.Tx) error {
		return fn(&Tx{codec: db.codec, Tx: tx})
	})
}

func (db *DB) Batch(fn func(*Tx) error) error {
	return db.DB.Batch(func(tx *bolt.Tx) error {
		return fn(&Tx{codec: db.codec, Tx: tx})
	})
}

type Tx struct {
	codec *Codecs
	*bolt.Tx
}

func (tx *Tx) Bucket(name []byte) *Bucket {
	return &Bucket{codec: tx.codec, bucket: tx.Tx.Bucket(name)}
}

func (tx *Tx) CreateBucket(name []byte) (*Bucket, error) {
	b, err := tx.Tx.CreateBucket(name)
	if err != nil {
		return nil, err
	}
	return &Bucket{codec: tx.codec, bucket: b}, nil
}

func (tx *Tx) CreateBucketIfNotExists(name []byte) (*Bucket, error) {
	b, err := tx.Tx.CreateBucketIfNotExists(name)
	if err != nil {
		return nil, err
	}
	return &Bucket{codec: tx.codec, bucket: b}, nil
}

func (tx *Tx) DB() *DB {
	return &DB{codec: tx.codec, DB: tx.Tx.DB()}
}

func (tx *Tx) ForEach(fn func(name []byte, b *Bucket) error) error {
	return tx.Tx.ForEach(func(name []byte, b *bolt.Bucket) error {
		return fn(name, &Bucket{codec: tx.codec, bucket: b})
	})
}

func (tx *Tx) Cursor() *Cursor {
	return &Cursor{codec: tx.codec, cursor: tx.Tx.Cursor()}
}

type Bucket struct {
	codec  *Codecs
	bucket *bolt.Bucket
}

func (b *Bucket) Cursor() *Cursor { return &Cursor{codec: b.codec, cursor: b.bucket.Cursor()} }
func (b *Bucket) Bucket(name []byte) *Bucket {
	return &Bucket{codec: b.codec, bucket: b.bucket.Bucket(name)}
}
func (b *Bucket) DeleteBucket(key []byte) error { return b.bucket.DeleteBucket(key) }
func (b *Bucket) NextSequence() (uint64, error) { return b.bucket.NextSequence() }
func (b *Bucket) Sequence() uint64              { return b.bucket.Sequence() }
func (b *Bucket) SetSequence(v uint64) error    { return b.bucket.SetSequence(v) }
func (b *Bucket) Stats() bolt.BucketStats       { return b.bucket.Stats() }
func (b *Bucket) Tx() *Tx                       { return &Tx{codec: b.codec, Tx: b.bucket.Tx()} }
func (b *Bucket) Writable() bool                { return b.bucket.Writable() }

func (b *Bucket) CreateBucket(key []byte) (*Bucket, error) {
	bb, err := b.bucket.CreateBucket(key)
	if err != nil {
		return nil, err
	}
	return &Bucket{codec: b.codec, bucket: bb}, nil
}

func (b *Bucket) CreateBucketIfNotExists(key []byte) (*Bucket, error) {
	bb, err := b.bucket.CreateBucketIfNotExists(key)
	if err != nil {
		return nil, err
	}
	return &Bucket{codec: b.codec, bucket: bb}, nil
}

func (b *Bucket) Delete(key interface{}) error {
	k, err := b.codec.Key.Marshal(key)
	if err != nil {
		return err
	}
	return b.bucket.Delete(k)
}

func (b *Bucket) ForEach(fn interface{}) error {
	fnv := reflect.ValueOf(fn)
	kt, vt := keyValFunc(fnv.Type())
	return b.bucket.ForEach(func(k, v []byte) error {
		kv := reflect.New(kt)
		if err := b.codec.Key.Unmarshal(k, kv.Interface()); err != nil {
			return err
		}
		vv := reflect.New(vt)
		if err := b.codec.Value.Unmarshal(v, vv.Interface()); err != nil {
			return err
		}
		ret := fnv.Call([]reflect.Value{kv.Elem(), vv.Elem()})
		if !ret[0].IsNil() {
			return ret[0].Interface().(error)
		}
		return nil
	})
}

func (b *Bucket) Get(key, value interface{}) (bool, error) {
	k, err := b.codec.Key.Marshal(key)
	if err != nil {
		return false, err
	}
	data := b.bucket.Get(k)
	if data == nil {
		return false, nil
	}
	if err := b.codec.Value.Unmarshal(data, value); err != nil {
		return false, err
	}
	return true, nil
}

func (b *Bucket) Put(key, value interface{}) error {
	k, err := b.codec.Key.Marshal(key)
	if err != nil {
		return err
	}
	v, err := b.codec.Value.Marshal(value)
	if err != nil {
		return err
	}
	return b.bucket.Put(k, v)
}

type Cursor struct {
	codec  *Codecs
	cursor *bolt.Cursor
}

func (c *Cursor) Bucket() *Bucket { return &Bucket{codec: c.codec, bucket: c.cursor.Bucket()} }
func (c *Cursor) Delete() error   { return c.cursor.Delete() }

func (c *Cursor) First(key, value interface{}) (bool, error) {
	k, v := c.cursor.First()
	if k == nil {
		return false, nil
	}
	if err := c.codec.Key.Unmarshal(k, key); err != nil {
		return false, err
	}
	if err := c.codec.Value.Unmarshal(v, value); err != nil {
		return false, err
	}
	return true, nil
}

func (c *Cursor) Last(key, value interface{}) (bool, error) {
	k, v := c.cursor.Last()
	if k == nil {
		return false, nil
	}
	if err := c.codec.Key.Unmarshal(k, key); err != nil {
		return false, err
	}
	if err := c.codec.Value.Unmarshal(v, value); err != nil {
		return false, err
	}
	return true, nil
}

func (c *Cursor) Next(key, value interface{}) (bool, error) {
	k, v := c.cursor.Next()
	if k == nil {
		return false, nil
	}
	if err := c.codec.Key.Unmarshal(k, key); err != nil {
		return false, err
	}
	if err := c.codec.Value.Unmarshal(v, value); err != nil {
		return false, err
	}
	return true, nil
}

func (c *Cursor) Prev(key, value interface{}) (bool, error) {
	k, v := c.cursor.Prev()
	if k == nil {
		return false, nil
	}
	if err := c.codec.Key.Unmarshal(k, key); err != nil {
		return false, err
	}
	if err := c.codec.Value.Unmarshal(v, value); err != nil {
		return false, err
	}
	return true, nil
}

func (c *Cursor) Seek(seek, key, value interface{}) (bool, error) {
	s, err := c.codec.Key.Marshal(key)
	if err != nil {
		return false, err
	}
	k, v := c.cursor.Seek(s)
	if k == nil {
		return false, nil
	}
	if err := c.codec.Key.Unmarshal(k, key); err != nil {
		return false, err
	}
	if err := c.codec.Value.Unmarshal(v, value); err != nil {
		return false, err
	}
	return true, nil
}

func Open(path string, mode os.FileMode, options *bolt.Options, codecs *Codecs) (*DB, error) {
	if codecs == nil {
		codecs = &Codecs{
			Key:   BinaryCodec{},
			Value: GobCodec{},
		}
	}
	db, err := bolt.Open(path, mode, options)
	if err != nil {
		return nil, err
	}
	return &DB{codec: codecs, DB: db}, nil
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()

func keyValFunc(t reflect.Type) (kt, vt reflect.Type) {
	if t.Kind() != reflect.Func {
		panic(fmt.Sprintf("not a function: %v", t))
	}
	if t.NumIn() != 2 {
		panic(fmt.Sprintf("wrong number of arguments: %d", t.NumIn()))
	}
	if t.NumOut() != 1 {
		panic(fmt.Sprintf("wrong number of return parameters: %d", t.NumOut()))
	}
	if t.Out(1) != errorType {
		panic("function must return an error")
	}
	return t.In(0), t.In(1)
}
