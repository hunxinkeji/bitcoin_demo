### 请求处理
网络结点之间的信息传输就是一个区块同步的过程
主要请求如下：
1. version:验证当前结点的末端区块是否是最新区块
2. getBlocks:从最长的链上面获取区块
3. Inv:向其他结点展示当前结点有哪些区块
4. getData:请求一个指定的区块
5. block:接收到新区块的时候，进行处理
