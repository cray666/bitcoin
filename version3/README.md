
到目前为止，我们已经构建了一个有工作量证明机制的区块链。有了工作量证明，挖矿也就有了着落。虽然目前距离一个有着完整功能的区块链越来越近了，但是它仍然缺少了一些重要的特性。我们会将区块链持久化到一个数据库中，然后会提供一个简单的命令行接口，用来完成一些与区块链的交互操作。

本质上，区块链是一个分布式数据库，不过，我们暂时先忽略 “分布式” 这个部分，仅专注于 “存储” 这一点。

## 数据库存储

目前，我们的区块链实现里面并没有用到数据库，而是在每次运行程序时，简单地将区块链存储在内存中。那么一旦程序退出，所有的内容就都消失了。我们没有办法再次使用这条链，也没有办法与其他人共享，所以我们需要把它存储到磁盘上。


实际上，任何一个数据库都可以。在比特币原始论文 中，并没有提到要使用哪一个具体的数据库，它完全取决于开发者如何选择。 Bitcoin Core ，最初由中本聪发布，现在是比特币的一个参考实现，它使用的是  LevelDB。而我们将要使用的是BoltDB

## BoltDB
```shell
非常简洁
用 Go 实现
不需要运行一个服务器
能够允许我们构造想要的数据结构
```
Bolt 是一个纯键值存储的 Go 数据库，启发自 Howard Chu 的 LMDB. 它旨在为那些无须一个像 Postgres 和 MySQL 这样有着完整数据库服务器的项目，提供一个简单，快速和可靠的数据库。

由于 Bolt 意在用于提供一些底层功能，简洁便成为其关键所在。它的 API 并不多，并且仅关注值的获取和设置。仅此而已。

Bolt 使用键值存储，这意味着它没有像 SQL RDBMS （MySQL，PostgreSQL 等等）的表，没有行和列。键值对被存储在 bucket 中,bucket就像我们的数据表。

需要注意的是，Bolt 数据库没有数据类型：键和值都是字节数组（byte array），也就是说我们需要将存储的数据序列化存储，用的时候需要反序列化。我们将会使用  encoding/gob  来完成这一目标，但实际上也可以选择使用 JSON, XML, Protocol Buffers 等等。之所以选择使用 encoding/gob, 是因为它很简单，而且是 Go 标准库的一部分


## 数据库结构

Bitcoin Core 使用两个 “bucket” 来存储数据
```
其中一个 bucket 是 blocks，它存储了描述一条链中所有块的元数据
另一个 bucket 是 chainstate，存储了一条链的状态，也就是当前所有的未花费的交易输出，和一些元数据
```
出于性能的考虑，Bitcoin Core 将每个区块（block）存储为磁盘上的不同文件。如此一来，就不需要仅仅为了读取一个单一的块而将所有（或者部分）的块都加载到内存中。但是，为了简单起见，我们并不会实现这一点。

我们会将整个数据库存储为单个文件，而不是将区块存储在不同的文件中。所以，我们也不会需要文件编号（file number）相关的东西。最终，我们会用到两个键值对.
```shell
# key存储区块的hash,value存储整个区块序列化后的数据
32 字节的 block-hash -> block 结构
# key 为字母L,value存储挖出来的最新的区块的hash
L  -> 链中最后一个块的 hash
```

## 区块数据序列化和反序列化

当我们挖出一个区块或者从网络中更新区块后，需要序列化到本地。使用 encoding/gob 来对这些结构进行序列化：
```go
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)

	return result.Bytes()
}
```
数据反序列化，根据传入的数据库的value值反序列化成区块
```go
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)

	return &block
}
```

## 数据库持久化

在数据库没有创建时候，我们的数据为空，每次运行程序的首时候我们会先检查是否本地已经有区块链
```
1.打开本地的数据库文件
2.检查是否已经存储一个区块链
```
如果已经有区块链
```
3.创建一个新的 Blockchain 实例
4.设置 Blockchain 实例的 tip 为数据库中存储的最后一个块的哈希
```
如果没有区块链：
```
创建创世块
将创世区块存储到数据库
tip 指向创世块的hash（tip 有尾部，尖端的意思，tip存储的是最后一个块的哈希）
```

更新blockchain结构
```go
const dbFile = "blockchain.db" //文件名字
const blocksBucket = "blocks"  //存储数据的桶名字(类似数据表)

type Blockchain struct {
	tip []byte      //存储最新区块的hash
	db  *bolt.DB	//数据文件
}
```
下面我们来更改` NewBlockchain() `函数
```go
func NewBlockchain() *Blockchain {
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			genesis := NewGenesisBlock()
			b, err := tx.CreateBucket([]byte(blocksBucket))
			err = b.Put(genesis.Hash, genesis.Serialize())
			err = b.Put([]byte("l"), genesis.Hash)
			tip = genesis.Hash
		} else {
			tip = b.Get([]byte("l"))
		}

		return nil
	})

	bc := Blockchain{tip, db}

	return &bc
}
```
我们来解析代码
```
db, err := bolt.Open(dbFile, 0600, nil)
```
这是打开一个 BoltDB 文件的标准做法。注意，即使不存在这样的文件，它也不会返回错误。

