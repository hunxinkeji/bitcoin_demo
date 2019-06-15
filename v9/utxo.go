package main

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"os"
)

const utxoBucket = "utxo"

//UTOXSet结构（保存指定区块链中的所有utxo）
type UTXOSet struct {
	BlockChain *BlockChain
}

func (utxoSet *UTXOSet) ResetUTXOSet() {
	//更新utxoSet
	//采用覆盖的方式
	err := utxoSet.BlockChain.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket))
		if nil != bucket {
			tx.DeleteBucket([]byte(utxoBucket))
		}

		createBucket, _ := tx.CreateBucket([]byte(utxoBucket))

		if nil != createBucket {
			// 查找所有未花费的输出
			utxoMap := utxoSet.BlockChain.FindUTXOMap()
			// 存入表
			for address, outputs := range utxoMap {
				outputsBytes := outputs.Serialize()
				err := createBucket.Put([]byte(address), outputsBytes)
				IfError("createBucket.Put([]byte(address), outputsBytes)", err)
			}
		}

		return nil
	})
	IfError("utxoSet.BlockChain.db.Update(func(tx *bolt.Tx)", err)
}

func (utxoSet *UTXOSet) FindUTXOsWithAddressFromUTXOSet(address string) (tXOutPuts *SuperUTXOs) {
	// 查找数据库中的utxoBucket
	fmt.Println("FindUTXOsWithAddressFromUTXOSet")
	err := utxoSet.BlockChain.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(utxoBucket))
		if nil == bucket {
			fmt.Println("tx.Bucket([]byte(utxoBucket))==nil")
			os.Exit(1)
		}
		data := bucket.Get([]byte(address))
		tXOutPuts = DeSerializeTXOutPuts(data)
		return nil
	})
	IfError("utxoSet.BlockChain.db.View", err)
	return
}

type SuperOutPut struct {
	// 所引用输出的交易ID
	TXID []byte
	// 所引用output的索引值
	Vout   int64
	Output TXOutput
}

type SuperUTXOs struct {
	UTXOs []SuperOutPut
}

// 获取余额
func (utxoSet *UTXOSet) GetBalance(address string) (balance float64) {
	tXOutPuts := utxoSet.FindUTXOsWithAddressFromUTXOSet(address)
	var utxos []SuperOutPut
	utxos = tXOutPuts.UTXOs

	for _, utxo := range utxos {
		balance += utxo.Output.Value
	}
	fmt.Println("(utxoSet *UTXOSet) GetBalance(address string) (balance float64)")
	return
}

func (this *SuperUTXOs) Serialize() (bytes []byte) {
	bytes, err := json.Marshal(this)
	IfError("json.Marshal(this)", err)
	return
}

func DeSerializeTXOutPuts(data []byte) (tXOutPuts *SuperUTXOs) {
	tXOutPuts = &SuperUTXOs{}
	err := json.Unmarshal(data, tXOutPuts)
	IfError("哈哈json.Unmarshal(data, block)", err)
	return
}
