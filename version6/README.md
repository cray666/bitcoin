我们已经初步实现了交易。但是我们的账户目前没有丝毫"个人"色彩的存在,都是以alice，bob这样的字符串代替：在比特币中，没有用户账户，不需要也不会在任何地方存储个人数据（比如姓名，护照号码或者 SSN）。但是，我们总要有某种途径识别出你是交易输出的所有者（也就是说，你拥有在这些输出上锁定的币）。这就是比特币地址（address）需要完成的使命。

之前我们把一个由用户定义的任意字符串当成是地址，现在我们将要实现一个跟比特币一样的真实地址。

。

## 公钥加密

公钥加密（public-key cryptography）算法使用的是成对的密钥：公钥和私钥。公钥并不是敏感信息，可以告诉其他人。但是，私钥绝对不能告诉其他人：只有所有者（owner）才能知道私钥，能够识别，鉴定和证明所有者身份的就是私钥。在加密货币的世界中，你的私钥代表的就是你，私钥就是一切。

本质上，比特币钱包也只不过是这样的密钥对而已。当你安装一个钱包应用，或是使用一个比特币客户端来生成一个新地址时，它就会为你生成一对密钥。在比特币中，谁拥有了私钥，谁就可以控制所有发送到这个公钥的币。

私钥和公钥只不过是随机的字节序列，因此它们无法在屏幕上打印，人类也无法通过肉眼去读取。这就是为什么比特币使用了一个转换算法，将公钥转化为一个人类可读的字符串（也就是我们看到的地址）。

如果你用过比特币钱包应用，很可能它会为你生成一个助记符。这样的助记符可以用来替代私钥，并且可以被用于生成私钥。BIP-039 已经实现了这个机制。

好了，现在我们已经知道了在比特币中证明用户身份的是私钥。那么，比特币如何检查交易输出（和存储在里面的币）的所有权呢？


## 数字签名

有一个典型的场景， Alice通过信道发给Bob一个文件（一份信息），Bob 如何获知所收到的文件即为 Alice 发出的原始版本？ 
```
确认是alice发过来的
确认是发过来的内容没有发生改变
```
Alice 可以先对文件内容进行摘要，然后用自己的私钥对摘要进行加密（签名），之后同时将文件和签名都发给 Bob。 

Bob 收到文件和签名后， 用 Alice 的公钥来解密签名，得到数字摘要，与收到文件进行摘要后的结果进行比对。如果 一致，说明该文件确实是 Alice 发过来的（别人无法拥有 Alice 的私钥），并且文件内容没有 被修改过（摘要结果一致）

知名的数字签名算法包括 `DSA (Digital Signature Algorithm）`和安全强度更高的 `ECSDA（Elliptic Curve D耶tal Signature Algorith）`

从alice方可以看出，生成数字签名需要
```
签名的数据
私钥
```
从bob方可以看出,为了对一个签名进行验证
```
被签名的数据
签名
公钥
```
>数据签名并不是加密，你无法从一个签名重新构造出数据。这有点像哈希：你在数据上运行一个哈希算法，然后得到一个该数据的唯一表示。签名与哈希的区别在于密钥对：有了密钥对，才有签名验证。但是密钥对也可以被用于加密数据：私钥用于加密，公钥用于解密数据。不过比特币并不使用加密算法。


## 交易上链的过程

起初，创世块里面包含了一个 coinbase 交易。在 coinbase 交易中，没有输入，所以也就不需要签名。coinbase 交易的输出包含了一个哈希过的公钥（使用的是 RIPEMD16(SHA256(PubKey)) 算法）

当一个人发送币时，就会创建一笔交易。这笔交易的输入会引用之前交易的输出。每个输入会存储一个公钥（没有被哈希）和整个交易的一个签名。

比特币网络中接收到交易的其他节点会对该交易进行验证。除了一些其他事情，他们还会检查：在一个输入中，公钥哈希与所引用的输出哈希相匹配（这保证了发送方只能花费属于自己的币）；签名是正确的（这保证了交易是由币的实际拥有者所创建）。

当一个矿工准备挖一个新块时，他会将交易放到块中，然后开始挖矿。

