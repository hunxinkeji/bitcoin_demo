package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
)

const reward = 12.5

type Transaction struct {
	// 交易ID
	TXID []byte
	// 交易输入
	TXInputs []TXInput
	// 交易输出
	TXOutputs []TXOutput
}

type TXInput struct {
	// 所引用输出的交易ID
	TXID []byte
	// 所引用output的索引值
	Vout int64

	// 数字签名
	Signature []byte
	// 公钥
	PublicKey []byte
}

func (input *TXInput) UnLockRipemd160Hash(ripemd160Hash []byte) bool {
	// 获取input的ripemd160哈希
	inputRipemd160 := Ripemd160Hash(input.PublicKey)
	return bytes.Compare(inputRipemd160, ripemd160Hash) == 0
}

type TXOutput struct {
	Value         float64
	Ripemd160Hash []byte
}

// 创建output对象
func NewTXOutput(value float64, address string) TXOutput {
	txOuput := TXOutput{}
	hash160 := Lock(address)
	txOuput.Value = value
	txOuput.Ripemd160Hash = hash160

	return txOuput
}

//用公钥检查地址是否合法
func (output *TXOutput) CanBeUnlockWith(address string) bool {
	hash160 := Lock(address)
	return bytes.Compare(hash160, output.Ripemd160Hash) == 0
}

//相当于锁定
func Lock(address string) []byte {
	publicKeyHash := Base58Decode([]byte(address))
	hash160 := publicKeyHash[1 : len(publicKeyHash)-addressChecksumLen]
	return hash160
}

//创建普通交易，send命令的普通函数
func NewTransaction(from, to string, amount float64, bc *BlockChain) (tx *Transaction) {
	//make(map[string][]int64) key:交易id，value:引用output的索引切片
	validUTXOs := make(map[string][]int64)
	var total float64
	validUTXOs /*所需要的，合理的utxos集合*/, total /*返回选择到的utxos的金额总和*/ = bc.FindSuitableUTXOs(from, amount)

	if total < amount {
		fmt.Println("Not enough money")
		os.Exit(1)
	}

	//进行output到input的转换
	var inputs []TXInput
	var outputs []TXOutput

	// 获取钱包集合
	wallets, _ := NewWallets()
	wallet := wallets.Wallets[from] // 指定地址对应的钱包结构
	for txid, outputsIndexes := range validUTXOs {
		for _, index := range outputsIndexes {
			input := TXInput{
				//所引用输出的交易ID
				TXID: []byte(txid),
				//所引用output的索引值
				Vout: index,

				Signature: nil,
				PublicKey: wallet.PublicKey,
			}
			inputs = append(inputs, input)
		}
	}

	//创建outputs
	//给对方支付的output
	output := NewTXOutput(amount, to)
	outputs = append(outputs, output)

	//找零钱的output
	if total > amount {
		output = NewTXOutput(total-amount, from)
		outputs = append(outputs, output)
	}

	tx = &Transaction{
		TXInputs:  inputs,
		TXOutputs: outputs,
	}
	tx.SetTXID()
	return
}

//创建coinbase交易，只有收款人，没有付款人，是矿工的奖励交易
func NewCoinbaseTx(address string, data string) (tx *Transaction) {
	if data == "" {
		data = fmt.Sprintf("reward to %s %v btc", address, reward)
	}

	input := TXInput{nil, -1, nil, nil}
	output := NewTXOutput(reward, address)
	tx = &Transaction{
		TXInputs:  []TXInput{input},
		TXOutputs: []TXOutput{output},
	}
	tx.SetTXID()
	return
}

func (this *Transaction) IsCoinBase() bool {
	if len(this.TXInputs) == 1 {
		if this.TXInputs[0].TXID == nil {
			return true
		}
	}
	return false
}

//设置交易ID，是一个哈希值
func (this *Transaction) SetTXID() (hash []byte) {
	bytes, err := json.Marshal(this) //([]byte, error)
	IfError("json.Marshal(this)", err)
	bytesArr := sha256.Sum256(bytes) //[Size]byte
	hash = bytesArr[:]
	this.TXID = hash
	return
}
