package main

import (
	"bytes"
	"crypto/ecdsa"
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

func isDBExist() bool {
	_, err := os.Stat(dbFile) //(FileInfo, error)
	if err != nil {
		return false
	}
	return true
}

func InitBlockChain(address string, data string) *BlockChain {
	if isDBExist() {
		fmt.Println("Blockchain already exists!")
		os.Exit(1)
	}
	//Open(path string, mode os.FileMode, options *Options) (*DB, error)
	db, err := bolt.Open(dbFile, 0600, nil)
	IfError("bolt.Open(dbFile, 0600, nil)", err)

	var lastHash []byte

	db.Update(func(tx *bolt.Tx) error {
		//没有bucket,
		coinbase := NewCoinbaseTx(address, data)
		genesis := NewGenesisBlock(coinbase)
		bucket, err := tx.CreateBucket([]byte(blockChainBucket)) //(*Bucket, error)
		IfError("tx.CreateBucket([]byte(blockChainBucket))", err)
		key := genesis.Hash
		value := genesis.Serialize()
		err = bucket.Put(key, value)
		IfError("bucket.Put(key, value)", err)
		err = bucket.Put([]byte(lastHashKey), genesis.Hash)
		IfError("bucket.Put([]byte(lastHashKey), genesis.Hash)", err)
		lastHash = genesis.Hash

		return nil
	})

	return &BlockChain{
		db:   db,
		tail: lastHash,
	}
}

func GetBlockChainHandler() *BlockChain {
	if !isDBExist() {
		fmt.Println("Pls create blockchain first!")
		os.Exit(1)
	}

	//Open(path string, mode os.FileMode, options *Options) (*DB, error)
	db, err := bolt.Open(dbFile, 0600, nil)
	IfError("bolt.Open(dbFile, 0600, nil)", err)

	var lastHash []byte

	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockChainBucket))
		if bucket != nil {
			//取出最后区块的哈希值返回
			lastHash = bucket.Get([]byte(lastHashKey))
		} else {
			fmt.Println("bucket does not exist!")
			os.Exit(1)
		}
		return nil
	})

	return &BlockChain{
		db:   db,
		tail: lastHash,
	}
}