在 BoltDB 中，数据库操作通过一个事务（transaction）进行操作。有两种类型的事务：只读（read-only）和读写（read-write），update是一个只读的事务
```go
 else {
	tip = b.Get([]byte("l"))
}
```
上述就是打开数据库文件，可以认为Bucket就是存储数据的一个表

不存在区块链，就创建我们区块链
```
b := tx.Bucket([]byte(blocksBucket))
if b == nil {
	genesis := NewGenesisBlock()
	b, err := tx.CreateBucket([]byte(blocksBucket))
	err = b.Put(genesis.Hash, genesis.Serialize())
	err = b.Put([]byte("l"), genesis.Hash)
	tip = genesis.Hash
}
```
存在区块链,我们只需要获取已经存在的区块链的最后一个区块的hash,也就是l,，在我们添加区块时只需用到这个最后一个区块的hash,也就是tip
```
tip = b.Get([]byte("l"))
```

## 迭代器

现在，产生的所有块都会被保存到一个数据库文件里面，但是在实现这一点后，我们失去了之前一个非常好的特性：再也无法打印区块链的区块了，因为现在不是将区块存储在一个数组，而是放到了数据库里面。让我们来解决这个问题！

BoltDB 所有的 key 都以字节序进行存储，此外，因为我们不想将所有的块都加载到内存中（因为我们的区块链数据库可能很大！或者现在可以假装它可能很大），我们将会一个一个地读取它们。我们只需要循环获取每一个区块hash，就可以从数据库中读取value值，反序列化得到区块，更新当前区块的标志变量tip。因此我们需要一个区块链迭代器（Iterator）
```go
type Iterator struct {
	currentHash []byte
	db          *bolt.DB
}
```
创建一个迭代器，存储当前迭代的块哈希（currentHash）和数据库的连接（db）。通过 db，构建我们的迭代器：
```
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}
	return bci
}
```
Iterator的目的就是我们能从这个结构中依次获取我们之前的每一个区块
```go
func (i *BlockchainIterator) Next() *Block {
	var block *Block
	//只读事务
	err := i.db.View(func(tx *bolt.Tx) error {
		//打开数据桶
		b := tx.Bucket([]byte(blocksBucket))
		//通过key获取数据表中存储的数据value
		encodedBlock := b.Get(i.currentHash)
		//反序列化字节数value得到区块
		block = DeserializeBlock(encodedBlock)
		return nil
	})
	handleError(err)
	//更新我们当前区块hash的标志变量
	i.currentHash = block.PrevBlockHash

	return block
}
```

## 测试我们的数据库

我们先创建我们的区块链，并发送两笔交易挖矿，注意一定要将难度值调低，最好调为16或者12。修改我们的main函数
```
func main() {
	bc := NewBlockchain()

	bc.AddBlock("Send 1 BTC to Alice")
	bc.AddBlock("Send 2 more BTC to Bob")
}
```
如果一切正常，你将看到如下信息。
```shell
Mining the block containing "Genesis Block"
0000824c1199605e637641137ff1f75e48bdb82b7731e7b8ad9f43f7b4790ef3

Mining the block containing "Send 1 BTC to Alice"
000003fd136c6991e0213abf95ba77c52c1014267214007d29aab8ee17c7cdad

Mining the block containing "Send 2 more BTC to Bob"
0000e886077ecae698db0eb97fea1f9fd118babac17910cbf790ae197e87603d

```
校验数据库，这时我们需要从数据文件中读取并遍历迭代我们的区块,修改我们的main函数。
```go

func main() {
	bc := NewBlockchain()

	//bc.AddBlock("Send 1 BTC to Alice")
	//bc.AddBlock("Send 2 more BTC to Bob")

	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

```
重新执行,正常的话,将看到下面的信息，从区块的hash我们也可以看出来，前面至少有4个0，这和我们定义的难度值有关
```shell
Prev. hash: 000003fd136c6991e0213abf95ba77c52c1014267214007d29aab8ee17c7cdad
Data: Send 2 more BTC to Bob
Hash: 0000e886077ecae698db0eb97fea1f9fd118babac17910cbf790ae197e87603d
PoW: true

Prev. hash: 0000824c1199605e637641137ff1f75e48bdb82b7731e7b8ad9f43f7b4790ef3
Data: Send 1 BTC to Alice
Hash: 000003fd136c6991e0213abf95ba77c52c1014267214007d29aab8ee17c7cdad
PoW: true

Prev. hash:
Data: Genesis Block
Hash: 0000824c1199605e637641137ff1f75e48bdb82b7731e7b8ad9f43f7b4790ef3
PoW: true
```

- [基本原型](https://github.com/cray666/bitcoin/tree/master/version1)
- [工作量证明](https://github.com/cray666/bitcoin/tree/master/version2)
- [数据库存储](https://github.com/cray666/bitcoin/tree/master/version3)
- [CLI](https://github.com/cray666/bitcoin/tree/master/version4)
- [交易一](https://github.com/cray666/bitcoin/tree/master/version5)
- [交易二](https://github.com/cray666/bitcoin/tree/master/version6)
- [UTXO集](https://github.com/cray666/bitcoin/tree/master/version7)
- [Merkle树](https://github.com/cray666/bitcoin/tree/master/version8)
- [网络](https://github.com/cray666/bitcoin/tree/master/version9)









