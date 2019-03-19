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

`version消息`打你启动一个新的节点的时候，他会给中心节点，发送消息
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

    //一直坚挺，当有消息来时就进行处理
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
        }
		go handleConnection(conn, bc)
	}
}

```

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


当一个节点接受到连接消息后，会运行 bytesToCommand 来提取命令名，并选择消息类型进行处理
```go
func handleConnection(conn net.Conn, bc *Blockchain) {
    request, err := ioutil.ReadAll(conn)
    command := bytesToCommand(request[:commandLength])
    fmt.Printf("Received %s command\n", command)

    switch command {
    ...
    case "version":
        handleVersion(request, bc)
    default:
        fmt.Println("Unknown command!")
    }

    conn.Close()
}
```

消息类型很多种，每一个send类型就有一个handle类型。具体的实现我们看代码。