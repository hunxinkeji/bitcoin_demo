package main

import "fmt"

func (cli *CLI) AddBlock(data string) {
	cli.bc.AddBlock(data)
}

func (cli *CLI) PrintChain() {
	// cli.bc
	// 打印数据
	it := cli.bc.NewBlockChainIterator()

	for it.CurrHash != nil {
		block := it.Next()
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

	fmt.Println("print blockchain over")
}
