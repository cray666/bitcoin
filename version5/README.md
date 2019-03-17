

交易（transaction）是比特币的核心所在，而区块链唯一的目的，也正是为了能够安全可靠地存储交易。在区块链中，交易一旦被创建，就没有任何人能够再去修改或是删除它。

不过交易的内容很复杂，但是本篇文章只是实现交易的基本框架，后面我们会继续深入细节

如果你对比特币的交易模型UTXO不是很了解，先去了解一下UTXO模型，整篇文章都是针对UTXO模型来编写的。

## 比特币交易

一笔合法的交易，即引用某些已存在交易的 `UTXO`作为交易的输入，并生成新的输出的过程。引入不同地址上的合法的UTXO作为输入，转给其他的地址上
，这些地址有可能是一个，也可能是多个。所以定义结构
```
type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}
```
我们也能发现，比特币的交易有如下特点
```
一个输入必须引用一个输出
一笔交易的输入可以引用之前多笔交易的输出,这主要取决于你的单笔输入是否大于你要转账的金额。
比特币交易本质就是消费一些地址UTXO，在别的地址上创建一些新的UTXO
```
> 需要注意的是，每笔输出的UTXO金额是不可以再分的，你想引用一笔输出上的UTXO相当于你就引用了所有的金额，而不是你只想引用一部分，
(这里注意是每笔输出，而不是每个地址，一个地址上可能有很币UTXO。)
当你引用的值超过你要转账的金额，那么就会产生一个找零，找零会返还给发送方，当前每笔交易你需要按照交易规则付一笔额外的交易费，


## 交易输出
```go
type TXOutput struct {
	Value        int
	ScriptPubKey string
}
```
简单来说。我想要给一个地址转账，我只需要知道地址是谁和转账金额.
由于在网络中确保安全，这里采用非对称加密的方式给账户给这笔金额上锁,
```go
type TXOutput struct {
	Value        int
	ScriptPubKey string
}
```
`value`就是转账金额，`ScriptPubKey`就是这个锁，一般称为锁定脚本,比特币使用了一个叫做 Script 的脚本语言，用它来定义锁定和解锁输出的逻辑。虽然这个语言相当的原始（这是为了避免潜在的黑客攻击和滥用而有意为之），并不复杂，但是我们也并不会在这里讨论它的细节。你可以在这里 找到详细解释。

由于还没有实现地址（address），所以目前我们会避免涉及逻辑相关的完整脚本。ScriptPubKey 将会存储一个任意的字符串（用户定义的钱包地址）

我们还需要添加一个函数，来证明这笔金额能不能被你提供的脚本解锁，也就是 确认你是否能花费这笔钱
```go
func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}
```

## 交易输入

```
这里是输入：
type TXInput struct {
	Txid      []byte
	Vout      int
	ScriptSig string
}
```
一个输入引用了之前交易的输出
```shell
Txid 存储的是之前交易的 ID，
Vout 存储的是该输出在那笔交易中所有输出的索引（因为一笔交易可能有多个输出，需要有信息指明是具体的哪一个）。
ScriptSig 是一个脚本，
```
ScriptSig 这个脚本提供了可解锁输出结构里面 ScriptPubKey 字段的数据。如果 ScriptSig 提供的数据是正确的，那么输出就会被解锁，然后被解锁的值就可以被用于产生新的输出；如果数据不正确，输出就无法被引用在输入中，或者说，无法使用这个输出。这种机制，保证了用户无法花费属于其他人的币。

由于还没有实现地址（address），所以目前我们会避免涉及逻辑相关的完整脚本。ScriptPubKey 将会存储一个字符串，目前来书就是相当于用户的账户，后面解锁部分会进行响应的改进。

我们需要创建一个函数，来确认这笔交易的输入是否已经花费了,因为一笔输出有没有被花费，只能看这笔交易有没有成为其他的交易的输入
```go
func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}
```

## coinbase交易

每一笔输入都是之前一笔交易的输出，那么假设从某一笔交易开始不断往前追溯，它所涉及的输入和输出到底是谁先存在呢？

每个区块的第一笔交易只有输出，没有输入，当矿工挖出一个新的块时，它会向新的块中添加一个 coinbase 交易。coinbase 交易是一种特殊的交易，它不需要引用之前一笔交易的输出。它"凭空"产生了币（也就是产生了新币），这是矿工获得挖出新块的奖励，也可以理解为“发行新币”。这也是比特币的来源。

每个区块都存在着一个coinbase交易，简单来说这个交易主要用于给矿工提供奖励
```go
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{[]byte{}, -1, data}
	txout := TXOutput{subsidy, to}
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.SetID()

	return &tx
}
```
在交易过程中，转账方需要通过签名脚本来证明自己是 UTXO 的合法使用者，并且指 定输出脚本来限制未来本交易的使用者（为收款方）。 对每笔交易，转账方需要进行签名确 认。 并且，对每一笔交易来说，总输入不能小于总输出。 总输入相比总输出多余的部分称为 交易费用（ Transaction Fee），为生成包含该交易区块的矿工所获得。 目前规定每笔交易的交 易费用不能小于 0.0001 BTC，交易费用越高，越多矿工愿意包含该交易，也就越早被放到 网络中。 交易费用在奖励矿工的同时，也避免了网络受到大量攻

