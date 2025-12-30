package db

import (
	"github.com/makinori/mikogo/env"
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

	// ensure home server exists and has address set to env.
	// should never be deleted or have its address modified.
	homeServer, err, _ := Servers.Get("home")
	if err != nil {
		return err
	}
	homeServer.Address = env.HOME_SERVER
	err = Servers.Put("home", homeServer)
	if err != nil {
		return err
	}

	return nil
}
