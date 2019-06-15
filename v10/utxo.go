package main

import (
	"bytes"
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

// 实现UTXOSet实时更新
func (utxoSet *UTXOSet) update() {
	fmt.Println("(utxoSet *UTXOSet) update()")
	// 找到需要删除的UTXO
	//1. 获取最新的区块
	latest_block := utxoSet.BlockChain.NewBlockChainIterator().Next()
	var inputs []TXInput // 存放最新区块的所有输入
	// 获取需要存入utxo set中的[]SuperOutPut
	var addUtxoSlice []SuperOutPut

	//2.查找需要删除的数据
	for _, tx := range latest_block.Transactions {
		//遍历输入
		for _, input := range tx.TXInputs {
			inputs = append(inputs, input)
		}
	}
	//查找需要存入utxo set中的superOutput
	for _, tx := range latest_block.Transactions {
		for index, output := range tx.TXOutputs {
			// 根据output求出address
			superOutput := SuperOutPut{tx.TXID, int64(index), output}
			addUtxoSlice = append(addUtxoSlice, superOutput)
		}
	}

	// 更新
	utxoSet.BlockChain.db.Update(func(tx *bolt.Tx) error {
		fmt.Println("utxoSet.BlockChain.db.Update")
		bucket := tx.Bucket([]byte(utxoBucket))
		if nil != bucket {
			// 删除已花费输出
			for _, input := range inputs {
				superUtxoBytes := bucket.Get([]byte(Adress(input.Ripemd160Hash)))
				if superUtxoBytes != nil {
					superUtxos := DeSerializeTXOutPuts(superUtxoBytes)
					for index, superOutput := range superUtxos.UTXOs {
						if bytes.Compare(input.TXID, superOutput.TXID) == 0 && input.Vout == superOutput.Vout {
							//把这笔superOutput从utxoset中删除
							superUtxos.UTXOs = append(superUtxos.UTXOs[:index], superUtxos.UTXOs[index+1:]...)

							bucket.Put([]byte(Adress(input.Ripemd160Hash)), superUtxos.Serialize())
						}

					}
				}

			}

			//添加新生成的superoutput
			for _, superOutput := range addUtxoSlice {
				superUtxoBytes := bucket.Get([]byte(Adress(superOutput.Output.Ripemd160Hash)))
				if superUtxoBytes != nil {
					superUtxos := DeSerializeTXOutPuts(superUtxoBytes)
					superUtxos.UTXOs = append(superUtxos.UTXOs, superOutput)
					bucket.Put([]byte(Adress(superOutput.Output.Ripemd160Hash)), superUtxos.Serialize())
				} else {
					var superUtxos SuperUTXOs
					superUtxos.UTXOs = append(superUtxos.UTXOs, superOutput)
					bytes := superUtxos.Serialize()
					bucket.Put([]byte(Adress(superOutput.Output.Ripemd160Hash)), bytes)
				}
			}

		} else {
			fmt.Println("tx.Bucket([]byte(utxoBucket))")
			os.Exit(1)
		}
		return nil
	})
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
