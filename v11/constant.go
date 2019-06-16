package main

const PROTOCOL = "tcp"
const VERSION = "version"
const BLOCK = "block"
const NODE_VERSION = 1

// 命令长度
const CMDLENGTH = 12

/*
1. version:验证当前结点的末端区块是否是最新区块
2. getBlocks:从最长的链上面获取区块
3. Inv:向其他结点展示当前结点有哪些区块
4. getData:请求一个指定的区块
5. block:接收到新区块的时候，进行处理
*/

// 请求命令
const CMD_VERSION = "version"
const CMD_GETBLOCKS = "getBlocks"
const CMD_INV = "Inv"
const CMD_GETDATA = "getData"
const CMD_BLOCK = "block"

const BLOCK_TYPE = "block"
