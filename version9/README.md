目前我们所构建的原型已经具备了区块链所有的关键特性：匿名，安全，随机生成的地址；区块链数据存储；工作量证明系统；可靠地存储交易。

区块链网络就是一个程序社区，里面的每个程序都遵循同样的规则，正是由于遵循着同一个规则，才使得网络能够长存

## 区块链网络

区块链网络是去中心化的，这意味着没有服务器，客户端也不需要依赖服务器来获取或处理数据。或者可以这么理解区块链的每个节点即使客户端，又是服务器，

每个节点必须与很多其他节点进行交互，它必须请求其他节点的状态，与自己的状态进行比较，当状态过时时进行更新。

区块链网络中一般分为三种角色


矿工:矿工是区块链中唯一可能会用到工作量证明的角色，因为挖矿实际上意味着解决 PoW 难题,他们一般运行的就是全节点。

全节点：些节点验证矿工挖出来的块的有效性，并对交易进行确认。为此，他们必须拥有区块链的完整拷贝。同时，全节点执行路由操作，帮助其他节点发现彼此。对于网络来说，非常重要的一段就是要有足够多的全节点。因为正是这些节点执行了决策功能：他们决定了一个块或一笔交易的有效性。

SPV(Simplified Payment Verification)： 他的主要功能在于一个人不需要下载整个区块链，但是仍能够验证他的交易。


## 本次模拟
我们没有那么多的计算机来模拟一个多节点的网络。当然，我们可以使用虚拟机或是 Docker 来解决这个问题，但是这会使一切都变得更复杂。

所以，我们想要在一台机器上运行多个区块链节点，同时希望它们有不同的地址。为了实现这一点，我们将使用端口号作为节点标识符，而不是使用 IP 地址，比如将会有这样地址的节点：127.0.0.1:3000，127.0.0.1:3001，127.0.0.1:3002 等等。我们叫它端口节点（port node） ID，并使用环境变量 NODE_ID 对它们进行设置。故而，你可以打开多个终端窗口，设置不同的 NODE_ID 运行不同的节点。

个方法也需要有不同的区块链和钱包文件。它们现在必须依赖于节点 ID 进行命名，比如 blockchain_3000.db, blockchain_3001.db and wallet_3000.db, wallet_3001.db 等等

因此我们必须修改我们存储时的文件。

当你启动一个全新比特币节点时，它会连接到一个种子节点，获取全节点列表，随后从这些节点中下载区块链

不过在我们目前的实现中，无法做到完全的去中心化，因为会出现中心化的特点。我们会有三个节点：
```shell
一个中心节点。所有其他节点都会连接到这个节点，这个节点会在其他节点之间发送数据。

一个矿工节点。这个节点会在内存池中存储新的交易，当有足够的交易时，它就会打包挖出一个新块。

一个钱包节点。这个节点会被用作在钱包之间发送币。但是与 SPV 节点不同，它存储了区块链的一个完整副本
```


## 场景

```shell
中心节点创建一个区块链。
一个其他（钱包）节点连接到中心节点并下载区块链。
另一个（矿工）节点连接到中心节点并下载区块链。
钱包节点创建一笔交易。
矿工节点接收交易，并将交易保存到内存池中。
当内存池中有足够的交易时，矿工开始挖一个新块。
当挖出一个新块后，将其发送到中心节点。
钱包节点与中心节点进行同步。
钱包节点的用户检查他们的支付是否成功。
这就是比特币中的一般流程。尽管我们不会实

```

## 代码

网络中方便传输,我们发送消息的前12个字节用于让服务器角色识别是什么请求或者说是接受到什么消息，其他的都是我们的消息数据
```go
func commandToBytes(command string) []byte {
    var bytes [commandLength]byte

    for i, c := range command {
        bytes[i] = byte(c)
    }

    return bytes[:]
}


func bytesToCommand(bytes []byte) string {
    var command []byte

    for _, b := range bytes {
        if b != 0x0 {
            command = append(command, b)
        }
    }
    return fmt.Sprintf("%s", command)
}
```

## version消息

中心节点是已知的，当你启动一个新的节点的时候，会给中心节点，发送响应的消息，来获取最新的区块链消息
```go
type version struct {
    //区块链版本
    Version    int
    //高度
    BestHeight int
    //发送消息地址
    AddrFrom   string
}
```
由于我们仅有一个区块链版本，所以 Version 字段实际并不会存储什么重要信息。

注意我们一个节点及时客户端又是服务器，所以我们需要在同一个程序中编写我们的服务器和客户端代码

消息传递我们需要一个服务器