func (this *BlockChain) AddBlock(txs []*Transaction) {
	var lastHash []byte

	err := this.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockChainBucket))
		if bucket != nil {
			//取出最后区块的哈希值返回
			lastHash = bucket.Get([]byte(lastHashKey))

			//生成区块之前需要验证交易签名
			for _, tx := range txs {
				// 验证每一笔交易
				if !this.VerifyTransaction(tx) {
					fmt.Println("!this.VerifyTransaction(tx)")
					os.Exit(1)
				} else {
					fmt.Printf("该交易验证签名成功！！！！\n")
				}
			}

			newBlock := NewBlock(txs, lastHash)

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

//返回指定地址能够支配的utxo的交易集合
func (this *BlockChain) FindUTXOTransactions(address string) (utxotxs []*Transaction) {
	//存储使用过的utxo的集合 map[交易id][]int64
	//用来记录已经消费掉的output
	spentUTXO := make(map[string][]int64)

	it := this.NewBlockChainIterator()
	for it.CurrHash != nil {
		block := it.Next()
		for _, tx := range block.Transactions {
			if !tx.IsCoinBase() {
				for _, input := range tx.TXInputs {
					publicKeyHash := Base58Decode([]byte(address))
					ripemd160Hash := publicKeyHash[1 : len(publicKeyHash)-addressChecksumLen]
					if input.UnLockRipemd160Hash(ripemd160Hash) {
						//其实spetUTXO中保存的是被花费掉的以前的ouput的交易id是什么，该交易id中的第几个output
						//因为同一个交易id，可以对应多个output
						spentUTXO[string(input.TXID)] = append(spentUTXO[string(input.TXID)], input.Vout)
					}
				}
			}

		OUTPUTS:
			for currIndex, output := range tx.TXOutputs {
				//该交易id如果在spentUTXO中，就要判断该output是不是被后面的块消费掉了
				if spentUTXO[string(tx.TXID)] != nil { //这句话说明当前output所对应的交易id，曾经被消费过
					//下面就要判断当前的output所对应的输出顺序编号是否相等了，若相等说明该output被后来的区块消费掉了
					indexes := spentUTXO[string(tx.TXID)]
					for _, index := range indexes {
						if int64(currIndex) == index {
							continue OUTPUTS
						}
					}
				}
				if output.CanBeUnlockWith(address) {
					utxotxs = append(utxotxs, tx)
				}
			}
		}
	}

	return
}

func (this *BlockChain) FindUTXO(address string) (utxos []TXOutput) {
	//存储使用过的utxo的集合 map[交易id][]int64
	//用来记录已经消费掉的output
	spentUTXO := make(map[string][]int64)

	it := this.NewBlockChainIterator()
	for it.CurrHash != nil {
		block := it.Next()
		for _, tx := range block.Transactions {
			if !tx.IsCoinBase() {
				//判断每一笔交易输入中，有没有花费属于我的地址的以前的utxo，如果有把它保存下来
				for _, input := range tx.TXInputs {
					publicKeyHash := Base58Decode([]byte(address))
					ripemd160Hash := publicKeyHash[1 : len(publicKeyHash)-addressChecksumLen]
					if input.UnLockRipemd160Hash(ripemd160Hash) {
						//其实spetUTXO中保存的是被花费掉的以前的ouput的交易id是什么，该交易id中的第几个output
						//因为同一个交易id，可以对应多个output
						spentUTXO[string(input.TXID)] = append(spentUTXO[string(input.TXID)], input.Vout)
					}
				}
			}

		OUTPUTS:
			//判断属于我的地址的每一笔交易输出，有没有被之后的交易输入给消费掉，如果没有，保存下来，就是最后要返回的utxos
			for currIndex, output := range tx.TXOutputs {
				if !output.CanBeUnlockWith(address) {
					continue
				}
				//该交易id如果在spentUTXO中，就要判断该output是不是被后面的块消费掉了
				if spentUTXO[string(tx.TXID)] != nil { //这句话说明当前output所对应的交易id，是属于我的地址的，且曾经被消费过
					//下面就要判断当前的output所对应的输出顺序编号是否相等了，若相等说明该output被后来的区块消费掉了
					indexes := spentUTXO[string(tx.TXID)]
					for _, index := range indexes {
						if int64(currIndex) == index {
							continue OUTPUTS
						}
					}
				}
				utxos = append(utxos, output)
			}
		}
	}

	return
}

func (this *BlockChain) FindSuitableUTXOs(address string, amount float64) (validUTXOs map[string][]int64, total float64) {
	validUTXOs = make(map[string][]int64)

	//存储使用过的utxo的集合 map[交易id][]int64
	//用来记录已经消费掉的output
	spentUTXO := make(map[string][]int64)
	it := this.NewBlockChainIterator()
	for it.CurrHash != nil {
		block := it.Next()
		for _, tx := range block.Transactions {
			if !tx.IsCoinBase() {
				for _, input := range tx.TXInputs {
					publicKeyHash := Base58Decode([]byte(address))
					ripemd160Hash := publicKeyHash[1 : len(publicKeyHash)-addressChecksumLen]
					if input.UnLockRipemd160Hash(ripemd160Hash) {
						//其实spetUTXO中保存的是被花费掉的以前的ouput的交易id是什么，该交易id中的第几个output
						//因为同一个交易id，可以对应多个output
						spentUTXO[string(input.TXID)] = append(spentUTXO[string(input.TXID)], input.Vout)
					}
				}
			}

		OUTPUTS:
			for currIndex, output := range tx.TXOutputs {
				//该交易id如果在spentUTXO中，就要判断该output是不是被后面的块消费掉了
				if spentUTXO[string(tx.TXID)] != nil { //这句话说明当前output所对应的交易id，曾经被消费过
					//下面就要判断当前的output所对应的输出顺序编号是否相等了，若相等说明该output被后来的区块消费掉了
					indexes := spentUTXO[string(tx.TXID)]
					for _, index := range indexes {
						if int64(currIndex) == index {
							continue OUTPUTS
						}
					}
				}

				if output.CanBeUnlockWith(address) {
					if total < amount {
						//validUTXOs map[string][]int64
						validUTXOs[string(tx.TXID)] = append(validUTXOs[string(tx.TXID)], int64(currIndex))
						total += output.Value
					} else {
						return
					}
				}
			}
		}
	}

	return
}

// 查找指定的input所对应的output
func (this *BlockChain) FindOutputByInput(input *TXInput) *TXOutput {
	it := this.NewBlockChainIterator()
	for it.CurrHash != nil {
		block := it.Next()
		for _, tx := range block.Transactions {
			if bytes.Compare(tx.TXID, input.TXID) == 0 {
				for index, output := range tx.TXOutputs {
					if int64(index) == input.Vout {
						return &output
					}
				}
			}
		}
	}
	return nil
}

// 交易签名
func (blockChain *BlockChain) SignTransaction(tx *Transaction, privateKey ecdsa.PrivateKey) {
	// coinbase交易不需要签名
	if tx.IsCoinBase() {
		return
	}
	// 处理input，查找交易tx的input所对应的output
	prevOutputs := make(map[string]*TXOutput)
	for _, input := range tx.TXInputs {
		// 查找所引用的每一个交易
		output := blockChain.FindOutputByInput(&input)
		prevOutputs[string(input.TXID)] = output
	}
	// 实现签名函数
	tx.Sign(privateKey, prevOutputs)

}

// 验证签名
func (blockChain *BlockChain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinBase() {
		return true
	}

	prevOutputs := make(map[string]*TXOutput)
	for _, input := range tx.TXInputs {
		// 查找所引用的每一个交易
		output := blockChain.FindOutputByInput(&input)
		prevOutputs[string(input.TXID)] = output
	}

	//实现签名验证
	return tx.Verify(prevOutputs)
}
