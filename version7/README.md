回想一下我们是怎么找到有未花费输出的交易。我们是通过`NewBlockchain`获取我们最新区块标志位tip，tip指向最新区块的hash,
然后对区块hash进行迭代，一次从数据中读取每一个区块数据，检查每一笔交易

到目前为止，在比特币中已经超出 60万个区块，整个数据库所需磁盘空间超过 150 Gb。这意味着一个人如果想要验证交易，必须要运行一个全节点。此外，验证交易将会需要在许多块上进行迭代。这是非常耗时的,我们需要一个方法来解决这个问题...

## UTXO集

针对上述问题我们的想法是：我们


首先肯定UTXO集是和区块相关连的.因此定义结构如下
```go
type UTXOSet struct {
    Blockchain *Blockchain
}
```
我们的UTXO集肯定也是需要持久化，因为初始化UTXO集也是一份非常耗时的工作，幸运的是我们不用像之前，每次交易就需要重新迭代区块，这次我们只需要一次就可以，以后每次交易我们只需要更新即可,经常就是增删一些UTXO

```go
func (u UTXOSet) Reindex() {
	db := u.Blockchain.db
	bucketName := []byte(utxoBucket)
	//
	err := db.Update(func(tx *bolt.Tx) error {
		//首先，如果 bucket 存在就先移除，
		err := tx.DeleteBucket(bucketName)
		if err != nil && err != bolt.ErrBucketNotFound {
			log.Panic(err)
		}

		_, err = tx.CreateBucket(bucketName)
		HandleError(err)

		return nil
	})
	HandleError(err)
	//这就是我们的核心逻辑，找出UTXO集，其实和之前的找未被话费的交易逻辑差不多
	UTXO := u.Blockchain.FindUTXO()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			HandleError(err)
			err = b.Put(key, outs.Serialize())
			HandleError(err)
		}

		return nil
	})
}

func (bc *Blockchain) FindUTXO() map[string]TXOutputs {
	

```

`Blockchain.FindUTXO` 几乎跟 `Blockchain.FindUnspentTransactions` 一模一样`，但是现在它返回了一个 TransactionID -> TransactionOutputs 的 map。string存储交易id,TXOutputs是多个输出，允许一笔交易一个账户可以有多个UTXO。多个UTXO都可以，单个自然也没什么问题。

现在现在，UTXO 集可以用于发送币：我们可以直接找出未话费的交易输出，而不再是找交易了，
```go
func (u UTXOSet) FindSpendableOutputs(pubkeyHash []byte, amount int) (int, map[string][]int) {
	//交易id->交易输出索引切片
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	db := u.Blockchain.db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs := DeserializeOutputs(v)

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubkeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				}
			}
		}

		return nil
	})

	return accumulated, unspentOutputs
}
```
当前现在我们也可以找出某个账户的余额，现在我们不用去遍历区块查找，而是每次都去找UTXO，区块负责存储我们的交易数据
UTXO集负责存储我们的UTXO
```go
func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	db := u.Blockchain.db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := DeserializeOutputs(v)

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return UTXOs
}
```
有了 UTXO 集，也就意味着我们的数据（交易）现在已经被分开存储：实际交易被存储在区块链中，未花费输出被存储在 UTXO 集中。这样一来，我们就需要一个良好的同步机制，因为我们想要 UTXO 集时刻处于最新状态，并且存储最新交易的输出。但是我们不想每生成一个新块，就重新生成索引，因为这正是我们要极力避免的频繁区块链扫描。

因此我们区块链更新一个区块我们就需要更新我们的UTXO集，UTXO集的增删。
```go
func (u UTXOSet) Update(block *Block) {
	
    db := u.Blockchain.db
	//打开读写事务
    err := db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(utxoBucket))
	//遍历新增区块的交易
        for _, tx := range block.Transactions {
            if tx.IsCoinbase() == false {
		//遍历每一笔输入	
                for _, vin := range tx.Vin {
		    updatedOuts := TXOutputs{}
		    //根据交易id得到引用之前发生的交易输出的序列化数据
		    outsBytes := b.Get(vin.Txid)
		    //反序列化数据得到之前交易的输出
                    outs := DeserializeOutputs(outsBytes)
		    //遍历输出
                    for outIdx, out := range outs.Outputs {
			//将之前的每一笔交易输出中没有被此次生成区块所引用的输出添加到updatedOuts.Outputs
                        if outIdx != vin.Vout {
                            updatedOuts.Outputs = append(updatedOuts.Outputs, out)
                        }
                    }
		     //如果之前某笔交易的输出都被引用了，那么就从UTXO集合中删除该笔交易
		    //否则就更新
                    if len(updatedOuts.Outputs) == 0 {
                        err := b.Delete(vin.Txid)
                    } else {
                        err := b.Put(vin.Txid, updatedOuts.Serialize())
                    }

                }
            }
	    //同时将新得区块的每一集交易的产生的新的UTXO加入到UTXO集中
            newOutputs := TXOutputs{}
            for _, out := range tx.Vout {
                newOutputs.Outputs = append(newOutputs.Outputs, out)
            }

            err := b.Put(tx.ID, newOutputs.Serialize())
        }
    })
}
```
具体更新UTXO集的步骤其实也很简单：
```
更新区快产生所有引用的之前的交易的UTXO
添加新区快产生所产生的UTXO
```
那么现在我们来测试一下吧



## 测试

创建地址
```shell
D:\Download\code\goProject\src\bitcoin\version7>version7.exe createwallet
Your new address: 1cMN5foxpuJGqYerncCQbWuE4Q4zi56v4

