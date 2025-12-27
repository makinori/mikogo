package db

import (
	"go.etcd.io/bbolt"
)

var (
	db *bbolt.DB
)

func Init() error {
	var err error
	db, err = bbolt.Open("data.db", 0644, &bbolt.Options{})
	if err != nil {
		return err
	}

	db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(Servers.bucket))
		if err != nil {
			return err
		}
		return nil
	})

	return nil
}
