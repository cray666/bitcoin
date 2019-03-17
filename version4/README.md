到目前为止，我们的实现还没有提供一个与程序交互的接口：目前只是在 main 函数中简单执行了 NewBlockchain 和 bc.AddBlock 。是时候改变了！现在我们想要拥有这些命令...

## CLI

我们希望通过在命令行通过命令来执行生成创世区块，添加区块，查询区块操作，而不是将固定的操作永远写在main()函数里面

```
bitcoin addblock "Pay 0.031337 for a coffee"
bitoin  printchain
```
bitcoin是我们编译后生成的程序

所有命令行相关的操作都会通过 `CLI` 结构进行处理，也就是我们所有的生成区块链，添加区块等操作都会通过`CLI`相关函数来监听,下面来定义
`CLi`的结构，
```
type CLI struct {
	bc *Blockchain
}
```
因为添加区块，打印区块等相关操作都是和Blockchain绑定的，自然我们cli的结构需要绑定一一个这样的结构

入口函数为
```go
func (cli *CLI) Run() {
	cli.validateArgs()

	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	addBlockData := addBlockCmd.String("data", "", "Block data")

	switch os.Args[1] {
	case "addblock":
		err := addBlockCmd.Parse(os.Args[2:])
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			os.Exit(1)
		}
		cli.addBlock(*addBlockData)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}
}
```
我们会使用标准库里面的 flag 包来解析命令行参数：
```go
//添加子命令addblock
addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
//添加子命令printchain
printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
//addblock 添加 -data 标志
addBlockData := addBlockCmd.String("data", "", "Block data")
```
然后我们检查用户提供的命令，解析相关的flag命令
```go
switch os.Args[1] {
	case "addblock":
		err := addBlockCmd.Parse(os.Args[2:])
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
	default:
		cli.printUsage()
		os.Exit(1)
	}
```
根据解析的相关命令，调用相关的函数
```go
if addBlockCmd.Parsed() {
	if *addBlockData == "" {
		addBlockCmd.Usage()
		os.Exit(1)
	}
	cli.addBlock(*addBlockData)
}

if printChainCmd.Parsed() {
	cli.printChain()
}
```
那么我们该定义我们相关的函数：
```go
func (cli *CLI) addBlock(data string) {
	cli.bc.AddBlock(data)
	fmt.Println("Success!")
}
```
打印所有的区块信息的函数，将之前的main函数复制 过来即可，
```go
func (cli *CLI) printChain() {
	bci := cli.bc.Iterator()

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
最后，对main函数进行修改，注意defer关闭数据库连接的用法。
```go
func main() {
	bc := NewBlockchain()
	defer bc.db.Close()

	cli := CLI{bc}
	cli.Run()
}
```
来检查是否工作正常，编译运行即可
```shell
D:\Download\code\goProject\src\bitcoin\version4>blockchain.exe addblock -data "send 1BTC to Alice"
Mining the block containing "Genesis Block"
00005166640cb0e9499898d318060c9832f074440ea9341aace833a2789d78e5

Mining the block containing "send 1BTC to Alice"
000088e443340e2d41a2b8261a4c93ab411239c33caeb8e5b4a87fa6d5dca70f

Success!
```
我们可以看出，新建区块链的时候挖出来创世区块，然后又挖出了一个新的区块，再次执行试试
```shell
D:\Download\code\goProject\src\bitcoin\version4>blockchain.exe addblock -data "send 2BTC to Bob"
Mining the block containing "send 2BTC to Bob"
0000b35199e18b04f61356a30fd8f98fd6683f426785f038f3f5275609f749e6

Success!
```
下面我们来打印我们的所有的区块详情，看是不是在数据库存储正常。
```shell
D:\Download\code\goProject\src\bitcoin\version4>blockchain.exe printchain
Prev. hash: 000088e443340e2d41a2b8261a4c93ab411239c33caeb8e5b4a87fa6d5dca70f
Data: send 2BTC to Bob
Hash: 0000b35199e18b04f61356a30fd8f98fd6683f426785f038f3f5275609f749e6
PoW: true

Prev. hash: 00005166640cb0e9499898d318060c9832f074440ea9341aace833a2789d78e5
Data: send 1BTC to Alice
Hash: 000088e443340e2d41a2b8261a4c93ab411239c33caeb8e5b4a87fa6d5dca70f
PoW: true

Prev. hash:
Data: Genesis Block
Hash: 00005166640cb0e9499898d318060c9832f074440ea9341aace833a2789d78e5
PoW: true
```
结果和预期一样。打印区块的结构也是从最后一个区块一次往前打印。

