package main

import "fmt"

func (cli *CLI) CreateChain(address string, data string) {
	bc := InitBlockChain(address, data)
	defer bc.db.Close()
	fmt.Println("Create Chain Successfully!")
}

func (cli *CLI) AddBlock(txs []*Transaction) {
	// bc := GetBlockChainHandler()
	// defer bc.db.Close()
	// bc.AddBlock(txs)
}

func (this *CLI) GetBalance(address string) (balance float64) {
	bc := GetBlockChainHandler()
	defer bc.db.Close()
	utxos := bc.FindUTXO(address)
	for _, utxo := range utxos {
		balance += utxo.Value
	}
	return
}

func (cli *CLI) PrintChain() {
	bc := GetBlockChainHandler()
	defer bc.db.Close()
	// 打印数据
	it := bc.NewBlockChainIterator()

	for it.CurrHash != nil {
		block := it.Next()
		fmt.Printf("Version:%d\n", block.Version)
		fmt.Printf("PrevBlockHash:%x\n", block.PrevBlockHash)
		fmt.Printf("Hash:%x\n", block.Hash)
		fmt.Printf("MerkelRoot:%x\n", block.MerkelRoot)
		fmt.Printf("TimeStamp:%d\n", block.TimeStamp)
		fmt.Printf("Bits:%d\n", block.Bits)
		fmt.Printf("Nonce:%d\n", block.Nonce)
		// fmt.Printf("Data:%v\n", block.Transactions)

		fmt.Printf("Isvalid:%v\n", NewProofOfWork(block).IsValid())
		fmt.Println()
	}

	fmt.Println("print blockchain over")
}

func (this *CLI) Send(from, to string, amount float64) {
	bc := GetBlockChainHandler()
	defer bc.db.Close()

	tx := NewTransaction(from, to, amount, bc)

	var txs []*Transaction
	txs = append(txs, tx)

	bc.AddBlock(txs)
	fmt.Println("send successfully")
}

// 创建钱包集合
func (cli *CLI) CreateWallets() {
	wallets, _ := NewWallets()
	wallets.CreateWallet()
	fmt.Printf("wallets:%v\n", wallets)
}

func (cli *CLI) GetAddressLists() {
	fmt.Println("打印所有钱包地址...")
	wallets, _ := NewWallets()
	for address := range wallets.Wallets {
		fmt.Printf("address:[%s]\n", address)
	}
}
