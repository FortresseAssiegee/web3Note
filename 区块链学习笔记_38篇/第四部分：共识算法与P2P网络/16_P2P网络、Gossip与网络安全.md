# P2P 网络、Gossip 与网络安全

## 本篇目录
- [一、本篇学习目标](#一本篇学习目标)
- [二、参考答案与解题要点](#二参考答案与解题要点)
- [三、代码文件与复现记录](#三代码文件与复现记录)
- [四、与密码学教材的阅读衔接](#四与密码学教材的阅读衔接)
- [五、参考资料与实际查阅范围](#五参考资料与实际查阅范围)
  - [5.1 基础模型与 Bitcoin](#51-基础模型与-bitcoin)
  - [5.2 Ethereum](#52-ethereum)
  - [5.3 DPoS、BFT 与 Fabric](#53-dposbft-与-fabric)
  - [5.4 P2P、Go API 与教材](#54-p2pgo-api-与教材)
- [六、P2P 网络的基本结构](#六p2p-网络的基本结构)
- [七、Flooding 式传播](#七flooding-式传播)
- [八、冲突交易的 Race Condition](#八冲突交易的-race-condition)
- [九、区块传播为何影响共识](#九区块传播为何影响共识)
- [十、网络传播不是可靠广播](#十网络传播不是可靠广播)
- [十一、链上支付与线下履约](#十一链上支付与线下履约)
- [十二、本讲小结](#十二本讲小结)
- [十三、本篇与知识体系的关系](#十三本篇与知识体系的关系)
- [一图看懂本篇流程](#一图看懂本篇流程)
- [官方实现与代码阅读入口](#官方实现与代码阅读入口)


## 一图看懂本篇流程


Gossip 像朋友之间转发通知：收到新消息后告诉自己的联系人，联系人再继续转发。消息不一定同时到达每个人。

> **怎么读这张图：**从左向右看，先确认输入是什么，再看系统如何检查和处理，最后看结果如何保存或验证。图中是帮助理解的简化流程，协议细节仍以本篇后续说明和官方规范为准。


例如，社区通知通过住户之间转发：收到通知的人告诉邻居，邻居再转给其他人。为了避免群聊刷屏，要识别重复消息并限制转发。

## 官方实现与代码阅读入口

以下链接指向对应技术的官方文档、标准或项目源码。阅读代码时，建议先从本篇涉及的核心数据结构或验证入口开始，再沿调用关系查看后续处理。

- [Bitcoin P2P 参考](https://developer.bitcoin.org/reference/p2p_networking.html)
- [Go Ethereum P2P](https://github.com/ethereum/go-ethereum/tree/master/p2p)

**代码来源说明：**官方仓库中的实现代码才代表相应项目的真实实现。正文中用于解释概念的短代码或实验片段，如未明确标出官方文件路径，应按教学演示理解，不能直接当作 Bitcoin Core、Go Ethereum 或其他客户端的生产代码。实际开发应对照上方官方版本、规范和测试。


## 一、本篇学习目标
- P2P 网络、Gossip 与网络安全
- 理解核心概念、运行流程和参与角色
- 能够区分协议保证、实现细节与应用层假设
- 能够结合代码或实际场景分析边界与风险

## 二、参考答案与解题要点

1. 两份签名都可能是真实授权，但二者争用同一份状态。节点需要共同顺序和相同有效性规则，否则会形成相互矛盾的余额历史。密码学认证不负责全局排序。
2. $105-100+1=6$。它只说明该观察者当前视图中的深度；隔离节点可能不知道更高工作量分支或最新最终性状态，所以不能把局部数字当作全网事实。
3. 目标包括 0 到 7，共 8 个输出，成功率 $8/256=1/32$，期望 32 次。期望不保证第 32 次必定成功。
4. A 为 $4\times3=12$，B 为 $3\times5=15$，选择 B。前提是两条分支都满足实际协议全部有效性规则。
5. 真实目标由历史与难度调整规则约束。教学程序故意允许合成目标，用来隔离累计工作量算术，没有实现真实链的上下文难度验证。
6. 常用 $q=2f+1=5$，交集至少 $5+5-7=3$，其中至少一人诚实，因为坏副本最多 2。
7. 两个三人集合最少只重合一人，这个人可能就是唯一坏副本。需要满足 $2q-n>f$，例如 q=4；不能机械套用与最紧配置绑定的阈值。
8. 需要 101。100 只是恰好三分之二。整数算术可避免浮点精度和边界错误；实现还需避免计算两倍总权重时溢出。
9. 满足，因为 $3\times64=2\times96$。但还必须是正确 epoch、source、target、成员与状态条件，最终化也需要相应状态机规则，不能直接把数值条件扩大为完整结论。
10. 32 ETH 是最小激活余额；2048 ETH 是符合复利凭证等相应规则的单个验证者最大有效余额上限，两者不是同一种门槛。
11. slot 是计划时间位置，可能因为无有效及时提案而为空；时间推进和成功产块数量不同。
12. 不是。收据代币代表的资产权益、赎回权或收益关系由具体应用规定，底层投票仍由相应验证者签名密钥按职责执行。
13. 根节点处比较子树权重，A 得 65、B 得 35，所以先走 A；在 A 内再比较 A1 的 40 与 A2 的 25。不能跨层只挑直接得票最多的任意节点。
14. prepared 需要请求、匹配 pre-prepare 和不同备份的 2f 份 prepare；committed-local 还要求本地 prepared 和 2f+1 份匹配 commit。消息类型、视图、序号及摘要都不能混用。
15. 不能。轮次是签名语义的一部分，不同轮的票不是同一提交证据。相同高度仍不足以让它们相加。
16. 锁承载先前证据产生的安全约束。无条件丢锁可能让新领导覆盖旧承诺；只有满足协议规定的较新或本轮相应证据条件，才能按状态机改变支持。
17. 排序解决先后，提交验证还检查背书和读写版本等条件。无效交易可在区块中保留并标记无效，其写集不更新世界状态。
18. 去重防止同一成员重复计权；本例还禁止同一公钥以多个注册身份重复出现，以简化身份绑定。不同密钥仍可能由同一个人控制，这些检查不能证明组织独立性。
19. TCP 暴露字节流，读取可能拆分或合并应用写入。需要长度、分隔符或其他明确 framing；本例使用受限长度字段和 ReadFull。
20. TTL 每跳变化，若进入 ID 会使同一消息不断变成新标识，破坏去重。哈希正确只证明内容与标识一致，授权还需签名和业务规则。
21. $100\oplus101=001=1$，$111\oplus101=010=2$，所以 100 更近。XOR 距离属于逻辑空间，不代表物理网络延迟。
22. 不能。它说明对端控制相应密钥，并在握手的信任条件下与预期身份关联；攻击者可以合法拥有很多密钥并完成握手。
23. 每秒净增加约 4000 条待处理消息，内存持续增长。必须通过限流、背压、预算和适当优先级控制输入，不能无限排队。
24. 旧密钥可以在更晚时刻签出另一个历史；签名一般不证明真实产生时间。节点需要可信的成员变化链和足够新的弱主观性起点等协议条件。
25. 它只验证签名、域、集合、去重和权重，缺少锁、轮次状态机、持久化防双签、网络恢复、区块执行与成员轮换。一个子检查通过不能替代整个协议。
26. 概率为 $2^{-10}=1/1024$。现实还需证明抽样方式、独立性、编码、承诺、对手响应和网络条件，简单概率不涵盖这些前提。

<a id="sec-085"></a>

## 三、代码文件与复现记录

四个程序只依赖 Go 标准库，不需要外部密码库或公链账号。请从本套笔记对应代码包的根目录运行，不要把四个 `package main` 文件合并到一个目录或一个源文件。

| 文件 | 检查的核心性质 | 已观察到的结果 |
|---|---|---|
| `code/ch05/pow_work/main.go` | 合成工作量、链式哈希检查、内容改动检测 | 累计 work 6 与 32，选择 B，改动被拒绝 |
| `code/ch05/weighted_qc/main.go` | 权重门槛、签名域、去重、集合摘要 | 70 权重通过，低权重/重复/改上下文均拒绝 |
| `code/ch05/framing/main.go` | 短读、连帧、长度上限、截断处理 | 正确解出 2 帧，超长与截断被拒绝 |
| `code/ch05/gossip/main.go` | 有限图去重交付、TTL、ID 一致性 | 5 节点各交付一次，重复抑制，TTL=0 不转发 |

复现环境为 **Go 1.27.3**，四次 `go run` 均正常退出。实现使用 `crypto/sha256`、`crypto/ed25519`、`crypto/rand`、`encoding/binary` 和 `io` 等标准库；API 语义可对照官方文档。[S34] [S35] [S36]

本讲没有执行真实节点网络测试、形式化验证、安全审计、故障注入集群测试或性能基准。代码中的拒绝案例针对文中明确展示的风险，是学习材料的功能验证，不是对任何生产系统安全性的认证。

<a id="sec-086"></a>

## 四、与密码学教材的阅读衔接

Katz 与 Lindell 的 *Introduction to Modern Cryptography* 可用于建立安全定义、攻击者能力、消息认证与签名等基础。读本讲时，可以把“验证一份签名”放回不可伪造性模型，把“哈希承诺”放回抗碰撞与抗原像等性质，再问协议额外需要哪些网络和状态假设。作者公开页面提供第三版介绍及目录/前言入口；本次核查的是这些公开信息，不宣称读取了整本教材。[S37]

《深入浅出密码学》相关英文教材 *Understanding Cryptography* 的公开出版页可用于按哈希、数字签名、MAC 和密钥管理等主题安排工程复习。所链接的 2024 年版本作者为 Christof Paar、Jan Pelzl 和 Tim Güneysu，公开页面展示章节目录与出版信息。本讲依据这些公开资料建立阅读衔接，没有复制教材正文或把未读章节冒称为已阅读全文。[S38]

“现代密码学”也可能是多本中文教材的名称，若未指定作者与版次，不能假定它唯一对应某一本书。对这一讲而言，密码学教材主要解释底层工具；共识阈值、锁、客户端分工和具体网络编码仍应回到论文与协议规范，不能从通用教材推断某条链当前的实现细节。

<a id="sec-087"></a>

## 五、参考资料与实际查阅范围

以下优先列官方规范、项目官方文档、原始论文及教材作者/出版社页面。对历史资料保留版本含义；对随开发分支变化的页面，核查范围以本讲标注日期为准。代码和数字案例为本讲自行编写，正文没有大段翻译或转载来源文章。

<a id="sec-088"></a>
### 5.1 基础模型与 Bitcoin

- **S3**：[Bitcoin: A Peer-to-Peer Electronic Cash System](https://bitcoin.org/bitcoin.pdf)。用于核对数字签名之外的双花问题、PoW、网络与确认的基本背景。
- **S4**：[Fischer、Lynch、Paterson：Impossibility of Distributed Consensus with One Faulty Process](https://groups.csail.mit.edu/tds/papers/Lynch/jacm85.pdf)。核对完全异步和故障模型下的终止限制。
- **S5**：[Bitcoin Core：pow.cpp](https://github.com/bitcoin/bitcoin/blob/master/src/pow.cpp)。核对目标验证与难度调整的实现逻辑。
- **S8**：[Bitcoin Core：kernel/chainparams.cpp](https://github.com/bitcoin/bitcoin/blob/master/src/kernel/chainparams.cpp)。核对主网目标间隔与目标周期，并注意不同网络参数区别。
- **S9**：[Bitcoin Core：chain.cpp](https://github.com/bitcoin/bitcoin/blob/master/src/chain.cpp)。核对按目标计算整数 work 的实现。

<a id="sec-089"></a>
### 5.2 Ethereum

- **S1**：[Node architecture](https://ethereum.org/developers/docs/nodes-and-clients/node-architecture/)。核对执行客户端、共识客户端、验证者与 Engine API 分工。
- **S6**：[Consensus Specs：Phase0 Beacon Chain](https://ethereum.github.io/consensus-specs/specs/phase0/beacon-chain/)。查阅相关状态容器、签名/违规谓词与 justification/finalization 计算；后续分叉需结合增量规范。
- **S7**：[Weak subjectivity](https://ethereum.org/developers/docs/consensus-mechanisms/pos/weak-subjectivity/)。核对长程替代历史与可信检查点的关系。
- **S10**：[The great renaming: what happened to Eth2?](https://blog.ethereum.org/2022/01/24/the-great-eth2-renaming)。核对 Eth1/Eth2 术语退役和执行层/共识层称呼。
- **S11**：[The Merge](https://ethereum.org/roadmap/merge/)。核对 2022 年 9 月 15 日完成共识转换的历史时间。
- **S12**：[Prague-Electra (Pectra)](https://ethereum.org/roadmap/pectra/)。核对官方拼写、层次组合及实际主网激活时间。
- **S13**：[EIP-7251](https://eips.ethereum.org/EIPS/eip-7251)。查阅 32 ETH、2048 ETH、复利凭证与合并相关设计。
- **S14**：[Consensus Specs：Electra Beacon Chain](https://ethereum.github.io/consensus-specs/specs/electra/beacon-chain/)。查阅 `get_max_effective_balance` 等与凭证类型有关的规则。
- **S15**：[Liquid & pooled staking](https://ethereum.org/staking/pools/)。核对外部池化服务和原生验证者协议的区别。
- **S16**：[Gasper](https://ethereum.org/developers/docs/consensus-mechanisms/pos/gasper/)。核对 LMD-GHOST 与 Casper FFG 的分工。
- **S17**：[Consensus Specs：Fork Choice](https://ethereum.github.io/consensus-specs/specs/phase0/fork-choice/)。查阅可行分支筛选、`get_head` 与最新消息权重相关结构。
- **S18**：[Block proposal](https://ethereum.org/developers/docs/consensus-mechanisms/pos/block-proposal/)。核对 slot、epoch 与提议职责的基本流程。
- **S40**：[Engine API：common specification](https://github.com/ethereum/execution-apis/blob/main/src/engine/common.md)。查阅客户端间接口约定与版本化背景；具体方法需继续读取相应分叉文件。
- **S41**：[Attestations](https://ethereum.org/developers/docs/consensus-mechanisms/pos/attestations/)。核对 head/source/target 的语义及聚合传播职责；容器字段按实际分叉规范校对。
- **S45**：[Networking layer](https://ethereum.org/developers/docs/networking-layer/)。查阅共识网络发现、gossip 与 request-response 分工。

<a id="sec-090"></a>
### 5.3 DPoS、BFT 与 Fabric

- **S20**：[EOSIO Technical White Paper](https://github.com/EOSIO/Documentation/blob/master/TechnicalWhitePaper.md)。仅用于历史 DPoS 代表选择机制背景，不采用其中营销性优越性表述，也不推断其当前网络实现。
- **S21**：[Castro、Liskov：Practical Byzantine Fault Tolerance](https://css.csail.mit.edu/6.824/2014/papers/castro-practicalbft.pdf)。查阅正常路径、prepared/committed-local、view change 和认证优化；历史密码参数不作为今天的配置建议。
- **S22**：[CometBFT 官方共识状态机说明](https://github.com/cometbft/cometbft/blob/main/spec/consensus/consensus.md)。查阅高度、轮次、PoLC、锁、严格阈值和提交步骤。
- **S23**：[Buchman、Kwon、Milosevic：The latest gossip on BFT consensus](https://arxiv.org/pdf/1807.04938)。查阅 Tendermint 的变量、轮次算法及安全/活性建模，不把论文算法与任意客户端版本逐行等同。
- **S24**：[Hyperledger Fabric v0.6：Consensus Algorithm FAQ](https://hlf.readthedocs.io/en/v0.6/FAQ/consensus_FAQ.html)。核对该历史版本的 No-op 与 Batch PBFT 选项。
- **S2**：[Hyperledger Fabric：Transaction Flow](https://hyperledger-fabric.readthedocs.io/en/latest/txflow.html)。查阅背书模拟、排序、验证和提交职责。
- **S25**：[Hyperledger Fabric：The Ordering Service](https://hyperledger-fabric.readthedocs.io/en/latest/orderer/ordering_service.html)。查阅当前文档中的 Raft、BFT/SmartBFT 及故障模型区别。
- **S26**：[HotStuff: BFT Consensus in the Lens of Blockchain](https://arxiv.org/abs/1803.05069)。查阅论文摘要和版本信息，用于说明部分同步、响应性和通信复杂度的研究方向。
- **S27**：[Ongaro、Ousterhout：In Search of an Understandable Consensus Algorithm](https://raft.github.io/raft.pdf)。查阅复制状态机、领导选举与日志复制的组织方式，作为 CFT 对照。

<a id="sec-091"></a>
### 5.4 P2P、Go API 与教材

- **S28**：[libp2p：Peers](https://libp2p.io/docs/peers/)。核对 Peer ID 与公钥身份的关系。
- **S29**：[libp2p：Secure Channels](https://libp2p.io/docs/secure-channels-overview/)。查阅 TLS、Noise 与传输层保护的职责。
- **S30**：[libp2p：Kademlia DHT](https://libp2p.io/docs/kademlia-dht/)。查阅 XOR 距离、路由表和查询思路。
- **S31**：[libp2p：Publish/Subscribe](https://libp2p.io/docs/pubsub/)。查阅 mesh、元数据传播及 IHAVE/IWANT；课堂程序未实现完整 GossipSub。
- **S32**：[libp2p：Discovery & Routing](https://libp2p.io/docs/discovery-routing-overview/)。查阅发现与定位的职责区别。
- **S33**：[libp2p：DoS Mitigation](https://libp2p.io/docs/dos-mitigation/)。查阅资源放大风险与连接、资源限制思路。
- **S34**：[Go：crypto/ed25519](https://pkg.go.dev/crypto/ed25519)。核对密钥生成、签名、验签和公钥长度要求。
- **S35**：[Go：io.ReadFull](https://pkg.go.dev/io#ReadFull)。核对短读、EOF 与 ErrUnexpectedEOF 语义。
- **S36**：[Go：crypto/sha256](https://pkg.go.dev/crypto/sha256)。核对摘要 API；程序采用标准库，不手写哈希压缩函数。
- **S37**：[Katz、Lindell 教材作者页](https://www.cs.umd.edu/~jkatz/imc.html)。查阅第三版介绍及目录/前言入口。
- **S38**：[Understanding Cryptography 出版页](https://link.springer.com/book/10.1007/978-3-662-69007-9)。查阅公开目录、作者和出版信息。
- **S42**：[RFC 9293：TCP](https://www.rfc-editor.org/rfc/rfc9293.html)。核对有序可靠字节流语义。
- **S43**：[libp2p：Hole Punching](https://libp2p.io/docs/hole-punching/)。查阅 NAT 下连接建立与协调思路。
- **S44**：[libp2p Peer IDs 规范](https://github.com/libp2p/specs/blob/master/peer-ids/peer-ids.md)。核对 identity multihash、SHA-256 multihash 和文本表示，避免过度简化公钥标识。
- **S46**：[libp2p：Circuit Relay](https://libp2p.io/docs/circuit-relay/)。查阅中继通信和端到端通道保护的关系。

[S1]: https://ethereum.org/developers/docs/nodes-and-clients/node-architecture/
[S2]: https://hyperledger-fabric.readthedocs.io/en/latest/txflow.html
[S3]: https://bitcoin.org/bitcoin.pdf
[S4]: https://groups.csail.mit.edu/tds/papers/Lynch/jacm85.pdf
[S5]: https://github.com/bitcoin/bitcoin/blob/master/src/pow.cpp
[S6]: https://ethereum.github.io/consensus-specs/specs/phase0/beacon-chain/
[S7]: https://ethereum.org/developers/docs/consensus-mechanisms/pos/weak-subjectivity/
[S8]: https://github.com/bitcoin/bitcoin/blob/master/src/kernel/chainparams.cpp
[S9]: https://github.com/bitcoin/bitcoin/blob/master/src/chain.cpp
[S10]: https://blog.ethereum.org/2022/01/24/the-great-eth2-renaming
[S11]: https://ethereum.org/roadmap/merge/
[S12]: https://ethereum.org/roadmap/pectra/
[S13]: https://eips.ethereum.org/EIPS/eip-7251
[S14]: https://ethereum.github.io/consensus-specs/specs/electra/beacon-chain/
[S15]: https://ethereum.org/staking/pools/
[S16]: https://ethereum.org/developers/docs/consensus-mechanisms/pos/gasper/
[S17]: https://ethereum.github.io/consensus-specs/specs/phase0/fork-choice/
[S18]: https://ethereum.org/developers/docs/consensus-mechanisms/pos/block-proposal/
[S20]: https://github.com/EOSIO/Documentation/blob/master/TechnicalWhitePaper.md
[S21]: https://css.csail.mit.edu/6.824/2014/papers/castro-practicalbft.pdf
[S22]: https://github.com/cometbft/cometbft/blob/main/spec/consensus/consensus.md
[S23]: https://arxiv.org/pdf/1807.04938
[S24]: https://hlf.readthedocs.io/en/v0.6/FAQ/consensus_FAQ.html
[S25]: https://hyperledger-fabric.readthedocs.io/en/latest/orderer/ordering_service.html
[S26]: https://arxiv.org/abs/1803.05069
[S27]: https://raft.github.io/raft.pdf
[S28]: https://libp2p.io/docs/peers/
[S29]: https://libp2p.io/docs/secure-channels-overview/
[S30]: https://libp2p.io/docs/kademlia-dht/
[S31]: https://libp2p.io/docs/pubsub/
[S32]: https://libp2p.io/docs/discovery-routing-overview/
[S33]: https://libp2p.io/docs/dos-mitigation/
[S34]: https://pkg.go.dev/crypto/ed25519
[S35]: https://pkg.go.dev/io#ReadFull
[S36]: https://pkg.go.dev/crypto/sha256
[S37]: https://www.cs.umd.edu/~jkatz/imc.html
[S38]: https://link.springer.com/book/10.1007/978-3-662-69007-9
[S40]: https://github.com/ethereum/execution-apis/blob/main/src/engine/common.md
[S41]: https://ethereum.org/developers/docs/consensus-mechanisms/pos/attestations/
[S42]: https://www.rfc-editor.org/rfc/rfc9293.html
[S43]: https://libp2p.io/docs/hole-punching/
[S44]: https://github.com/libp2p/specs/blob/master/peer-ids/peer-ids.md
[S45]: https://ethereum.org/developers/docs/networking-layer/
[S46]: https://libp2p.io/docs/circuit-relay/


## 六、P2P 网络的基本结构

课程中的比特币网络由对等节点构成。节点维护一组邻居，并通过网络消息发现其他节点、交换交易和区块信息。加入网络时，新节点需要先找到可联系的节点，再逐步发现更多对等节点。

网络拓扑与消息转发规则影响传播速度。节点不一定直接连接所有其他节点，也不保证每条消息以相同顺序抵达所有节点。

## 七、Flooding 式传播

课程描述的传播方式可概括为：

1.  节点首次收到某条交易或区块消息。
2.  节点先执行必要的基本检查，并记录已经见过该消息。
3.  节点把新消息转发给邻居。
4.  若再次收到相同消息，节点利用记录避免重复转发。

这种方式实现简单、具有一定鲁棒性，但会带来重复消息、带宽消耗和传播延迟。实际实现可以使用消息标识、库存消息等机制减少重复传输。

## 八、冲突交易的 Race Condition

假设 Alice 使用同一个 UTXO 创建两笔冲突交易：

- $T_1$：支付给 Bob；
- $T_2$：支付给 Carol。

由于网络延迟，节点 A 可能先收到 $T_1$，节点 B 可能先收到
$T_2$。它们本地的待处理交易集合（mempool）就可能暂时不同。

这并不意味着两笔交易都能在同一条有效链上被接受。节点会根据当前状态检查输入是否已被花费；区块被接受后，相关冲突交易需要从待处理集合中移除或被视为无效。

## 九、区块传播为何影响共识

新区块传播期间，部分矿工可能尚未收到它，仍在旧区块后继续挖矿。若另一个区块也在这段时间内产生，就可能出现暂时分叉。

区块越大，通常需要传输和验证的数据越多，传播到网络各处所需时间可能越长。区块传播延迟与出块间隔共同影响分叉概率。

## 十、网络传播不是可靠广播

P2P 网络中的消息传播可能遇到：

- 节点离线或连接中断；
- 消息延迟或丢失；
- 不同节点收到消息的顺序不同；
- 恶意节点拒绝转发合法消息，或传播无效消息。

因此，协议不能假设所有节点在同一瞬间看到完全相同的消息集合。节点必须独立验证数据，并由共识规则处理因传播延迟而形成的暂时分歧。

## 十一、链上支付与线下履约

区块链能够记录支付状态，却无法单独保证收款方一定发货、商品一定合格或线下服务一定完成。这些事项属于链外履约问题，需要额外的合同、平台、仲裁或其他治理机制处理。

## 十二、本讲小结

P2P
网络负责传播消息，不负责让所有节点瞬间同步。消息传播顺序不同会导致暂时的交易池差异和区块分叉；节点验证与共识机制共同决定最终被接受的账本历史。

[课程视频：P6
比特币网络](https://www.bilibili.com/video/BV1Vt411X7JF/?p=6) ·


## 十三、本篇与知识体系的关系

- 所属部分：第四部分：共识算法与P2P网络
- 返回总目录：[README](../README.md)