1.首先 启动新的节点，目前只知道中心节点，给中心节点 `sendVersion`，获取新的区块列表, 如果本节点区块高，`sendVersion`
```go
func StartServer(nodeID, minerAddress string) {
    //本机地址
    nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
    //指定矿工地址
	miningAddress = minerAddress
	ln, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()

	bc := NewBlockchain(nodeID)

    //knownNodes[0]是默认指定的中心节点，如果是其他节点
    //必须向中心节点发送 version 消息通过区块高度来查询是否自己的区块链已过时
	if nodeAddress != knownNodes[0] {
		sendVersion(knownNodes[0], bc)
	}

    //一直监听，当有消息来时就进行处理
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
        }
		go handleConnection(conn, bc)
	}
}

```

2.`sendVersion`过程，将本节点的区块链版本，区块高度，节点地址组成消息发送 `version`命令消息
```go
//发送version消息
func sendVersion(addr string, bc *Blockchain) {
    //获取自己的区块高度
    bestHeight := bc.GetBestHeight()
    //打包
	payload := gobEncode(verzion{nodeVersion, bestHeight, nodeAddress})
    //发送version信息
	request := append(commandToBytes("version"), payload...)

	sendData(addr, request)
}

//建立连接发送数据发送
//addr表示向谁发送，
func sendData(addr string, data []byte) {
	//建立连接
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		fmt.Printf("%s is not available\n", addr)
		var updatedNodes []string
		//去除连接不上的节点，更新已知节点信息
		for _, node := range knownNodes {
			if node != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}
		knownNodes = updatedNodes

		return
	}
	defer conn.Close()

	//表示先创建一个缓冲区，存储一定数据在发送数据，而不是一个一个发送，效率上会高很多
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

```

3.处理`version`消息，`handleVersion`,接受`version`消息的节点对比本地区块高度，

本地区块高度高：发送本地版本信息，对方接受到版本消息后处理就会请求区块列表

本地区块高度低：本地发送请求缺失区块的的列表
```go
//处理version消息
func handleVersion(request []byte, bc *Blockchain) {

    ...
    
	myBestHeight := bc.GetBestHeight()
	foreignerBestHeight := payload.BestHeight
	if myBestHeight < foreignerBestHeight {
		//获取我要下载的block的列表
		sendGetBlocks(payload.AddrFrom)
		//大于发送版本消息
	} else if myBestHeight > foreignerBestHeight {
		sendVersion(payload.AddrFrom, bc)
	}

	//sendAddr(payload.AddrFrom)
	//如果是一个未知节点，将未知节点加入进去
	if !nodeIsKnown(payload.AddrFrom) {
		knownNodes = append(knownNodes, payload.AddrFrom)
	}
}

```
4.请求区块列表,发送getblocks指令，消息为本节点地址
```go
type getblocks struct {
	AddrFrom string
}
//这个函数主要用于获取其他节点多个区块hash组成的数组，
func sendGetBlocks(address string) {
	payload := gobEncode(getblocks{nodeAddress})
	request := append(commandToBytes("getblocks"), payload...)
	//
	sendData(address, request)
}
```
5.`handlegetblocks`处理`getblocks`请求,迭代本地所有的区块hash添加到数组中, 给请求节点发送inv消息，
```go

func handleGetBlocks(request []byte, bc *Blockchain) {

    ...
    
	blocks := bc.GetBlockHashes()
	sendInv(payload.AddrFrom, "block", blocks)
}

//发送清单消息
type inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

func sendInv(address, kind string, items [][]byte) {
	inventory := inv{nodeAddress, kind, items}
	payload := gobEncode(inventory)
	request := append(commandToBytes("inv"), payload...)

	sendData(address, request)
}

```
6.处理inv消息

如果inv命令收到的是区块，那么请求消息体中的一个区块，
如果inv命令收到的是交易

```go
//子节点接受清单列表
func handleInv(request []byte, bc *Blockchain) {

    ...
    
	fmt.Printf("Recevied inventory with %d %s\n", len(payload.Items), payload.Type)

	if payload.Type == "block" {
        //将收到块哈希它们保存在 blocksInTransit 变量来跟踪已下载的块
		blocksInTransit = payload.Items
        //
		blockHash := payload.Items[0]
		//发送下载区块消息
		sendGetData(payload.AddrFrom, "block", blockHash)
        
        //去掉已经请求的hash
		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}


	if payload.Type == "tx" {
		txID := payload.Items[0]
		//如果发现别人发过来的交易本地没有，那么就获取这个交易
		if mempool[hex.EncodeToString(txID)].ID == nil {
			sendGetData(payload.AddrFrom, "tx", txID)
		}
	}
}

```

		
7.根据区块hash请求数据 `sendGetData`
```go

type getdata struct {
    //发出请求的节点地址
    AddrFrom string
    //单个交易id还是单个区块id
    Type     string
    ID       []byte
}

//用于某个块或交易的请求，它可以仅包含一个块或交易的Id
func sendGetData(address, kind string, id []byte) {
	payload := gobEncode(getdata{nodeAddress, kind, id})
	request := append(commandToBytes("getdata"), payload...)

	sendData(address, request)
}
```


