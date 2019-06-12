package main

import (
	"fmt"
	"os"

	"github.com/boltdb/bolt"
)

const dbFile = "blockChain.db"
const blockChainBucket = "bucket"
const lastHashKey = "key"

type BlockChain struct {
	// 数据库文件的具柄
	db *bolt.DB
	// 尾巴，表示最后一个区块的哈希值
	tail []byte
}

func NewBlockChain() *BlockChain {
	//Open(path string, mode os.FileMode, options *Options) (*DB, error)
	db, err := bolt.Open(dbFile, 0600, nil)
	IfError("bolt.Open(dbFile, 0600, nil)", err)

	var lastHash []byte

	db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockChainBucket))
		if bucket != nil {
			//取出最后区块的哈希值返回
			lastHash = bucket.Get([]byte(lastHashKey))
		} else {
			//没有bucket,
			genesis := NewGenesisBlock()
			bucket, err = tx.CreateBucket([]byte(blockChainBucket)) //(*Bucket, error)
			IfError("tx.CreateBucket([]byte(blockChainBucket))", err)
			key := genesis.Hash
			value := genesis.Serialize()
			err = bucket.Put(key, value)
			IfError("bucket.Put(key, value)", err)
			err = bucket.Put([]byte(lastHashKey), genesis.Hash)
			IfError("bucket.Put([]byte(lastHashKey), genesis.Hash)", err)
			lastHash = genesis.Hash
		}
		return nil
	})

	return &BlockChain{
		db:   db,
		tail: lastHash,
	}
}

func (this *BlockChain) AddBlock(data string) {
	var lastHash []byte

	err := this.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockChainBucket))
		if bucket != nil {
			//取出最后区块的哈希值返回
			lastHash = bucket.Get([]byte(lastHashKey))
			newBlock := NewBlock(data, lastHash)
			key := newBlock.Hash
			value := newBlock.Serialize()
			err := bucket.Put(key, value)
			IfError("bucket.Put(key, value)", err)
			lastHash = newBlock.Hash
			err = bucket.Put([]byte(lastHashKey), lastHash)
			IfError("bucket.Put([]byte(lastHashKey), lastHash)", err)
			this.tail = lastHash
		} else {
			fmt.Println("tx.Bucket([]byte(blockChainBucket)) return nil")
			os.Exit(1)
		}
		return nil
	})
	IfError("this.db.Update", err)
}

//迭代器，就是一个对象，它里面包含了一个游标，一直向前（后）移动，完成整个容器的遍历
type BlockChainIterator struct {
	CurrHash []byte
	DB       *bolt.DB
}

//创建迭代器，同时初始化为指向最后一个区块
func (this *BlockChain) NewBlockChainIterator() *BlockChainIterator {
	blockChainIterator := &BlockChainIterator{
		CurrHash: this.tail,
		DB:       this.db,
	}
	return blockChainIterator
}

//迭代器向后移动，并返回移动前迭代器所指向的数据
func (this *BlockChainIterator) Next() (block *Block) {
	if this.CurrHash == nil {
		fmt.Println("this.CurrHash == nil")
		return
	}
	err := this.DB.View(func(tx *bolt.Tx) (err error) {
		bucket := tx.Bucket([]byte(blockChainBucket))
		if bucket == nil {
			fmt.Println("没有你要查找的bucket")
			os.Exit(0)
		} else {
			blockHash := bucket.Get(this.CurrHash)
			block = DeSerialize(blockHash)
			this.CurrHash = block.PrevBlockHash
		}
		return
	})
	IfError("this.DB.View()", err)
	return
}
