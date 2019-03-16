
在分布式系统中，共识（ consensus）在很多时候会与一致性（consistency）术语放在一起讨论，两个实际上还是不一样的。

一致性往往指分布式系统中多个副本对外呈现的数据的状态。 如前面提到的顺序一致 性、 线性一致性，描述了多个节点对数据状态的维护能力。
 
共识则描述了分布式系统中多 个节点之间，彼此对某个状态达成一致结果的过程。

因此，一致性描述的是结果状态，共识 则是一种手段。 注意，达成某种共识并不意味着就保障了一致性。


## 工作量证明算法


在版本version1中我们添加生成区块的过程非常简单，只要将交易打包进区块就可以生成，向链中加入区块太容易，也太廉价了，每个人都很容易的将数据生成区块写入账本广播，全网络中的账本根本不肯能达成一致。而现实中也是如此，当你打包一笔交易后，你必须付出一定的努力才能将数据放入到区块链中，正是由于这种困难的工作，才保证了区块链的安全和一致。

工作量证明：简单来说就是将打包的数据和一个随机的数进行拼接，通过一个固定的函数得到一个结果，如果结果小于网络此时的目标值，这个过程就是工作量证明.

工作量证明（ PoW）通过计算来猜测一个数值（ nonce），使得拼凑上交易数据后内容的 Hash 值满足规定的上限（来源于 hashcash）。 由于 Hash 难题在目前计算模型下需要大量的 计算，这就保证在一段时间内系统中只能出现少数合法提案。 反过来，能够提出合法提案， 也证明提案者确实付出了一定的工作量。 同时，这些少量的合法提案会在网络中进行广播，收到的用户进行验证后，会在用户 认为的最长链基础上继续难题的计算。 因此，系统中可能出现链的分叉（ fork），但最终会有 一条链成为最长的链。

听起来是不是很简单，但是实际中却很困难，你要找到一个小于目标值的值太困难了，只能通过暴力计算。而且目标值会随着全网的算力进行调整，难度也会加大,保证平均10分钟左右找到一个这样的hash。

 
## 定义难度

我们会首先定义一个难度值,这个值就决定的挖矿难度。
```
const targetBits = 16
```
目标值和这个数紧密相关，这个数值越大，目标值前面的0就越多，目标值就越小。

## 定义工作量证明的结构

我们先定义一个工作量证明的结构，挖矿的数据就来自于这个结构体

```go
type ProofOfWork struct {
	block  *Block
	target *big.Int  //表示数学难题的目标值
}
```

## 构建`ProofOfWork`的target

当我们将打包的交易数据准备好之后，构建我们的这个
```go
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{b, target}

	return pow
}
```

## 修改block结构

计算难题还需要一个随机数，这个随机数也会添加到我们的block中，修改我们block的结构。

```go
type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}
```
## 修改生成区块的函数

```go
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{},0}
	//准备数据
	pow := NewProofOfWork(block)
	//工作量证明计算过程
	nonce, hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}
```
NewProofOfWork就是我们打包之后准备的交易数据，当我们将数据准备好之后，就通过run函数暴力计算难题。nonce和区块的hash就是我们要计算的数学难题的答案


## 准备数据
```go
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.Data,
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}
```
其实就是打包的交易数据，目标值，nonce进行合并，变成一个数据，，主要是为了后面的计算。看到下面POW算法for循环，你就明白了。


## 工作量证明POW核心
```go
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)

	for nonce < maxNonce {
		data := pow.prepareData(nonce)

		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")

	return nonce, hash[:]
}
```
我们看出主要的就是这个暴力计算的for循环
```
通能过nonce自增准备数据
用 SHA-256 对数据进行哈希
将哈希转换成一个大整数
将这个大整数与目标进行比较
```
x.Cmp(y) 
```
x < y  返回 -1
x == y 返回 0 
x > y  返回 1
```
`fmt.Printf("\r%x", hash)`
```shell
\r  # 表示将光标定位到本行开头，所以我们可以看到挖矿过程中数字在不停变化

```



## 总结

现在区块链离真实的架构更近了一步：现在添加区块需要困难的工作了，这就让挖坑成为了可能。

但仍旧缺少一些至关重要的功能：区块链数据库并未持久化，也没有钱包，地址，交易，也没有一致化机制。

所有的这些我们都要在未来的文章中实现，现在，祝挖矿快乐！