8,处理getData消息
```go
//如果它们请求一个块，则返回块；如果它们请求一笔交易，则返回交易,我们暂时我们并不检查实际上是否已经有了这个块或交易。这是一个缺陷 :)
func handleGetData(request []byte, bc *Blockchain) {
    
    ...
    
	if payload.Type == "block" {
		block, err := bc.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}

		sendBlock(payload.AddrFrom, &block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := mempool[txID]

		sendTx(payload.AddrFrom, &tx)
		// delete(mempool, txID)
	}
}



```
9，发送单个区块或者交易请求，交易这边就省略了
```go
type block struct {
	AddrFrom string
	Block    []byte
}

//节点给其他节点发送一个区块区块
func sendBlock(addr string, b *Block) {
	data := block{nodeAddress, b.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("block"), payload...)

	sendData(addr, request)
}

```
10。处理接受单个区块区块或者交易，前面那么多请求数据实际完成数据转移的正是这些消息。
```go
//处理区块
func handleBlock(request []byte, bc *Blockchain) {
    ... 
    //反序列化数据添加到区块链中
    blockData := payload.Block
    block := DeserializeBlock(blockData)

    fmt.Println("Recevied a new block!")
    bc.AddBlock(block)

    fmt.Printf("Added block %x\n", block.Hash)
    //如果还有更多的区块需要下载，我们继续从上一个下载的块的那个节点继续请求
    if len(blocksInTransit) > 0 {
        blockHash := blocksInTransit[0]
        sendGetData(payload.AddrFrom, "block", blockHash)

        blocksInTransit = blocksInTransit[1:]
    //当最后把所有块都下载完后，对 UTXO 集进行重新索引
    } else {
        UTXOSet := UTXOSet{bc}
        UTXOSet.Reindex()
    }
}
```
处理接受到的交易
```go
func handleTx(request []byte, bc *Blockchain) {
    ...
    txData := payload.Transaction
    tx := DeserializeTransaction(txData)
    //将接收到的交易添加到本地的内存池中
    mempool[hex.EncodeToString(tx.ID)] = tx
    //如果本机节点是中心节点
    if nodeAddress == knownNodes[0] {
        for _, node := range knownNodes {
            if node != nodeAddress && node != payload.AddFrom {
                //需要给其他结点发送该笔交易的信息
                sendInv(node, "tx", [][]byte{tx.ID})
            }
        }
    } else {
        //有挖矿节点且内存池的交易数量大于2
        if len(mempool) >= 2 && len(miningAddress) > 0 {
        MineTransactions:
            var txs []*Transaction
            //遍历内存池中的交易，验证每一遍交易是否合法
            for id := range mempool {
                tx := mempool[id]
                if bc.VerifyTransaction(&tx) {
                    txs = append(txs, &tx)
                }
            }
            //通过验证后的交易如果数量等于0
            if len(txs) == 0 {
                fmt.Println("All transactions are invalid! Waiting for new ones...")
                return
            }
            //每个区块都需要一笔币基交易
            cbTx := NewCoinbaseTX(miningAddress, "")
            txs = append(txs, cbTx)
            //挖出新区快
            newBlock := bc.MineBlock(txs)
            //更新本地的UTXO集
            UTXOSet := UTXOSet{bc}
            UTXOSet.Reindex()

            fmt.Println("New block is mined!")
            //从内存池中删除挖矿中已经发生的交易
            for _, tx := range txs {
                txID := hex.EncodeToString(tx.ID)
                delete(mempool, txID)
            }

            //将新挖出的区块发送给其他节点，注意是通过Inv命令，下面会介绍sendBlock和sendInv的区别
            for _, node := range knownNodes {
                if node != nodeAddress {
                    sendInv(node, "block", [][]byte{newBlock.Hash})
                }
            }
            //挖出新的区块后如果发现内存池还有其他交易，就继续挖矿   
            if len(mempool) > 0 {
                goto MineTransactions
            }
        }
    }
}

```
>注意handleInv和 handleTx,handleBlock的区别，对比一下发送的数据好像基本差不多，但是逻辑是完全不同的
handleInv主要是因为本地没有区块或者交易，需要同步到本地，用的inv命令额外区分block的区别，不然两个都是 `block`,
当我接收到`block`指令，我怎么知道我是该处理接受一个区块的逻辑，还是处理给你发送一个区块的逻辑，一个节点既是服务器又是客户端。

