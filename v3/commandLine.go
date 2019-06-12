package main

import (
	"flag"
	"fmt"
	"os"
)

const usage = `
	addBlock --data DATA	"add a block to blockchain"
	printChain				"print all blocks"
`

const AddBlockCmdString = "addBlock"
const PrintChainCmdString = "printChain"

type CLI struct {
	bc *BlockChain
}

func (cli *CLI) printUsage() {
	fmt.Println(usage)
	os.Exit(1)
}

func (cli *CLI) parameterCheck() {
	if len(os.Args) < 2 {
		fmt.Println("invalid input")
		cli.printUsage()
	}
}

func (cli *CLI) Run() {
	cli.parameterCheck()

	addBlockCmd := flag.NewFlagSet(AddBlockCmdString, flag.ExitOnError)
	printChainCmd := flag.NewFlagSet(PrintChainCmdString, flag.ExitOnError)

	//func (f *FlagSet) String(name string, value string, usage string) *string
	//在内存中创建了一个存放string的空间，并把该空间的地址传出来
	addBlockCmdPara := addBlockCmd.String("data", "", "block trasaction info")

	switch os.Args[1] {
	case AddBlockCmdString:
		//添加动作
		err := addBlockCmd.Parse(os.Args[2:])
		IfError("addBlockCmd.Parse(os.Args[2:])", err)
		if addBlockCmd.Parsed() {
			if *addBlockCmdPara == "" {
				cli.printUsage()
			}
			cli.AddBlock(*addBlockCmdPara)
		}
	case PrintChainCmdString:
		//打印
		err := printChainCmd.Parse(os.Args[2:])
		IfError("printChainCmd.Parse(os.Args[2:])", err)
		if printChainCmd.Parsed() {
			cli.PrintChain()
		}
	default:
		cli.printUsage()
	}
}
