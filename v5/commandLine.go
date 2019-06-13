package main

import (
	"flag"
	"fmt"
	"os"
)

const usage = `
	createWallet				"create wallet"
	getAddressLists				"GetAddressLists"
	createChain --address ADDRESS "create a blockchain"
	send --from FROM --to TO --amount AMOUNT "send coin from FROM to TO"
	getBalance --address ADDRESS "get balance of the address"
	printChain				"print all blocks"
`

const CreateChainCmdString = "createChain"
const GetBalanceCmdString = "getBalance"
const PrintChainCmdString = "printChain"
const CreateWalletString = "createWallet"
const GetAddressListsString = "getAddressLists"
const SendCmdString = "send"
const geniusInfo = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

type CLI struct {
	// bc *BlockChain
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

func (this *CLI) PrintBalance(address string) {
	fmt.Printf(address+" total balance =%f\n", this.GetBalance(address))
}

func (cli *CLI) Run() {
	cli.parameterCheck()

	createChainCmd := flag.NewFlagSet(CreateChainCmdString, flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet(GetBalanceCmdString, flag.ExitOnError)
	printChainCmd := flag.NewFlagSet(PrintChainCmdString, flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet(CreateWalletString, flag.ExitOnError)
	getAddressListsCmd := flag.NewFlagSet(GetAddressListsString, flag.ExitOnError)
	sendCmd := flag.NewFlagSet(SendCmdString, flag.ExitOnError)

	//func (f *FlagSet) String(name string, value string, usage string) *string
	//在内存中创建了一个存放string的空间，并把该空间的地址传出来
	getBalanceCmdPara := getBalanceCmd.String("address", "", "address info")
	createChainCmdPara := createChainCmd.String("address", "", "address info")

	//send相关参数
	sendCmdFrom := sendCmd.String("from", "", "send address info")
	sendCmdTo := sendCmd.String("to", "", "to address info")
	sendCmdAmount := sendCmd.Float64("amount", 0, "amount info")

	switch os.Args[1] {
	case SendCmdString:
		//添加动作
		err := sendCmd.Parse(os.Args[2:])
		IfError("sendCmd.Parse(os.Args[2:])", err)
		if sendCmd.Parsed() {
			if *sendCmdFrom == "" || *sendCmdTo == "" || *sendCmdAmount == 0 {
				cli.printUsage()
			}
			cli.Send(*sendCmdFrom, *sendCmdTo, *sendCmdAmount)
		}
	case GetBalanceCmdString:
		//添加动作
		err := getBalanceCmd.Parse(os.Args[2:])
		IfError("getBalanceCmd.Parse(os.Args[2:])", err)
		if getBalanceCmd.Parsed() {
			if *getBalanceCmdPara == "" {
				cli.printUsage()
			}
			cli.PrintBalance(*getBalanceCmdPara)
		}
	case CreateChainCmdString:
		//创建区块链
		err := createChainCmd.Parse(os.Args[2:])
		IfError("createChainCmd.Parse(os.Args[2:])", err)
		if createChainCmd.Parsed() {
			if *createChainCmdPara == "" {
				cli.printUsage()
			}
			cli.CreateChain(*createChainCmdPara, geniusInfo)
		}
	case PrintChainCmdString:
		//打印
		err := printChainCmd.Parse(os.Args[2:])
		IfError("printChainCmd.Parse(os.Args[2:])", err)
		if printChainCmd.Parsed() {
			cli.PrintChain()
		}
	case CreateWalletString:
		//打印
		err := createWalletCmd.Parse(os.Args[2:])
		IfError("printChainCmd.Parse(os.Args[2:])", err)
		if createWalletCmd.Parsed() {
			cli.CreateWallets()
		} else {
			fmt.Println("createWalletCmd.Parsed() fail")
		}
	case GetAddressListsString:
		//打印
		err := getAddressListsCmd.Parse(os.Args[2:])
		IfError("getAddressListsCmd.Parse(os.Args[2:])", err)
		if getAddressListsCmd.Parsed() {
			cli.GetAddressLists()
		} else {
			fmt.Println("getAddressListsCmd.Parsed() fail")
		}
	default:
		cli.printUsage()
	}
}
