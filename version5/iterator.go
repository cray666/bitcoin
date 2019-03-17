package main

import (
	"github.com/boltdb/bolt"
)

type Iterator struct {
	currentHash []byte
	db          *bolt.DB
}


func (bc *Blockchain) Iterator() *Iterator {
	bci := &Iterator{bc.tip, bc.db}

	return bci
}

func (i *Iterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)

		return nil
	})

	handleError(err)

	i.currentHash = block.PrevBlockHash

	return block
}