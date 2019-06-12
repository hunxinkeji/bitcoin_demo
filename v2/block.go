package main

import (
	"crypto/sha256"
	"encoding/json"
	"time"
)

const targetBits = 24

type Block struct {
	//版本
	Version int64
	//前区块的哈希值
	PrevBlockHash []byte
	//当前区块的哈希值，为了简化代码
	Hash []byte //本区块的哈希值，在比特币中是不存储在区块头中的，这里为了简单而设置了这个字段
	//梅克尔根
	MerkelRoot []byte
	//时间戳
	TimeStamp int64
	//难度值
	Bits int64
	//随机值
	Nonce int64

	//交易信息
	Data []byte //比特币中的区块体是和区块头分开的，这里为了简单，就放在一起了
}

func (this *Block) BlockHash() []byte {
	data, err := json.Marshal(this) //([]byte, error)
	IfError("json.Marshal(this)", err)
	hash := sha256.Sum256(data) //[Size]byte
	return hash[:]
}

//自由的函数，不属于任何结构体
func NewBlock(data string, prevBlockHash []byte) *Block {
	var block *Block
	block = &Block{
		Version:       1,
		PrevBlockHash: prevBlockHash,
		MerkelRoot:    []byte{},
		TimeStamp:     time.Now().Unix(),
		Bits:          targetBits,
		Nonce:         0,
		Data:          []byte(data),
	}

	pow := NewProofOfWork(block)
	block.Nonce, block.Hash = pow.Run()
	return block
}

func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{}) //*Block
}
