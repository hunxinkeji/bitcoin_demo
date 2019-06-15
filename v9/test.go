package main

import "fmt"

func (cli *CLI) TestMethod() {
	bc := GetBlockChainHandler()
	defer bc.db.Close()

	utxoMap := bc.FindUTXOMap()
	for key, value := range utxoMap {
		fmt.Printf("key:[%s]\n", key)
		for _, output := range value.UTXOs {
			fmt.Printf("value:%v Ripemd160hash:[%x]\n", output.Output.Value, output.Output.Ripemd160Hash)
		}
	}
	fmt.Println("-----------------")
}
