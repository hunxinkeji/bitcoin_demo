package main

import (
	"fmt"
)

func (cli *CLI) CreateChain(address string, data string, nodeID string) {
	bc := InitBlockChain(address, data, nodeID)
	defer bc.db.Close()
	fmt.Println("Create Chain Successfully!")

	// 设置utxoSet操作
	utxoSet := &UTXOSet{bc}
	utxoSet.ResetUTXOSet() // 更新UTXO bucket
}

func (this *CLI) GetBalance(address string, nodeID string) (balance float64) {
	bc := GetBlockChainHandler(nodeID)
	defer bc.db.Close()
	utxoSet := &UTXOSet{bc}
	balance = utxoSet.GetBalance(address)
	//utxos := bc.FindUTXO(address)
	//for _, utxo := range utxos {
	//	balance += utxo.Value
	//}
	return
}

func (cli *CLI) PrintChain(nodeID string) {
	bc := GetBlockChainHandler(nodeID)
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

func (this *CLI) Send(from, to string, amount float64, nodeID string) {
	bc := GetBlockChainHandler(nodeID)
	defer bc.db.Close()
	var txs []*Transaction
	tx := NewCoinbaseTx(from, "")
	txs = append(txs, tx)
	tx = NewTransactionFromUTXOSet(from, to, amount, bc, nodeID)
	//tx = NewTransaction(from, to, amount, bc)
	txs = append(txs, tx)

	bc.AddBlock(txs)
	fmt.Println("send successfully")

	// 设置utxoSet操作
	utxoSet := &UTXOSet{bc}
	//utxoSet.ResetUTXOSet() // 更新UTXO bucket
	utxoSet.update()
}

// 创建钱包集合 wallets_3000.dat
func (cli *CLI) CreateWallets(nodeID string) {
	wallets, _ := NewWallets(nodeID)
	wallets.CreateWallet(nodeID)
	fmt.Printf("wallets:%v\n", wallets)
}

func (cli *CLI) GetAddressLists(nodeID string) {
	fmt.Println("打印所有钱包地址...")
	wallets, _ := NewWallets(nodeID)
	for address := range wallets.Wallets {
		fmt.Printf("address:[%s]\n", address)
	}
}

func (cli *CLI) StartNode(nodeID string) {
	StartServer(nodeID)
}
