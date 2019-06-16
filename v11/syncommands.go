package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"net"
)

type BlockData struct {
	AddrFrom string
	Block    *Block
}

type GetData struct {
	AddrFrom string
	ID       []byte
	Type     string
}

type GetBlocks struct {
	AddrFrom string
}

// 代表当前区块的版本信息（决定是否需要进行同步）
type Version struct {
	Version    int    // 版本
	BestHeight int    // 当前结点区块的高度
	AddrFrom   string // 当前结点的地址
}

// 向其他结点展示当前结点有哪些区块
type Inv struct {
	Hashes   [][]byte
	AddrFrom string
	Type     string // 类型(交易或者区块，当前没有交易信息的展示)
}

// 专门用于处理各种相关请求

//1. version:验证当前结点的末端区块是否是最新区块
func HandleVersion(request []byte, bc *BlockChain) {
	fmt.Println("HandleVersion")
	var buff bytes.Buffer
	var version Version
	// 解析request 命令解析
	// 解析request 获取数据
	dataBytes := request[CMDLENGTH:]
	buff.Write(dataBytes)
	//buff是输入
	decoder := gob.NewDecoder(&buff)
	//data是输出
	err := decoder.Decode(&version)
	IfError("decoder.Decode(&version)", err)

	// 获取区块高度
	bestHeight := bc.GetBestHeight()
	fmt.Println("receive version.BestHeight =", version.BestHeight)
	if bestHeight > version.BestHeight {
		SendVersion(version.AddrFrom, bc)
	} else if bestHeight < version.BestHeight {
		SendGetBlocks(version.AddrFrom)
	}
}

//2. getBlocks:从最长的链上面获取区块
func HandleGetBlocks(request []byte, bc *BlockChain) {
	fmt.Println("HandleGetBlocks")
	var buff bytes.Buffer
	var data GetBlocks
	// 解析request 命令解析
	// 解析request 获取数据
	dataBytes := request[CMDLENGTH:]
	buff.Write(dataBytes)
	//buff是输入
	decoder := gob.NewDecoder(&buff)
	//data是输出
	decoder.Decode(&data)
	hashes := bc.GetBlockHashes()

	SendInv(data.AddrFrom, BLOCK_TYPE, hashes)
}

func HandleInv(request []byte, bc *BlockChain) {
	fmt.Println("HandleInv")
	var buff bytes.Buffer
	var data Inv
	// 解析request 命令解析
	// 解析request 获取数据
	dataBytes := request[CMDLENGTH:]
	buff.Write(dataBytes)
	//buff是输入
	decoder := gob.NewDecoder(&buff)
	//data是输出
	decoder.Decode(&data)
	blockHash := data.Hashes[0]
	SendGetData(data.AddrFrom, BLOCK_TYPE, blockHash)
}

//4. getData:请求一个指定的区块
func HandleGetData(request []byte, bc *BlockChain) {
	fmt.Println("HandleGetData")
	var buff bytes.Buffer
	var data GetData
	// 解析request 命令解析
	// 解析request 获取数据
	dataBytes := request[CMDLENGTH:]
	buff.Write(dataBytes)
	//buff是输入
	decoder := gob.NewDecoder(&buff)
	//data是输出
	decoder.Decode(&data)
	// 获取指定ID的区块信息
	// GetBlock(ID []byte)
	//TODO
	block, err := bc.GetBlock(data.ID)
	if err != nil {
		return
	}
	SendBlock(data.AddrFrom, block)
}

//5. block:接收到新区块的时候，进行处理
func HandleBlock(request []byte, bc *BlockChain) {
	fmt.Println("HandleBlock")
	var buff bytes.Buffer
	var data BlockData
	// 解析request 命令解析
	// 解析request 获取数据
	dataBytes := request[CMDLENGTH:]
	buff.Write(dataBytes)
	//buff是输入
	decoder := gob.NewDecoder(&buff)
	//data是输出
	decoder.Decode(&data)
	block := data.Block
	//调用添加区块到区块链中的函数
	bc.addblock(block)

	utxoSet := UTXOSet{bc}
	utxoSet.update()
}

//"客户端（结点）"向服务器发送请求
func SendMessage(to string, msg []byte) {
	fmt.Println("SendMessage")
	fmt.Println("向服务器发送请求...")
	conn, err := net.Dial(PROTOCOL, to)
	IfError("net.Dial(PROTOCOL, to)", err)
	defer conn.Close()
	// 要发送的数据添加到请求中
	//n, err := conn.Write(msg)
	//fmt.Println("write to server n= ", n)
	_, err = io.Copy(conn, bytes.NewReader(msg))
	IfError("io.Copy", err)
}

// 数据同步的函数rm
func SendVersion(to string, bc *BlockChain) {
	fmt.Println("SendVersion")
	// 在比特币中，消息是底层的比特序列，前12个字节指定命令名（version）
	// 后面的字节包含的是gob编码过的消息结构
	bestHeight := bc.GetBestHeight()
	data := GobEncode(Version{NODE_VERSION, bestHeight, nodeAddress})
	request := append(CommandToBytes(CMD_VERSION), data...)
	SendMessage(to, request)
}

// 向其他结点展示区块信息
func SendInv(to string, kind string, hash [][]byte) {
	fmt.Println("SendInv")
	data := GobEncode(Inv{
		Hashes:   hash,
		AddrFrom: nodeAddress,
		Type:     kind,
	})
	request := append(CommandToBytes(CMD_INV), data...)
	SendMessage(to, request)
}

func SendGetBlocks(to string) {
	fmt.Println("SendGetBlocks")
	data := GobEncode(GetBlocks{
		AddrFrom: nodeAddress,
	})
	request := append(CommandToBytes(CMD_GETBLOCKS), data...)
	SendMessage(to, request)
}

// 向其他人展示交易或者区块信息
func SendGetData(to string, kind string, hash []byte) {
	fmt.Println("SendGetData")
	data := GobEncode(GetData{
		AddrFrom: nodeAddress,
		ID:       hash,
		Type:     kind,
	})
	request := append(CommandToBytes(CMD_GETDATA), data...)
	SendMessage(to, request)
}

func SendBlock(to string, block *Block) {
	fmt.Println("SendBlock")
	data := GobEncode(BlockData{
		AddrFrom: nodeAddress,
		Block:    block,
	})
	request := append(CommandToBytes(CMD_BLOCK), data...)
	SendMessage(to, request)
}