D:\Download\code\goProject\src\bitcoin\version7>version7.exe createwallet
Your new address: 1Jmy4SQTzeGUzsfLHMzQ5dAQ2Mtbmom8mD

```
创建区块链

```shell
D:\Download\code\goProject\src\bitcoin\version7>version7.exe createblockchain -address  1cMN5foxpuJGqYerncCQbWuE4Q4zi56v4
00000c0965957384833aa964112a51682ef15fa5ad6d8fee860c6afa10e4f8f0

Done!

```
查询余额
```shell

D:\Download\code\goProject\src\bitcoin\version7>version7.exe getbalance -address 1cMN5foxpuJGqYerncCQbWuE4Q4zi56v4
Balance of '1cMN5foxpuJGqYerncCQbWuE4Q4zi56v4': 50

```
转账
```
D:\Download\code\goProject\src\bitcoin\version7>version7.exe send  -from  1cMN5foxpuJGqYerncCQbWuE4Q4zi56v4  -to 1Jmy4SQTzeGUzsfLHMzQ5dAQ2Mtbmom8mD  -amount 20
00004442ff2f425abc402ed4e8ed3d2df785a0bd901869f53f66d871a29e7967

Success!
```
查询余额

```
D:\Download\code\goProject\src\bitcoin\version7>version7.exe getbalance -address 1cMN5foxpuJGqYerncCQbWuE4Q4zi56v4
Balance of '1cMN5foxpuJGqYerncCQbWuE4Q4zi56v4': 80

D:\Download\code\goProject\src\bitcoin\version7>version7.exe getbalance -address 1Jmy4SQTzeGUzsfLHMzQ5dAQ2Mtbmom8mD
Balance of '1Jmy4SQTzeGUzsfLHMzQ5dAQ2Mtbmom8mD': 20
```
>注意，因为挖矿挖出每一个区块，都有区块链奖励，我们这里奖励就固定给转账人了，所以每次转账人的金额都会加上奖励的金额
打印区块链信息
```
============ Block 00004442ff2f425abc402ed4e8ed3d2df785a0bd901869f53f66d871a29e7967 ============
Prev. block: 00000c0965957384833aa964112a51682ef15fa5ad6d8fee860c6afa10e4f8f0
PoW: true

--- Transaction cea5ab6d0122715e60b6f178f9753f1de635bde3800f43bd56717415ed1e5630:
     Input 0:
       TXID:
       Out:       -1
       Signature:
       PubKey:    32613866633535373738396436656638303031643936323837323237343531393734383831353730
     Output 0:
       Value:  50
       Script: 06af8ea5ff916d427e6e043c64ad17b168ba3152
--- Transaction fbbfb0b5655044cf1f6e7a57316c6cf4ecee672aa78e8224ef727995ca230bf1:
     Input 0:
       TXID:      c55ded508afa2c80cde2969994e1beba95ac88d49b18f49bde5358536842c8bf
       Out:       0
       Signature: 70d9b3b848ee14128893f1e6a6cd90f31bfe324720edfd8bb647dc788acc3ff9e44ad5f2b946031ce9231ffb297de05b44d48ad16609a8f0bfa10ba6bd4d5f5d
       PubKey:    0a7c02cc25f81b103bda5ecf6089a90f893ee21702402a99ce43f9c878b64a1967cbc97cae935b1ecafb49be7f655a00dcf1243270194e69e2fa5f74fa931762
     Output 0:
       Value:  20
       Script: c2fb3c8fd93609eec9e36e50d3cd873c62d24ede
     Output 1:
       Value:  30
       Script: 06af8ea5ff916d427e6e043c64ad17b168ba3152


============ Block 00000c0965957384833aa964112a51682ef15fa5ad6d8fee860c6afa10e4f8f0 ============
Prev. block:
PoW: true

--- Transaction c55ded508afa2c80cde2969994e1beba95ac88d49b18f49bde5358536842c8bf:
     Input 0:
       TXID:
       Out:       -1
       Signature:
       PubKey:    5468652054696d65732030332f4a616e2f32303039204368616e63656c6c6f72206f6e206272696e6b206f66207365636f6e64206261696c6f757420666f722062616e6b73
     Output 0:
       Value:  50
       Script: 06af8ea5ff916d427e6e043c64ad17b168ba3152

```
我们已经可以非常清楚的看到区块的相关信息了。


