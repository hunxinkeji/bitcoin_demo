package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

func IfError(data string, err error) {
	if err != nil {
		fmt.Println(data+"err=", err)
	}
}

// 反转切片
func Reverse(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

// 将结构体序列化为字节数组
func GobEncode(data interface{}) []byte {
	var buff bytes.Buffer
	//buff是输出
	encoder := gob.NewEncoder(&buff)
	//data是输入
	err := encoder.Encode(data)
	IfError("encoder.Encode(data)", err)
	return buff.Bytes()
}

// 将命令转为字节数组
// 指令长度最长为12位
func CommandToBytes(command string) []byte {
	var bytes [12]byte // 命令长度
	for i, c := range command {
		bytes[i] = byte(c) // 转换
	}
	return bytes[:]
}

// 将字节数组转成cmd
func BytesToCommand(bytes []byte) string {
	var command []byte // 接受命令
	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	res := fmt.Sprintf("%s", command)
	fmt.Println("command = ", res)
	return res
}
