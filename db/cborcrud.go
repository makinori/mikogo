package db

import (
	"errors"

	"github.com/elliotchance/orderedmap/v3"
	"github.com/fxamacker/cbor/v2"
	"go.etcd.io/bbolt"
)

type cborCrud[T any] struct {
	bucket string
}

func (c *cborCrud[T]) GetAll() (*orderedmap.OrderedMap[string, T], error) {
	all := orderedmap.NewOrderedMap[string, T]()

	err := db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(c.bucket))
		if bucket == nil {
			return errors.New(c.bucket + " bucket not found")
		}

		return bucket.ForEach(func(key, data []byte) error {
			var value T
			err := cbor.Unmarshal(data, &value)
			if err != nil {
				return err
			}
			all.Set(string(key), value)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	return all, nil
}

func (c *cborCrud[T]) Get(key string) (value T, err error) {
	err = db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(c.bucket))
		if bucket == nil {
			return errors.New(c.bucket + " bucket not found")
		}

		data := bucket.Get([]byte(key))
		if len(data) == 0 {
			return errors.New("not found")
		}

		return cbor.Unmarshal(data, &value)
	})
	return
}

func (c *cborCrud[T]) Put(key string, value T) error {
	data, err := cbor.Marshal(value)
	if err != nil {
		return err
	}

	return db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(c.bucket))
		if bucket == nil {
			return errors.New(c.bucket + " bucket not found")
		}

		return bucket.Put([]byte(key), data)
	})
}

func (c *cborCrud[T]) Add(key string, value T) error {
	data, err := cbor.Marshal(value)
	if err != nil {
		return err
	}

	return db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(c.bucket))
		if bucket == nil {
			return errors.New(c.bucket + " bucket not found")
		}

		if len(bucket.Get([]byte(key))) > 0 {
			return errors.New("already exists")
		}

		return bucket.Put([]byte(key), data)
	})
}

func (c *cborCrud[T]) Delete(key string) error {
	return db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(c.bucket))
		if bucket == nil {
			return errors.New(c.bucket + " bucket not found")
		}

		if len(bucket.Get([]byte(key))) == 0 {
			return errors.New("not found")
		}

		return bucket.Delete([]byte(key))
	})
}
