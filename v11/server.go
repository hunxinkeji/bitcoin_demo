package main

import (
	"fmt"
	"io/ioutil"
	"net"
)

// 4000作为主结点地址
var knowNodes = []string{"localhost:4000"}

// 服务处理文件
var nodeAddress string // 结点地址
// 启动服务器
func StartServer(nodeID string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID) // 服务结点地址
	// 监听结点
	listener, err := net.Listen(PROTOCOL, nodeAddress)
	IfError("net.Listen(PROTOCOL, nodeAddress)", err)
	defer listener.Close()

	bc := GetBlockChainHandler(nodeID)
	if nodeAddress != knowNodes[0] {
		// 非主结点，向主结点发送请求，同步数据
		// sendVersion
		SendVersion(knowNodes[0], bc)
	}

	for {
		// 主结点接受请求
		conn, err := listener.Accept()
		IfError("listener.Accept()", err)
		go HandleConnection(conn, bc)
	}
}

func HandleConnection(conn net.Conn, bc *BlockChain) {
	request, err := ioutil.ReadAll(conn)
	IfError("conn.Read(request[:])", err)
	cmd := BytesToCommand(request[:CMDLENGTH])
	fmt.Printf("Receive a message:%s\n", cmd)
	switch cmd {
	case CMD_VERSION:
		HandleVersion(request, bc)
	case CMD_GETBLOCKS:
		HandleGetBlocks(request, bc)
	case CMD_INV:
		HandleInv(request, bc)
	case CMD_GETDATA:
		HandleGetData(request, bc)
	case CMD_BLOCK:
		HandleBlock(request, bc)
	default:
		fmt.Println("UnknownCmd")
	}
	conn.Close()
}
