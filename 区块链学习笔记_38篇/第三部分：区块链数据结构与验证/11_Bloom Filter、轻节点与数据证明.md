# Bloom Filter、轻节点与数据证明

## 本篇目录
- [一、本篇学习目标](#一本篇学习目标)
- [二、附录 C：本讲代码验证记录](#二附录-c本讲代码验证记录)
- [三、附录 D：阅读资料与来源边界](#三附录-d阅读资料与来源边界)
  - [3.1 1 Merkle、MPT 与编码](#31-1-merklempt-与编码)
  - [3.2 2 共识层、区块与执行承诺](#32-2-共识层区块与执行承诺)
  - [3.3 3 比特币与 Bloom](#33-3-比特币与-bloom)
  - [3.4 4 可以继续深入的学习问题](#34-4-可以继续深入的学习问题)
  - [3.5 5 与现代密码学教材衔接](#35-5-与现代密码学教材衔接)
- [四、哈希指针](#四哈希指针)
- [五、区块如何连接成链](#五区块如何连接成链)
- [六、Merkle 树如何工作](#六merkle-树如何工作)
- [七、Merkle Proof：只证明一笔交易](#七merkle-proof只证明一笔交易)
  - [7.1 证明的边界](#71-证明的边界)
- [八、区块头与区块主体](#八区块头与区块主体)
- [九、为什么这两种结构重要](#九为什么这两种结构重要)
- [十、本讲小结](#十本讲小结)
- [十一、本篇与知识体系的关系](#十一本篇与知识体系的关系)
- [一图看懂本篇流程](#一图看懂本篇流程)
- [官方实现与代码阅读入口](#官方实现与代码阅读入口)


## 一图看懂本篇流程


Bloom Filter 像图书馆入口的快速筛查表：它可以告诉你某本书肯定不在某个区域，或可能在；说“可能在”时还要去书架核实。

> **怎么读这张图：**从左向右看，先确认输入是什么，再看系统如何检查和处理，最后看结果如何保存或验证。图中是帮助理解的简化流程，协议细节仍以本篇后续说明和官方规范为准。


例如，商场活动系统先用 Bloom Filter 快速筛掉肯定没有某优惠券的日期，再检查候选日期的完整订单记录，避免把“可能有”当成确实有。

## 官方实现与代码阅读入口

以下链接指向对应技术的官方文档、标准或项目源码。阅读代码时，建议先从本篇涉及的核心数据结构或验证入口开始，再沿调用关系查看后续处理。

- [Bitcoin Core 源码](https://github.com/bitcoin/bitcoin/tree/master/src)
- [Go Ethereum bloom bits](https://github.com/ethereum/go-ethereum/tree/master/core/bloombits)

**代码来源说明：**官方仓库中的实现代码才代表相应项目的真实实现。正文中用于解释概念的短代码或实验片段，如未明确标出官方文件路径，应按教学演示理解，不能直接当作 Bitcoin Core、Go Ethereum 或其他客户端的生产代码。实际开发应对照上方官方版本、规范和测试。


## 一、本篇学习目标
- Bloom Filter、轻节点与数据证明
- 理解核心概念、运行流程和参与角色
- 能够区分协议保证、实现细节与应用层假设
- 能够结合代码或实际场景分析边界与风险

## 二、附录 C：本讲代码验证记录

四个完整程序均在 Go 1.27.3 环境中实际执行，进程正常退出：

| 示例 | 核对结果 | 结论边界 |
|---|---|---|
| Merkle | 正确证明通过；改值、错索引、改兄弟摘要均拒绝 | 自定义教学树，不是比特币/Ethereum 链上兼容验证器 |
| HP | 四种奇偶/节点组合及空 leaf 往返一致；非法填充和 nibble 16 拒绝 | 编码示例，不包含完整 MPT 树形合法性检查 |
| Bloom | 参数得到 9586 bit、7 次位置映射；已插入元素命中；一位过滤器出现预设假阳性 | 不是 Ethereum logsBloom 的编码实现 |
| Trie | 更新改变根；删除前缀键保留长键；旧快照不受修改影响；空值可区分不存在 | 自定义 byte trie；全量内存快照，不含磁盘耐久性 |

实测说明这些具体示例与列出的行为一致，不等于经过生产审计、模糊测试或大规模性能验证。完整 Go 程序都已嵌入本文，复制到各自独立的 `main.go` 即可运行；不要把四个 `package main` 块拼进同一个文件。

<a id="sec-075"></a>

## 三、附录 D：阅读资料与来源边界

本笔记采用原创教学说明和小例子，技术细节以协议、标准和实现作者维护的资料为核对入口。网页或源码的 `master/latest` 链接可能更新；复现实务项目应锁定提交或版本，并记录目标网络的分叉规则。

<a id="sec-076"></a>
### 3.1 1 Merkle、MPT 与编码

1. [RFC 9162，§2.1 Merkle Trees](https://www.rfc-editor.org/rfc/rfc9162.html#section-2.1)：叶子/内部节点域分离、树形、包含与一致性证明。本文参考其结构思想编写自己的 Go 示例，不复刻完整 CT 协议。
2. [Ethereum：Merkle Patricia Trie](https://ethereum.org/en/developers/docs/data-structures-and-encoding/patricia-merkle-trie/)：节点种类、HP、Trie 的使用对象。
3. [Ethereum：RLP](https://ethereum.org/en/developers/docs/data-structures-and-encoding/rlp/)：确定性编码、空串与空列表、整数编码。
4. [Geth：trie/hasher.go](https://github.com/ethereum/go-ethereum/blob/master/trie/hasher.go)：小节点内联与强制根哈希。
5. [Geth：trie/trie.go](https://github.com/ethereum/go-ethereum/blob/master/trie/trie.go)：查找、更新、删除、根计算与 Commit 的职责。
6. [Geth：trie 包文档](https://pkg.go.dev/github.com/ethereum/go-ethereum/trie)：StateTrie、VerifyProof、VerifyRangeProof 等 API 说明。
7. [Solidity 0.8.27：Storage Layout](https://docs.soliditylang.org/en/v0.8.27/internals/layout_in_storage.html)：mapping 与动态数组的槽定位规则。
8. [EIP-1186：eth_getProof](https://eips.ethereum.org/EIPS/eip-1186)：账户和存储证明接口的定义。EIP 页面的流程状态与客户端实现支持是不同概念，应另核对实际节点能力。

<a id="sec-077"></a>
### 3.2 2 共识层、区块与执行承诺

9. [Ethereum：Light Clients](https://ethereum.org/en/developers/docs/nodes-and-clients/light-clients/)：轻客户端目标与基本信任关系。
10. [Consensus Specs：Altair Light Client Sync Protocol](https://ethereum.github.io/consensus-specs/specs/altair/light-client/sync-protocol/)：启动、更新、委员会和最终性相关认证材料。
11. [Ethereum：SSZ](https://ethereum.org/en/developers/docs/data-structures-and-encoding/ssz/)：类型化编码、树化与广义索引的入门。
12. [Consensus Specs：Capella Beacon Chain](https://ethereum.github.io/consensus-specs/specs/capella/beacon-chain/)：执行负载和共识层负载头的明确结构。
13. [Geth：core/types/block.go](https://github.com/ethereum/go-ethereum/blob/master/core/types/block.go)：执行头、区块体与区块哈希。
14. [Geth：core/types/receipt.go](https://github.com/ethereum/go-ethereum/blob/master/core/types/receipt.go)：收据的共识编码、存储表示与附加字段。
15. [EIP-1559](https://eips.ethereum.org/EIPS/eip-1559)：基础费用和头扩展。
16. [EIP-3675](https://eips.ethereum.org/EIPS/eip-3675)：PoS 切换后的执行字段约束。
17. [EIP-4399](https://eips.ethereum.org/EIPS/eip-4399)：prevRandao 与原 mixHash/DIFFICULTY 语义变化。
18. [EIP-4895](https://eips.ethereum.org/EIPS/eip-4895)：提款列表及执行层提款根。
19. [EIP-4844](https://eips.ethereum.org/EIPS/eip-4844)：blob 交易、相关头字段与数据访问边界。
20. [EIP-4788](https://eips.ethereum.org/EIPS/eip-4788)：父信标区块根。
21. [EIP-7685](https://eips.ethereum.org/EIPS/eip-7685)：执行层请求与 requestsHash 计算。

<a id="sec-078"></a>
### 3.3 3 比特币与 Bloom

22. [Bitcoin Developer Reference：Block Chain](https://developer.bitcoin.org/reference/block_chain.html)：80 字节头、交易 Merkle 根及字节序。
23. [Bitcoin Developer Reference：Transactions](https://developer.bitcoin.org/reference/transactions.html)：交易结构与序列化。
24. [BIP-141](https://github.com/bitcoin/bips/blob/master/bip-0141.mediawiki)：txid、wtxid 与见证承诺。
25. [Bloom，1970 原论文 PDF](https://courses.cs.washington.edu/courses/csep521/21wi/readings/bloom_cacm.pdf)：允许误报的空间/查询权衡。本文实际核对正文中的位数组机制，没有把搜索摘要当论文正文。
26. [Geth：core/types/bloom9.go](https://github.com/ethereum/go-ethereum/blob/master/core/types/bloom9.go)：以太坊 logsBloom 的输入与位映射。
27. [BoomFilters：counting.go](https://github.com/tylertreat/BoomFilters/blob/master/counting.go)：计数 Bloom 的公开 Go 实现阅读入口；并非 Go 标准库或以太坊协议实现。
28. [Bloofi: Multidimensional Bloom Filters，完整正文](https://arxiv.org/html/1501.01941v3)：分层 OR 索引、查询与剪枝局限，重点见第 4 和第 6 节。
29. [BIP-37](https://github.com/bitcoin/bips/blob/master/bip-0037.mediawiki)：Bloom 筛选式轻钱包消息及设计权衡。
30. [BIP-158](https://github.com/bitcoin/bips/blob/master/bip-0158.mediawiki)：紧凑区块过滤器与 Golomb-coded sets。

<a id="sec-079"></a>
### 3.4 4 可以继续深入的学习问题

完成本讲后，适合继续追问的关键在于：

- 当客户端只验证部分状态时，根从哪里来，何时更新？
- 一份证明覆盖的集合和业务要求的集合是否相同？
- 证明如何在多个查询之间共享，又如何防止上下文混用？
- 为保留历史状态，需要付出多少磁盘与写放大成本？
- 对可变数据，如何让更新、快照和证明生成保持一致？
- 对大规模日志，怎样同时设计便宜筛选与完整性校验？
- 一个新的认证结构是否改变了密钥、可信设置、更新模型或数据可用性假设？

这些问题可以形成后续研究或工程实践，但本文没有把它们宣称为已验证的新贡献。选择一个方向时，应先确定查询语义与威胁模型，再对照已有工作和真实成本进行实验。

<a id="sec-080"></a>
### 3.5 5 与现代密码学教材衔接

Jonathan Katz 与 Yehuda Lindell 的《Introduction to Modern Cryptography》第三版可作为理论衔接入口。本次核对的是[作者书目页](https://www.cs.umd.edu/~jkatz/imc.html)以及作者公开的[目录与前言 PDF](https://www.cs.umd.edu/~jkatz/imc/toc-preface-3rd.pdf)，没有取得或声称读过整本教材。

公开目录中的第 6 章讨论哈希函数及应用，其中 §6.6.2 是 Merkle 树、§6.6.5 是承诺方案；第 4 章与第 13 章分别讨论消息认证码和数字签名。学习本讲时，可以据此组织三个问题：哈希安全性怎样支持数据绑定；树怎样把大量数据归约到一个根；验证者凭什么信任这个根。这里的章节对应是学习建议，具体 Ethereum MPT 编码、证明和区块字段仍以本文已引用的协议与实现资料为准。


## 四、哈希指针

普通指针通常表示本地内存地址；哈希指针这一说法强调两项信息：

- 如何定位或引用某份数据；
- 该数据对应的哈希摘要。

在分布式网络中，本地内存地址不能直接作为跨机器引用。实际系统会传递哈希值，并通过本地存储索引或向其他节点请求相应数据。课程用“哈希指针”来描述这种结构关系。

## 五、区块如何连接成链

每个区块头包含前一个区块的哈希。设区块 $B_i$ 的哈希为
$H(B_i)$，下一个区块记录该值：

$$
B_{i+1}.\text{prevHash}=H(B_i)
$$

如果修改较早区块的内容，其哈希会改变；后续区块保存的前序哈希便无法与修改后的内容对应。若要让后续链条重新匹配，攻击者需要重新计算后续区块，并面对共识机制的工作量要求。

这种结构支持篡改检测，并不能单独阻止拥有足够资源的攻击者重算链条。安全性还依赖共识规则、算力分布和节点验证。

## 六、Merkle 树如何工作

Merkle 树是一种用哈希摘要组织多份数据的树形结构：

1.  每笔交易先计算自己的哈希，作为叶节点。
2.  相邻节点的哈希拼接后再次哈希，得到父节点。
3.  持续向上计算，最终得到根哈希（Merkle Root）。
4.  区块头保存根哈希，区块主体保存交易列表。比特币中的 Merkle
    树按区块中的交易组织，主要用于包含证明；它本身不是面向任意键值查找和不存在证明而设计的有序字典结构。

根哈希把整组交易绑定到一个固定长度的摘要中。交易内容改变，相关叶节点及其祖先节点都会变化，最终影响根哈希。

## 七、Merkle Proof：只证明一笔交易

轻节点不必下载整棵树。若要验证交易 $t$ 是否被某个区块承诺，只需取得：

- 交易 $t$ 本身；
- 从该交易叶节点到根节点路径上的兄弟哈希；
- 可用于核对的区块头及其根哈希。

验证方从交易哈希开始，按路径顺序拼接兄弟哈希并逐层计算，最终得到的根哈希若与区块头中的根哈希一致，则交易被包含在该树所承诺的交易集合中。

若树有 $n$ 个叶节点，平衡二叉树中的证明通常只需要 $O(\log n)$
个兄弟哈希，而非整棵树的数据。

### 7.1 证明的边界

Merkle Proof
证明的是“交易被纳入该根哈希所承诺的集合”。它本身不证明交易在业务规则上有效，也不单独证明该区块处于最终被接受的链上。轻节点还需检查区块头链与工作量证明，并依赖全节点对交易有效性的验证。

## 八、区块头与区块主体

- **区块头**：保存前序区块哈希、交易 Merkle
  根、时间戳、难度编码、随机数等字段。
- **区块主体**：保存区块中的交易数据。

区块头较小，便于轻节点同步与验证；区块主体包含具体交易内容，通常需要全节点保存和检查。使用哈希引用构建的数据结构还应避免循环依赖：若两个对象的哈希互相依赖，就无法先确定任一对象的最终内容与哈希。

## 九、为什么这两种结构重要

| 结构      | 主要作用                       | 主要限制                         |
|-----------|--------------------------------|----------------------------------|
| 哈希链    | 连接区块历史、检测历史内容变化 | 需要配合共识才能抵抗重写         |
| Merkle 树 | 汇总交易、提供短包含证明       | 包含证明不等于完整交易验证       |
| 哈希索引  | 根据哈希定位数据               | 具体数据仍需节点保存或从网络取得 |

## 十、本讲小结

哈希链把区块历史连接起来；Merkle
树把区块内部的交易汇总起来。区块头中的哈希摘要使轻节点可以验证部分信息，而不必保存完整账本。

[课程视频：P3
数据结构](https://www.bilibili.com/video/BV1Vt411X7JF/?p=3) ·


## 十一、本篇与知识体系的关系

- 所属部分：第三部分：区块链数据结构与验证
- 返回总目录：[README](../README.md)