当新块被挖出来以后，网络中的所有其他节点会接收到一条消息，告诉其他人这个块已经被挖出并被加入到区块链。

当一个块被加入到区块链以后，交易就算完成，它的输出就可以在新的交易中被引用。



## 实现地址

### 修改输入
```go
type TXInput struct {
	Txid      []byte
	Vout      int
	Signature []byte
	PubKey    []byte
}

func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := HashPubKey(in.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

```
现在我们已经不再需要 ScriptPubKey 和 ScriptSig 字段，因为我们不会实现一个脚本语言。相反，ScriptSig 会被分为 Signature 和 PubKey 字段，ScriptPubKey 被重命名为 PubKeyHash。我们会实现跟比特币里一样的输出锁定/解锁和输入签名逻辑，不同的是我们会通过方法（method）来实现

`UserKey`这个函数用来判断使用指定密钥来解锁一个输出。通过是否相等来判断这笔输入自身带的pubkey能否解锁一个输出字段的锁。
输入的字段信息是转账人提供的。

## 修改输出
```go
type TXOutput struct {
	Value      int
	PubKeyHash []byte
}

func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}
```
`lock`这个函数用于锁定一个输出，当我们给某个人发送币时，我们只知道他的地址，这个函数就是通过地址解码获得公钥，并保存在 PubKeyHash 字段
用于锁定输出。

`IsLockedWithKey`用于检查是否提供的公钥哈希被用于锁定输出。简单来说就是要验证一个输出是被哪个数据加锁。


其他的和输入输出相关的地方都要进行修改
```go
func NewCoinbaseTX(to, data string) *Transaction {
    ...
	txin := TXInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXOutput(subsidy, to)
	...
}
```

由于本章的代码量过大，而且很多函数不是很好理解。一个区块啦无非就是要做这么几件事。
```
# 创建账户，这里账户就是合法的比特币地址
# 根据比特币地址创建区块链，每个区块的产生都包含一笔币基交易，创世区块就只有一个交易，因为其他的地址上都没有币，那么需要一个地址获
# 转账操作
# 列出所有的账户
# 打印区块的所有信息
```
其他的操作我们会在后面进行补充

## 创建账户

### 椭圆曲线加密算法(ECDSA)

