package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"
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
	// 所引用output的地址
	Ripemd160Hash []byte

	// 数字签名
	Signature []byte
	// 公钥
	PublicKey []byte
}

func (this *Transaction) Serialize() (bytes []byte) {
	//改成Merkle树的根哈希
	bytes, err := json.Marshal(this)
	IfError("json.Marshal(this)", err)
	return
}

func DeSerializeTransaction(data []byte) (transaction *Transaction) {
	transaction = &Transaction{}
	err := json.Unmarshal(data, transaction)
	IfError("哈哈json.Unmarshal(data, block)", err)
	return
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

func Adress(Ripemd160Hash []byte) string {
	ripemd160Hash := Ripemd160Hash
	//根据Ripemd160Hash求出地址
	version_ripemd160Hash := append([]byte{version}, ripemd160Hash...)
	checkSumBytes := CheckSum(version_ripemd160Hash)
	bytes := append(version_ripemd160Hash, checkSumBytes...)
	address := string(Base58Encode(bytes))
	return address
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

func NewTransactionFromUTXOSet(from, to string, amount float64, bc *BlockChain, nodeID string) (tx *Transaction) {
	validUTXOs := make(map[string][]OutputInfo)
	var total float64
	validUTXOs /*所需要的，合理的utxos集合*/, total /*返回选择到的utxos的金额总和*/ = bc.FindSuitableUTXOsFromUTXOSet(from, amount)

	if total < amount {
		fmt.Println("Not enough money")
		os.Exit(1)
	}

	//进行output到input的转换
	var inputs []TXInput
	var outputs []TXOutput

	// 获取钱包集合
	wallets, _ := NewWallets(nodeID)
	wallet := wallets.Wallets[from] // 指定地址对应的钱包结构
	for txid, outputsIndexes := range validUTXOs {
		for _, index := range outputsIndexes {
			input := TXInput{
				//所引用输出的交易ID
				TXID: []byte(txid),
				//所引用output的索引值
				Vout:          index.Vout,
				Ripemd160Hash: index.Ripemd160Hash,
				Signature:     nil,
				PublicKey:     wallet.PublicKey,
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

	// 对交易进行签名
	// signTransaction()
	// 参数主要为：tx, wallet.PrivateKey
	bc.SignTransaction(tx, wallet.PrivateKey)
	return
}

//创建普通交易，send命令的普通函数
func NewTransaction(from, to string, amount float64, bc *BlockChain, nodeID string) (tx *Transaction) {
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
	wallets, _ := NewWallets(nodeID)
	wallet := wallets.Wallets[from] // 指定地址对应的钱包结构
	for txid, outputsIndexes := range validUTXOs {
		for _, index := range outputsIndexes {
			input := TXInput{
				//所引用输出的交易ID
				TXID: []byte(txid),
				//所引用output的索引值
				Vout:      index,
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

	// 对交易进行签名
	// signTransaction()
	// 参数主要为：tx, wallet.PrivateKey
	bc.SignTransaction(tx, wallet.PrivateKey)
	return
}

//创建coinbase交易，只有收款人，没有付款人，是矿工的奖励交易
func NewCoinbaseTx(address string, data string) (tx *Transaction) {
	if data == "" {
		data = fmt.Sprintf("reward to %s %v btc", address, reward)
	}

	input := TXInput{nil, -1, nil, nil, nil}
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
	this.TXID = nil
	bytes, err := json.Marshal(this) //([]byte, error)
	IfError("json.Marshal(this)", err)
	bytesArr := sha256.Sum256(bytes) //[Size]byte
	hash = bytesArr[:]
	this.TXID = hash
	return
}

//交易签名
func (this *Transaction) Sign(privKey ecdsa.PrivateKey, prevOutputs map[string]*TXOutput) {
	if this.IsCoinBase() {
		return
	}

	for _, input := range this.TXInputs {
		if prevOutputs[string(input.TXID)] == nil {
			fmt.Println("prevOutputs[string(input.TXID)] == nil")
			os.Exit(1)
		}
	}

	// 提取需要签名的属性
	// 获取copy tx
	txCopy := this.TrimmedCopy()
	for index, input := range txCopy.TXInputs {
		output := prevOutputs[string(input.TXID)] // 获取input所对应的output
		//只修改了PublicKey只一项，其他的都是txCopy中的内容
		txCopy.TXInputs[index].PublicKey = output.Ripemd160Hash
		//签名的信息是：Transaction{nil, []TXInput中的TXInput的PublicKey做了修改，修改为原来output中的Ripemd160Hash了，其他的Signature以及PublicKey都为nil}
		txCopy.TXID = txCopy.SetTXID()
		//求出签名的哈希后，把设置过的PublicKey重新设置为nil，恢复为原来的txCopy
		txCopy.TXInputs[index].PublicKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.TXID)
		if err != nil {
			IfError("ecdsa.Sign", err)
		}

		//ECDSA的签名算法就是一对数字
		//Sig = (R,S)
		signature := append(r.Bytes(), s.Bytes()...)

		//最终是为了设置每一个input中的Signature
		this.TXInputs[index].Signature = signature
		//恢复为原来的txCopy
		txCopy.TXID = nil
	}

}

//添加一个交易的拷贝，注意是深拷贝，用于交易签名，返回需要签名的交易
func (this *Transaction) TrimmedCopy() *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	// 把原来交易中的input中的Signature和PublicKey设置为nil，这两个元素是公开的
	for _, input := range this.TXInputs {
		inputs = append(inputs, TXInput{input.TXID, input.Vout, input.Ripemd160Hash, nil, nil})
	}

	for _, output := range this.TXOutputs {
		outputs = append(outputs, TXOutput{output.Value, output.Ripemd160Hash})
	}

	txCopy := Transaction{nil, inputs, outputs}
	return &txCopy
}

func (this *Transaction) Verify(prevOutputs map[string]*TXOutput) bool {
	if this.IsCoinBase() {
		return true
	}

	for _, input := range this.TXInputs {
		if prevOutputs[string(input.TXID)] == nil {
			fmt.Println("prevOutputs[string(input.TXID)] == nil")
			os.Exit(1)
		}
	}

	// 提取需要签名的属性
	// 获取copy tx
	txCopy := this.TrimmedCopy()
	for index, input := range txCopy.TXInputs {
		output := prevOutputs[string(input.TXID)] // 获取input所对应的output
		//只修改了PublicKey只一项，其他的都是txCopy中的内容
		txCopy.TXInputs[index].PublicKey = output.Ripemd160Hash
		//签名的信息是：Transaction{nil, []TXInput中的TXInput的PublicKey做了修改，修改为原来output中的Ripemd160Hash了，其他的Signature以及PublicKey都为nil}
		txCopy.TXID = txCopy.SetTXID()
		//求出签名的哈希后，把设置过的PublicKey重新设置为nil，恢复为原来的txCopy
		txCopy.TXInputs[index].PublicKey = nil

		//用公钥解码签名取出的值跟txCopy.TXID相比较
		//获取r,s（r和s长度相等，根据椭圆加密计算的结果）
		//r,s代表签名
		r := big.Int{}
		s := big.Int{}
		sigLen := len(this.TXInputs[index].Signature)
		r.SetBytes(this.TXInputs[index].Signature[:(sigLen / 2)])
		s.SetBytes(this.TXInputs[index].Signature[(sigLen / 2):])

		//生成x和y(首先，签名是一个数字对，公钥是x,y坐标组合，
		// 在生成公钥时，需要将X Y坐标组合到一起, 在验证时需要将公钥的X,Y拆开)
		x := big.Int{}
		y := big.Int{}
		pubKeyLen := len(this.TXInputs[index].PublicKey)
		x.SetBytes(this.TXInputs[index].PublicKey[:(pubKeyLen / 2)])
		y.SetBytes(this.TXInputs[index].PublicKey[(pubKeyLen / 2):])

		// 生成验证签名所需要的公钥
		rawPubKey := ecdsa.PublicKey{elliptic.P256(), &x, &y}

		// 验证签名
		if !ecdsa.Verify(&rawPubKey, txCopy.TXID, &r, &s) {
			return false
		}

		//恢复为原来的txCopy
		txCopy.TXID = nil
	}
	return true
}
