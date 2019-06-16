package main

import (
	"fmt"
	"math"
	"math/big"
)

type ProofOfWork struct {
	Block *Block
	//目标值，比这个值小的时候就算是我们找到了哪个nonce值
	Target *big.Int
}

func NewProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(1)
	target = target.Lsh(target, uint(256-block.Bits))
	return &ProofOfWork{
		Block:  block,
		Target: target,
	}
}

func (pow *ProofOfWork) Run() (int64, []byte) {
	var hashInt big.Int
	var nonce int64 = 0
	hash := pow.Block.BlockHash()

	fmt.Println("Begin Mining...")
	fmt.Printf("target hash:    %x\n", pow.Target.Bytes())

	for nonce < math.MaxInt64 {
		if hashInt.SetBytes(hash).Cmp(pow.Target) < 0 {
			fmt.Printf("found  hash:%x\n nonce:%d\n", hash, nonce)
			break
		}
		nonce++
		pow.Block.Nonce = nonce
		hash = pow.Block.BlockHash()
		//hash = pow.Block.MerkelHash()
	}

	return nonce, hash
}

func (pow *ProofOfWork) IsValid() bool {
	var bigInt big.Int
	pow.Block.Hash = nil
	if bigInt.SetBytes(pow.Block.BlockHash()).Cmp(pow.Target) < 0 {
		return true
	} else {
		return false
	}
}