在比特币中，私钥就是一个256位的二进制数据，就像我抛硬币，只要我们足够随机，256位的0和1的组合就是一个可以用的私钥
```
随机产生私钥
私钥经过椭圆曲线加密算法生成公钥，
公钥经过一些列算法生成地址(不可读)
地址经过base58编码生成真实的比特币地址
```
我们这里不再讨论`ECDSA`的细节，有兴趣的话可以去网上查看相关的资料，比如[比特币背后的数学](https://mp.weixin.qq.com/s/MigSmfzRfyw-ux1lm0iBhw)。

### 比特币地址

比特币地址是完全公开的，如果你想要给某个人发送币，只需要知道他的地址就可以了。但是，地址（尽管地址也是独一无二的）并不是用来证明你是一个“钱包”所有者的信物。

实际上，所谓的地址，只不过是将公钥表示成人类可读的形式而已，因为原生的公钥人类很难阅读。

在比特币中，你的身份（identity）就是一对（或者多对）保存在你的电脑（或者你能够获取到的地方）上的公钥（public key）和私钥（private key）。比特币基于一些加密算法的组合来创建这些密钥，并且保证了在这个世界上没有其他人能够取走你的币，除非拿到你的密钥。下面，让我们来讨论一下这些算法到底是什么

### 代码
 1ApMbeKMBCXq8N6VZPx5HLhwdgcRa1UckL 
 18EaaxZCbaJwmHHW8zdbdsC9SwEJwMuNbs
首先cli监听到创建账户的命令，执行命令
```
func (cli *CLI) createWallet() {
	wallets, _ := NewWallets()
	address := wallets.CreateWallet()
	wallets.SaveToFile()

	fmt.Printf("Your new address: %s\n", address)
}
```
加载本地地址文件中已经保存的地址数据，创建新的地址，持久化到文件中。

钱包的结构
```go
type Wallets struct {
	Wallets map[string]*Wallet
}
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}
```
创建地址的过程其实还是很简单的的
```
func NewWallet() *Wallet {
	//通过椭圆曲线算法生成密钥对
	private, public := newKeyPair()
	wallet := Wallet{private, public}
	return &wallet
}
```
`address := wallets.CreateWallet()`通过公钥生成比特币地址
将一个公钥转换成一个 Base58 地址需要以下步骤：
```shell
使用 RIPEMD160(SHA256(PubKey)) 哈希算法，取公钥并对其哈希两次

给哈希加上地址生成算法版本的前缀

对于第二步生成的结果，使用 SHA256(SHA256(payload)) 再哈希，计算校验和。校验和是结果哈希的前四个字节。

将校验和附加到 version+PubKeyHash 的组合中。

使用 Base58 对 version+PubKeyHash+checksum 组合进行编码
```

至此，就可以得到一个真实的比特币地址。不过我可以负责任地说，无论生成一个新的地址多少次，检查它的余额都是 0。这就是为什么选择一个合适的公钥加密算法是如此重要：考虑到私钥是随机数，生成同一个数字的概率必须是尽可能地低。理想情况下，必须是低到“永远”不会重复。

另外，注意：你并不需要连接到一个比特币节点来获得一个地址。地址生成算法使用的多种开源算法可以通过很多编程语言和库实现
下面我们来看具体的函数
```go
func (w Wallet) GetAddress() []byte {
	//将公钥经过两次hash生成一个160位的二进制数
	pubKeyHash := HashPubKey(w.PublicKey)
	添加版本前缀
	versionedPayload := append([]byte{version}, pubKeyHash...)
	//获取校验和
	checksum := checksum(versionedPayload)
	//拼接
	fullPayload := append(versionedPayload, checksum...)
	//经过编码就生成地址
	address := Base58Encode(fullPayload)

	return address
}
```
### base58编码

现在，我们已经知道了这是公钥用人类可读的形式表示而已。如果我们对它进行解码，就会看到公钥的本来面目（16 进制表示的字节）：
```
0062E907B15CBF27D5425399EBF6F0FB50EBB88F18C29B7D93
```
160位的二进制也就是40位的16进制数。

比特币使用 Base58 算法将公钥转换成人类可读的形式。这个算法跟著名的 Base64 很类似，区别在于它使用了更短的字母表：为了避免一些利用字母相似性的攻击，从字母表中移除了一些字母。也就是，没有这些符号：0(零)，O(大写的 o)，I(大写的i)，l(小写的 L)，因为这几个字母看着很像。另外，也没有 + 和 / 符号。

去除这些字符其实也是为了防止地址相近似攻击，比如我要给三个 `000` 转账，结果别人改成`0O0` 中间是个字母o,但是你却以为地址还是原地址，很容易让别人达到攻击效果。

由于哈希函数是单向的（也就说无法逆转回去），所以不可能从一个哈希中提取公钥。不过通过执行哈希函数并进行哈希比较，我们可以检查一个公钥是否被用于哈希的生成。

```go
var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

func Base58Encode(input []byte) []byte {
	//获取进过版本前缀+哈希后的公钥+校验和三者拼接的字符串
	var result []byte
	//将输入转化为一个大整数
	x := big.NewInt(0).SetBytes(input)

	//创建整数58
	base := big.NewInt(int64(len(b58Alphabet)))
	//创建整数0
	zero := big.NewInt(0)
	mod := &big.Int{}

	// X>zero 返回1，x == zero 返回0
	for x.Cmp(zero) != 0 {
		// x除以base 后的商重新赋给x,余数赋给mod,自然mod是 0到57中间的一个数
		x.DivMod(x, base, mod)
		result = append(result, b58Alphabet[mod.Int64()])
	}

	if input[0] == 0x00 {
		result = append(result, b58Alphabet[0])
	}
	//是一个反转的算法，也比较简单
	ReverseBytes(result)

	return result
}
```

## 转账


## 签名
交易必须被签名，因为这是比特币里面保证发送方不会花费属于其他人的币的唯一方式。如果一个签名是无效的，那么这笔交易就会被认为是无效的，因此，这笔交易也就无法被加到区块链中。

一笔交易的哪些部分需要签名？用于签名的这个数据，必须要包含能够唯一识别数据的信息。比如，如果仅仅对输出值进行签名并没有什么意义，因为签名不会考虑发送方和接收方

那么必须要签名以下数据：
```shell
存储在已解锁输出的公钥哈希。它识别了一笔交易的“发送方”

存储在新的锁定输出里面的公钥哈希。它识别了一笔交易的“接收方”。

转账金额
```
这三个数据都包含在交易里面，我们需要用私钥进行签名
```go
//根据私钥和交易生成数字签名
func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		//因为我们每一笔交易都来自之前交易的输出
		//这里需要确认引入的交易id是否之前发生过
		prevTX, err := bc.FindTransaction(vin.Txid)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

////这里需要确认引入的交易id是否之前发生过
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()
		//遍历每个区块的每一笔交易判断交易是否发生过
		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction is not found")


```

代码实现：
```go
//签名肯定要用私钥进行签名，那么才能
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	//每个区块的奖励交易不需要签名，因为没有发送方
	if tx.IsCoinbase() {
		return
	}
	//组建交易的数据，包含了所有的输入和输出，但是 TXInput.Signature 和 TXIput.PubKey 被设置为 nil。
	txCopy := tx.TrimmedCopy()
	//迭代每一笔交易的输入
	for inID, vin := range txCopy.Vin {
		//
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		//txCopy.Vin[inID]代表的是哪一笔输入，因为一个交易的输入可能由很多笔UTXO组成
		txCopy.Vin[inID].Signature = nil
		//指定每一笔交易所引用的每一笔输入的PubKey = 该笔输入(之前交易的输出)的PubKeyHash，这样保证花费是自己账户的钱
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		//提取交易的摘要，其实就是将数据进行hash计算
		txCopy.ID = txCopy.Hash()
		//在获取完哈希，我们应该重置 PubKey 字段，以便于它不会影响后面的迭代
		txCopy.Vin[inID].PubKey = nil

		//一个随机数，私钥，交易的摘要数据生成数字签名。一个签名就是一对数字，一个公钥就是一对坐标。
		//我们之前为了存储将它们连接在一起，现在我们需要对它们进行解包在 crypto/ecdsa 函数中使用
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Vin[inID].Signature = signature
	}
}
```
节点在接受到交易在添加到交易内存池之前，首先也需要进行交易验证。

给你一笔交易，验证是否合法
```go
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	//string存储交易的id编码后的id,Transaction存储的是该交易id所代表的交易
	//相当于是用另外一种方式map结构存储交易数据
	prevTXs := make(map[string]Transaction)
 
	for _, vin := range tx.Vin {
		//FindTransaction(vin.Txid)找到返回的就是这个交易id的交易
		//找不到返回为空
		prevTX, err := bc.FindTransaction(vin.Txid)
		//
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inID, vin := range tx.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil
		//解码`TXInput.Signature`获取数字签名,一个签名就是一对数字
		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])
		
		//解码`TXInput.PubKey`,一个公钥就是一对坐标
		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keyLen / 2)])
		y.SetBytes(vin.PubKey[(keyLen / 2):])
		//
		rawPubKey := ecdsa.PublicKey{curve, &x, &y}
		//根据公钥、摘要、数字签名就可以验证
		//所有的输入都被验证，返回 true；如果有任何一个验证失败，返回 false
		if ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false {
			return false
		}
	}

	return true
}

```
通过上述实现我们也明白了，交易被打包到区块前需要去验证，那么修改工作量证明中的区块
```go
func (bc *Blockchain) AddBlock(transactions []*Transaction) {
	var lastHash []byte

	for _, tx := range transactions {
		if bc.VerifyTransaction(tx) != true {
			log.Panic("ERROR: Invalid transaction")
		}
	}
	...
}
```
现在创建一笔新的交易也必须进行签名了
```go
func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	...
	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	bc.SignTransaction(&tx, wallet.PrivateKey)

	return &tx
}
```
如果之前的整个流程你都很熟悉的话，修改cli是很简单的事情，由于目前我们的命令过多，文件过大，我们可以将cli中每个操作
单独放在一个文件中。