>在比特币中，第一笔 coinbase 交易包含了如下信息：“The Times 03/Jan/2009 Chancellor on brink of second bailout for banks”
[可点击这里查看](https://blockchain.info/tx/4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b?show_adv=true)


## 更新block
从现在开始，每个区块存储的是交易数据，而不是之前的Data了，移除 Block 的 Data 字段，取而代之的是存储交易
```go
type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}
```
相应的其他的和block相关的地方都应该做相应的改变，`block.go`文件

```go
func NewBlock(transactions  []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), transactions, prevBlockHash, []byte{},0}
	...
}

func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

```
`blockchain.go`文件中
```go
func CreateBlockchain(address string) *Blockchain {
	...
	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(cbtx)

		b, err := tx.CreateBucket([]byte(blocksBucket))
		err = b.Put(genesis.Hash, genesis.Serialize())
		...
	})
	...
}
```
这个函数会接受一个地址作为参数，这个地址将会被用来接收挖出创世块的奖励。



 `blockchain.go`文件中的AddBlock方法：
```go
func (bc *Blockchain) AddBlock(transactions []*Transaction) {
	...
	newBlock := NewBlock(transactions, lastHash)
	...
}
```

## 工作量证明
和block相关的还有工作量证明
```go
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.HashTransactions(),
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}

```
这里主要就是修改`pow.block.HashTransactions()`讲解交易转化为转化为字节,编写block中的`HashTransactions`函数
```go
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]

```
我们想要通过仅仅一个哈希，就可以识别一个块里面的所有交易。为此，先获得每笔交易的哈希，然后将它们关联起来，最后获得一个连接后的组合哈希。

>比特币使用了一个更加复杂的技术：它将一个块里面包含的所有交易表示为一个  Merkle tree ，然后在工作量证明系统中使用树的根哈希（root hash）。这个方法能够让我们快速检索一个块里面是否包含了某笔交易，即只需 root hash 而无需下载所有交易即可完成判断。


## 交易核心

目前我们还没有实现关于地址的相关的信息，假设`alice`向`bob`转10个`bitcoin` ,我们知道我们的数据库死一个键值对数据库
`key` 是区块哈希，`value`是区块的数据的哈希,我们如何知道那个账户，当有很多个区块，每个区块有很多笔交易，，这时就会更复杂。
下面我们就来解决这个核心问题

找到`一个账户的所有的包含未花费输出的交易`，这一步其实相当困难:
```go
func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
  var unspentTXs []Transaction
  //string对应交易的id，[]int存储的是交易的输出的索引，
  //因为一笔交易可能有多个输出，那么属于address这个账户的是哪几个索引
  spentTXOs := make(map[string][]int)
  bci := bc.Iterator()

  for {
	//第一次执行Next()函数其实返回的是最新的一个区块
    block := bci.Next()
	//遍历区块的所有的交易
    for _, tx := range block.Transactions {
	  //交易的id是一个sha256后的字节数组，我们将其编码为字符串，方便存储
      txID := hex.EncodeToString(tx.ID)

	Outputs:
	 //遍历一笔交易所有的输出
      for outIdx, out := range tx.Vout {
        //判断交易输出是否被花费
        if spentTXOs[txID] != nil {
		  //这里只是确认一下
          for _, spentOut := range spentTXOs[txID] {
            if spentOut == outIdx {
              continue Outputs
            }
          }
        }
		//如果这笔输出经过上面验证过了，没有被花费，而且能被解锁，目前解锁过程还是比较简单的，只需要判断地址是否相等
		//那么就将该笔交易追加到`unspentTXs`切片中
        if out.CanBeUnlockedWith(address) {
          unspentTXs = append(unspentTXs, *tx)
        }
      }

	  //判断是不是`Coinbase`交易，因为`Coinbase`交易是没有输入的
      if tx.IsCoinbase() == false {
		//遍历一笔交易所有的输入，如果交易的输入嫩个
        for _, in := range tx.Vin {
		  //如果输入被这个地址解锁，那么说明之前这笔交易对应的地址输出已经被花费
          if in.CanUnlockOutputWith(address) {
			  //获取输入的交易ID
			inTxID := hex.EncodeToString(in.Txid)
			//将花费的交易id追加到花费列表中
            spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
          }
        }
      }
    }
	//已经遍历到创世区块，停止遍历
    if len(block.PrevBlockHash) == 0 {
      break
    }
  }

  return unspentTXs
}
```
`FindUnspentTransactions`函数可以找到一个账户上未为花费的交易，自然就可以查到一个账户上未被话费的交易输出
，其实也就是余额，相当于所有的未被花费的交易输出的累积

找打未被花费的输出
```go
func (bc *Blockchain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}
```
这个函数也比较简单，如果想知道金额，直接遍历这个函数结果就可以知道了。

因为比特币的交易输出是不可以再分的，所以需要从未花费的交易中找出足够多的金额来进行消费转账，一旦我们累积的
金额足够，我们就停止遍历返回。

我们需要创建`FindSpendableOutputs`函数，它接受两个参数，收钱的地址和转账的金额，返回的参数是实际找到的金额和一个map
注意这里的金额有可能遍历完都小于amount，所以转账的时候有一个判断，金额够才会转账
```go
func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	//string对应交易id,int是输出的索引
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
```
这个函数其实就比较简单了，就是从未被话费的交易中找出金额，同时返回的还有通过交易 ID 进行分组的输出索引


## 转账


发送币意味着创建新的交易，并通过挖出新块的方式将交易打包到区块链中。

不过，比特币并不是一连串立刻完成这些事情（虽然我们目前的实现是这么做的）。相反，它会将所有新的交易放到一个内存池中（mempool），然后当矿工准备挖出一个新块时，它就从内存池中取出所有交易，创建一个候选块。只有当包含这些交易的块被挖出来，并添加到区块链以后，里面的交易才开始确认。

先定义转账交易
```go
func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	
	acc, validOutputs := bc.FindSpendableOutputs(from, amount)
	//金额足够才会转账
	if acc < amount {
		log.Panic("ERROR: Not enough funds")
	}

	// 从可以花费的交易中创建输入
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)

		for _, out := range outs {
			input := TXInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}

	// 创建输出
	outputs = append(outputs, TXOutput{amount, to})
	if acc > amount {
		outputs = append(outputs, TXOutput{acc - amount, from}) // a change
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx
}
```
转账的操作其实就是创建一个新的`transaction`的过程.

## 更新CLI
我们希望所有的操作都能从命令行进行，而不用一次一次修改`main.go`文件来执行我们的操作

具体的步骤就是
```
1.将所有可能的操作都通过命令行来质性，也就是在cli.go文件中注册命令
2.cli中监听并解析相应的命令，就可以执行对应的函数
3.对应的函数就会去执行区块链的相关操作
```

```go
func (cli *CLI) send(from, to string, amount int) {
	bc := NewBlockchain(from)
	defer bc.db.Close()

	tx := NewUTXOTransaction(from, to, amount, bc)
	bc.MineBlock([]*Transaction{tx})
	fmt.Println("Success!")
}
```
获取余额
```go
func (cli *CLI) getBalance(address string) {
	//主要是获取当前区块链最新的区块的hash
	bc := NewBlockchain()
	defer bc.db.Close()

	balance := 0
	UTXOs := bc.FindUTXO(address)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}
```

后面的注册命令，和用法和以前一样

别忘了修改`main`函数
```go
func main() {
	cli := CLI{}
	cli.Run()
}
```
## 测试

创建区块链
```shell
D:\Download\code\goProject\src\bitcoin\version5>version5.exe createblockchain -address "alice"
0000d5517255276f0ecf5c2d1c33cadafac7523107cf111b157cf6da01e33eaa

Done!
```
打印区块链
```shell
D:\Download\code\goProject\src\bitcoin\version5>version5.exe printchain
Prev. hash:
Hash: 0000d5517255276f0ecf5c2d1c33cadafac7523107cf111b157cf6da01e33eaa
PoW: true
```
查询余额
```shell
D:\Download\code\goProject\src\bitcoin\version5>version5.exe getbalance -address "alice"
Balance of 'alice': 50
```
转账
```go
D:\Download\code\goProject\src\bitcoin\version5>version5.exe send -from "alice" -to "bob" -amount 20
0000123ba1bae85336cb0d174d013cf43b3e2e7afef56735e22f6252aa5dc391

Success!

```
查询余额
```shell
D:\Download\code\goProject\src\bitcoin\version5>version5.exe getbalance -address "bob"
Balance of 'bob': 20
```
打印区块链
```shell
D:\Download\code\goProject\src\bitcoin\version5>version5.exe getbalance -address "bob"
Balance of 'bob': 20

D:\Download\code\goProject\src\bitcoin\version5>version5.exe printchain
Prev. hash: 0000d5517255276f0ecf5c2d1c33cadafac7523107cf111b157cf6da01e33eaa
Hash: 0000123ba1bae85336cb0d174d013cf43b3e2e7afef56735e22f6252aa5dc391
PoW: true

Prev. hash:
Hash: 0000d5517255276f0ecf5c2d1c33cadafac7523107cf111b157cf6da01e33eaa
PoW: true
```
包含一笔创世交易生成的区块和转账交易生成的区块

