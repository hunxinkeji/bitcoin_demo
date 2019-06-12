package main

type BlockChain struct {
	Blocks []*Block
}

func NewBlockChain() *BlockChain {
	block := NewGenesisBlock()
	blockChain := &BlockChain{}
	blockChain.Blocks = append(blockChain.Blocks, block)
	return blockChain
}

func (this *BlockChain) AddBlock(data string) {
	preBlockHash := this.Blocks[len(this.Blocks)-1].Hash
	block := NewBlock(data, preBlockHash)
	this.Blocks = append(this.Blocks, block)
}
