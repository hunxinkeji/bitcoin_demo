package main

import (
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
	// 解锁脚本
	ScriptSig string
}

//用公钥检测签名是否合法
func (input *TXInput) CanUnlockUTXOWtith(unlockData string) bool {
	return input.ScriptSig == unlockData
}

type TXOutput struct {
	Value        float64
	ScriptPubKey string
}

//用公钥检查地址是否合法
func (output *TXOutput) CanBeUnlockWith(unlockData string) bool {
	return output.ScriptPubKey == unlockData
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
	for txid, outputsIndexes := range validUTXOs {
		for _, index := range outputsIndexes {
			input := TXInput{
				//所引用输出的交易ID
				TXID: []byte(txid),
				//所引用output的索引值
				Vout: index,
				//解锁脚本，指明可以使用某个output的条件
				ScriptSig: from,
			}
			inputs = append(inputs, input)
		}
	}

	//创建outputs
	//给对方支付的output
	output := TXOutput{
		//支付给收款方的金额
		Value: amount,
		//锁定脚本，指定收款方的地址
		ScriptPubKey: to,
	}
	outputs = append(outputs, output)

	//找零钱的output
	if total > amount {
		output = TXOutput{
			//支付给收款方的金额
			Value: total - amount,
			//锁定脚本，指定收款方的地址
			ScriptPubKey: from,
		}
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

	input := TXInput{nil, -1, data}
	tx = &Transaction{
		TXInputs: []TXInput{input},
		TXOutputs: []TXOutput{
			TXOutput{
				Value:        reward,
				ScriptPubKey: address,
			},
		},
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
