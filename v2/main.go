package main

import (
	"fmt"
)

func main() {
	blockchain := NewBlockChain()

	blockchain.AddBlock("A send to B 1BTC")
	blockchain.AddBlock("B send to C 1BTC")

	for _, block := range blockchain.Blocks {
		fmt.Printf("Version:%d\n", block.Version)
		fmt.Printf("PrevBlockHash:%x\n", block.PrevBlockHash)
		fmt.Printf("Hash:%x\n", block.Hash)
		fmt.Printf("MerkelRoot:%x\n", block.MerkelRoot)
		fmt.Printf("TimeStamp:%d\n", block.TimeStamp)
		fmt.Printf("Bits:%d\n", block.Bits)
		fmt.Printf("Nonce:%d\n", block.Nonce)
		fmt.Printf("Data:%s\n", string(block.Data))

		fmt.Printf("Isvalid:%v\n", NewProofOfWork(block).IsValid())
		fmt.Println()
	}
}
