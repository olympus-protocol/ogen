package main

import (
	"fmt"
	bls12 "github.com/herumi/bls-eth-go-binary/bls"
	"github.com/olympus-protocol/ogen/internal/keystore"
	"go.etcd.io/bbolt"
	"os"
	"path/filepath"
)

func init() {
	if err := bls12.Init(bls12.BLS12_381); err != nil {
		panic(err)
	}
	if err := bls12.SetETHmode(bls12.EthModeDraft07); err != nil {
		panic(err)
	}
}

func main() {

	files := map[string]string{}
	err := filepath.Walk("./ks", func(path string, info os.FileInfo, err error) error {
		if info != nil {
			if !info.IsDir() {
				files[info.Name()] = path
			}
			return nil
		}
		return nil
	})

	var keys [][]byte
	for _, f := range files {
		db, err := bbolt.Open(f, 0600, nil)
		if err != nil {
			fmt.Println(err)
		}
		err = db.View(func(tx *bbolt.Tx) error {
			bkt := tx.Bucket([]byte("keys"))
			c := bkt.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				keys = append(keys, v)
			}
			return nil
		})
		if err != nil {
			fmt.Println(err)
		}
		_ = db.Close()
	}
	if err != nil {
		fmt.Println(err)
	}
	ks := keystore.NewKeystore("./", nil)
	for _, k := range keys {
		err = ks.AddKey(k)
		if err != nil {
			fmt.Println(err)
		}
	}
	_ = ks.Close()
}
