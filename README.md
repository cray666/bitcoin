作为区块链思想诞生的源头，比特币网络是首个得到大规模部署的区块链技术应用，并且是首个得到实践检验的数字货币实现。


## 原理和设计

比特币网络是一个分布式的点对点网络，网络中的矿工通过“挖矿”来完成对交易记 录的记账过程，维护网络的正常运行。 

区块链网络提供一个公共可见的记账本，该记账本并非记录每个账户的余额，而是用 来记录发生过的交易的历史信息。

比特币首次真正从实践意义上实现了安全可靠的去中心化数字货币机制 ，这也是它受 到无数金融科技从业者热捧的根本原因

## 什么是比特币

比特币是一种去中心化的点对点的数字资产，由中本聪在2009挖出来第一个区块，奖励了50个比特币，

为什么比特币的总量是2100万个？

每个区块的奖励最初是 50 个比特币，每隔 21 万个区块自动减半，比特币根据全网络中算力进行调整，保证每10分钟出一个区块，计算下来也就是 4 年时间，而且每个比特币最多能细分到小数点后8位，也即是亿分之一(聪)，挖矿奖励每四年减半一次，按照这种计算方式，2045年 99.95%的比特币会发行完毕，2140年比特币无法细分，至此比特币完全发行完毕，总量稳定在 2100 万个。

比特币的这种发行机制激励这很多矿工尽早的投入到比特币的挖矿中。保证比特币网络的安全和可靠性。


## 谁是中本聪

除了比特币精妙的设计理念外， 比特币最为人津津乐道的一点是发明人“中本聪”到目前为 止尚无法确认真实身份。 也有人推测，“中本聪”背后可能不止一个人，而是一个团队。 这 些猜测都为比特币项目带来了不少传奇色彩。

## 比特币交易过程

当你发起一笔比特币交易后，你需要将交易广播至全网中，挖矿节点通过P2P接收到这笔交易后，先将其放入本地内存进行一些基本验证，比如该笔交易中交易输入是否是UTXO，如果验证通过，将其放入未确认交易池，等待被打包。如果验证失败，该笔交易就会标记为无效交易，不会被打包。

挖矿节点从未确认交易池中每次抽取一定量的交易进行打包，有时候我们的发出的交易不能被及时打包上链，是因为未确认交易池中得交易数量过多，而每个区块能记录的交易数量有限，比如最大不能超过1M，这时候就会造成区块拥堵。


## 对称加密和非对称加密

什么是对称加密和非对称加密？

```shell
加密过程中，通过加密算法和加密密钥，对明文进行加密，获得密文。 
解密过程中，通过解密算法和解密密钥，对密文进行解密，获得明文。
```
根据加解密过程中所使用的密钥是否相同，算法可以分为对称加密`(symmetric cryptography)`和非对称加密(Asymmetric encryption)。 两种模式适用于不同的需求，恰好形成互补。



