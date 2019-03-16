本系列文章将会构建一个基于简单区块链实现的简单加密货币




## 基本介绍

比特币是基于区块链技术的一种 数字货币实现，比特币网络是历史上首个经过大规模、 长时间检验的数字货币系统。 比特币网络是一个分布式的点对点网络，网络中的矿工通过“挖矿”来完成对交易记录的记账过程，维护网络的正常运行。

区块链网络提供一个公共可见的记账本，该记账本并非记录每个账户的余额，而是用来记录发生过的交易的历史信息。 

我们本片文章的逻辑是围绕一个简单的模型，
```
创建区块链->添加一个区块添加到区块链中->再次添加一个区块到区块链中->打印输出所有的区块验证区块链刘晨
```
我们暂时只需要熟悉这个逻辑，每个过程到后面都会变得越来越复杂，比如产生一个区块没这么简单等等
## block 

区块链的本质是一个分布式数据库，每个节点都保存这自己的一个账本数据库，这些数据本质上是以为文件形式存在于硬盘中，具体的文件是由一个区块和一个区块连接形成，
每个区块都存储着上一个区块的hash。

现在我们先定义一个简单的区块的结构，暂时我们会对数据进行简化，后面我们会逐渐深入，但是原理是一致的

```go
type Block struct {
    Timestamp     int64    
    Data          []byte
    PrevBlockHash []byte
    Hash          []byte
}
```
字段解释
```
Timestamp     当前时间戳
Data          整个打包的交易的数据的一个摘要
PrevBlockHash 前一个区块hash
Hash          本区块的hash  
```

本区块的一个hash就是当准备好了其他三个数据之后，通过下面函函数

```go
func (b *Block) SetHash() {
    timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
    headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
    hash := sha256.Sum256(headers)

    b.Hash = hash[:]
}
```
## 生成区块

下面我们需要定义生成新的区块的函数，准备好了交易数据我们就可以生成区块了，因为`前一个区块的hash`,`时间戳`都是已知的。
```
func NewBlock(data string, prevBlockHash []byte) *Block {
    block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}}
    block.SetHash()
    return block
}
```

我们需要定义好第一个区块，也就是我们的创世区块，
```go
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}
```

## 区块链

区块链的本质是保存着特定的数据结构的数据库。它是一个有序的，尾部相连的链表。这就意味着区块是以插入的顺序被存储的，每个区块连接着前一个区块。

为了便于理解，在本章的原型之中我们会首先使用数组创建比较简单的区块链。先理解区块链的大概含义，后续会逐渐变得复杂。


区块链就是把一个一个区块进行连接就是区块,我们使用数组来表示这种关系

```go
type Blockchain struct {
	blocks []*Block
}
```

通过创世区块初始化我们的区块链
```go
func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Block{NewGenesisBlock()}}
}
```

当我们挖出一个新的区块后需要将区块添加到区块链中
```go
func (bc *Blockchain) AddBlock(data string) {
	prevBlock := bc.blocks[len(bc.blocks)-1]
	newBlock := NewBlock(data, prevBlock.Hash)
	bc.blocks = append(bc.blocks, newBlock)
}
```



## 确认区块链模型

```go
func main() {
	bc := NewBlockchain()

	bc.AddBlock("Send 1 BTC to Alice")
	bc.AddBlock("Send 2 more BTC to Bob")

	for _, block := range bc.blocks {
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Println()
	}
}
```

创建区块链->添加一个区块-再次添加一个区块->打印输出所有的区块，

发现打印出来的结果是
```shell
Data: Genesis Block
Hash: 536dab9799d104aa36ebe4d2950af44f989287e2893abad4a3f65879d16c038f

Prev. hash: 536dab9799d104aa36ebe4d2950af44f989287e2893abad4a3f65879d16c038f
Data: Send 1 BTC to Alice
Hash: 5cb2ba632596ad03547cb878a31df50fd80c354e4ca5f7323205cada0f18f27b

Prev. hash: 5cb2ba632596ad03547cb878a31df50fd80c354e4ca5f7323205cada0f18f27b
Data: Send 2 more BTC to Bob
Hash: 259d364c5df874fdbd1c0d3bfca1ada2f7584604e30702baee8a1cebd8305a7f
```

## 总结

我们创建了一个超级简单的区块链原型：它只是一个包含有区块的数组，每个区块和前一个连接起来。

真实的区块链远比这个要复杂。在我们的区块链中添加一个区块很快，也很简单，不过在真实的区块链中，添加一个新的区块需要费一番功夫：一是需要在添加区块前做一些复杂的计算来获取添加的权限(这个过程被称为工作量证明)。

而且，区块链并非是一个单节点决策的东西，它是一个分布式数据库。所以，一个新的区块必须被网络上的其他参与者确认和接受(这个过程被称作一致性)。