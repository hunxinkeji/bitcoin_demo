package main

import (
	"crypto/sha256"
	"fmt"
)

//merkle树的实现
type MerkleTree struct {
	//根结点
	RootNode *MerkleNode
}

//merkle结点
type MerkleNode struct {
	Data  []byte
	Left  *MerkleNode
	Right *MerkleNode
}

//创建结点
func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	node := &MerkleNode{}
	if left == nil && right == nil {
		//叶子结点
		hash := sha256.Sum256(data)
		node.Data = hash[:]
	} else if left != nil && right != nil {
		//非叶子结点，保存左子结点和右子结点哈希合到一起之后的哈希
		mergeData := append(left.Data, right.Data...)
		hash := sha256.Sum256(mergeData)
		node.Data = hash[:]
		node.Left = left
		node.Right = right
	}

	return node
}

//创建merkle树
//当前区块中的所有交易
func NewMerkleTree(txs []*Transaction) *MerkleTree {
	fmt.Println("NewMerkleTree(txs []*Transaction)")
	merkleTree := &MerkleTree{}
	var nodes []MerkleNode // 保存结点
	var datas [][]byte
	for _, tx := range txs {
		data := tx.Serialize()
		datas = append(datas, data)
	}
	// 判断交易数据有多少条，如果是奇数条，把最后一条拷贝一份
	if len(datas)%2 != 0 {
		datas = append(datas, datas[len(datas)-1])
	}

	// 为每个[]data,创建成一个叶子结点
	for _, data := range datas {
		node := NewMerkleNode(nil, nil, data)
		nodes = append(nodes, *node)
	}

	// 创建非叶子结点（上级结点）
	for i := 0; i < len(datas)/2; i++ {
		var newNodes []MerkleNode // 父结点的列表
		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
			newNodes = append(newNodes, *node)
		}

		if len(newNodes) == 1 {
			merkleTree.RootNode = &newNodes[0]
			break
		}

		if len(newNodes)%2 != 0 {
			newNodes = append(newNodes, newNodes[len(newNodes)-1])
		}
		nodes = newNodes
	}

	return merkleTree
}