## 测试

测试我们按照如下过程

中心节点：只创建区块链，并向网络中的其他节点广播交易                           3000
一个钱包节点：可以创建交易，然后发送给中心节点，可以同步区块数据，但是不挖矿   3001
挖矿节点：从中心节点接受交易，挖矿，广播区块                                  3002


本地开启三个终端,以window为例

### 3000节点
```shell
# 指定NODE_ID临时变量为3000，这里以window为例
set NODE_ID=3000
```
创建钱包和区块链
```shell
blcokchain.exe createwallet 
# 生成地址 wallet1
blockchain.exe createblockchain -address wallet1

# 复制生成的数据
copy blockchain_3000.db blockchain_genesis.db 
```
生成了一个仅包含创世块的区块链。copy数据的目的在于我们需要保存块，并在其他节点使用。创世块承担了一条链标识符的角色（在 Bitcoin Core 中，创世块是硬编码的），

后面我们会
```shell
copy blockchain_genesis.db blockchain_3001.db
copy blockchain_genesis.db blockchain_3002.db
```
`copy blockchain_3000.db blockchain_genesis.db`的目的在于我们所有节点必须是同一个区块链，那么创世的区块的数据必须保证一致，今后的
网络中的交易最后会同步到自己的数据中，这样其他结点就不用创建区块链了。

### 3001节点
```shell
# 指定NODE_ID临时变量为3001
set NODE_ID=3001
```
生成多个钱包地址WALLET2, WALLET3, WALLE4，WALLE5，执行多次,这里生成4个
```shell
blcokchain.exe createwallet 
```
### 3000节点

给3001节点的钱包地址发送一些币:此时只有wallet1才有钱
```shell
blcokchain.exe  send -from wallet1 -to WALLET2 -amount 20 -mine
blcokchain.exe  send -from wallet1 -to WALLET3 -amount 20 -mine
```
如果没有-mine，本节点是不能挖矿的，必须要有这个标志，因为初始状态时，网络中没有矿工节点，预定的3002节点还没有启动。

启动节点
```shell
blockchain.exe startnode
```

### 3001节点
复制上面保存创世块节点的区块链：
```shell
copy blockchain_genesis.db blockchain_3001.db
```
启动3001节点 
```
blockchain.exe startnode
```
它会给中心结点`发送version信息`进行数据交互，中心节点现在已经有三个区块了，下载所有的区块。为了检查一切正常，暂停节点运行并检查余额：
```
$ blockchain.exe getbalance -address WALLET2
Balance of 'WALLET2': 20

$ blockchain.exe getbalance -address WALLET3
Balance of 'WALLET3': 20
```
你还可以检查 wallet1 地址的余额，因为 node 3001 现在有它自己的区块链：
$ blockchain.exe getbalance -address wallet1
Balance of 'CENTRAL_NODE': 110

## 节点3002

指定NODE_ID临时变量为3002
```shell
set NODE_ID=3002
```
创建钱包地址wallet_miner
```
blockchain.exe createwallet
```
初始化区块链
```
copy blockchain_genesis.db blockchain_3002.db
```
启动节点
```
blockchain.exe startnode -miner MINER_WALLET
```

## 节点3001
执行交易
```
$ blockchain.exe send -from WALLET2 -to WALLET4 -amount 10
$ blockchain.exe send -from WALLET3 -to WALLET5 -amount 15
```

## 节点3002

迅速切换到矿工节点，你会看到挖出了一个新块！同时，检查中心节点的输出

## 节点3001
切换到钱包节点并启动：
```shell
blockchain.exe startnode
```
他会下载刚刚节点3002产生的区块，检查节点3001四个地址的余额...

- [基本原型](https://github.com/cray666/bitcoin/tree/master/version1)
- [工作量证明](https://github.com/cray666/bitcoin/tree/master/version2)
- [数据库存储](https://github.com/cray666/bitcoin/tree/master/version3)
- [CLI](https://github.com/cray666/bitcoin/tree/master/version4)
- [交易一](https://github.com/cray666/bitcoin/tree/master/version5)
- [交易二](https://github.com/cray666/bitcoin/tree/master/version6)
- [UTXO集](https://github.com/cray666/bitcoin/tree/master/version7)
- [Merkle树](https://github.com/cray666/bitcoin/tree/master/version8)
- [网络](https://github.com/cray666/bitcoin/tree/master/version9)

