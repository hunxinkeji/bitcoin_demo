package main

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
)

// 钱包集合的文件
const walletFile = "wallets_%s.dat" // 存储钱包集合的文件

// 钱包的集合机构
type Wallets struct {
	//key:string->地址
	Wallets map[string]*Wallet
}

// 初始化一个钱包的集合
func NewWallets(nodeID string) (*Wallets, error) {
	walletFile := fmt.Sprintf(walletFile, nodeID)
	// 1.判断文件是否存在
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		wallets := &Wallets{}
		wallets.Wallets = make(map[string]*Wallet)
		return wallets, err
	}

	// 2.文件存在，读取内容
	fileContent, err := ioutil.ReadFile(walletFile)
	IfError("ioutil.ReadFile(walletFile)", err)
	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	IfError("decoder.Decode(&wallets)", err)
	return &wallets, nil
}

// 创建新的钱包，并且将其添加到集合中
func (wallets *Wallets) CreateWallet(nodeID string) {
	wallet := NewWallet() // 新钱包对象
	wallets.Wallets[string(wallet.GetAddress())] = wallet
	//把钱包存储到文件中
	wallets.SaveWallets(nodeID)
}

// 持久化钱包信息(写入文件)
func (wallets *Wallets) SaveWallets(nodeID string) {
	var content bytes.Buffer
	//注册
	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(wallets)
	IfError("encoder.Encode(wallets)", err)
	//清空文件再存储
	walletFile := fmt.Sprintf(walletFile, nodeID)
	err = ioutil.WriteFile(walletFile, content.Bytes(), 0600)
	IfError("ioutil.WriteFile()", err)
}
