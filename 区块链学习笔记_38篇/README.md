# 区块链与 Web3 学习笔记

本仓库包含 38 篇连续编号的 Markdown 学习笔记，按九个知识部分组织。
内容覆盖区块链基础、密码学、数据结构、共识网络、跨链与模块化架构、Bitcoin、Ethereum、智能合约安全和工程实践。

## 阅读说明

- 建议按编号顺序学习，也可以通过下方目录直接进入专题。
- 每篇笔记均有本篇目录，章节编号在单篇文档内连续排列。
- 正文不依赖本地图片或 `assets` 文件夹，可直接在 GitHub 上阅读。
- Go 与 Solidity 代码示例集中放在 `附加资料/代码示例/`，保留必要的 `go.mod` 和 `go.sum`。
- 数学公式使用 Markdown 数学语法，请使用支持 MathJax 或 KaTeX 的阅读器预览。
- 历史协议与当前主网规则分别说明，具体版本和适用范围以各篇正文为准。

## 学习目录

### 第一部分：区块链基础理论

- [区块链与Web3的关系及发展趋势](第一部分：区块链基础理论/01_区块链与Web3的关系及发展趋势.md)
- [区块链发展历史、定义与核心特性](第一部分：区块链基础理论/02_区块链发展历史、定义与核心特性.md)
- [区块链系统架构与运行流程](第一部分：区块链基础理论/03_区块链系统架构与运行流程.md)

### 第二部分：区块链密码学基础

- [密码学基础与哈希函数](第二部分：区块链密码学基础/04_密码学基础与哈希函数.md)
- [对称加密、非对称加密与密钥管理](第二部分：区块链密码学基础/05_对称加密、非对称加密与密钥管理.md)
- [数字签名、承诺与高级密码学](第二部分：区块链密码学基础/06_数字签名、承诺与高级密码学.md)

### 第三部分：区块链数据结构与验证

- [哈希指针、区块链与Merkle树](第三部分：区块链数据结构与验证/07_哈希指针、区块链与Merkle树.md)
- [区块、交易与UTXO数据结构](第三部分：区块链数据结构与验证/08_区块、交易与UTXO数据结构.md)
- [Trie、Patricia Trie与MPT](第三部分：区块链数据结构与验证/09_Trie、Patricia Trie与MPT.md)
- [状态树、交易树与收据树](第三部分：区块链数据结构与验证/10_状态树、交易树与收据树.md)
- [Bloom Filter、轻节点与数据证明](第三部分：区块链数据结构与验证/11_Bloom Filter、轻节点与数据证明.md)

### 第四部分：共识算法与P2P网络

- [分布式共识基础与FLP理论](第四部分：共识算法与P2P网络/12_分布式共识基础与FLP理论.md)
- [PoW、挖矿与难度调整](第四部分：共识算法与P2P网络/13_PoW、挖矿与难度调整.md)
- [PoS、最终性与以太坊共识](第四部分：共识算法与P2P网络/14_PoS、最终性与以太坊共识.md)
- [DPoS、PBFT与拜占庭容错](第四部分：共识算法与P2P网络/15_DPoS、PBFT与拜占庭容错.md)
- [P2P网络、Gossip与网络安全](第四部分：共识算法与P2P网络/16_P2P网络、Gossip与网络安全.md)

### 第五部分：链类型、跨链与模块化区块链

- [公链、私有链与联盟链](第五部分：链类型、跨链与模块化区块链/17_公链、私有链与联盟链.md)
- [跨链互操作与跨链桥](第五部分：链类型、跨链与模块化区块链/18_跨链互操作与跨链桥.md)
- [Layer1、Layer2与Rollup](第五部分：链类型、跨链与模块化区块链/19_Layer1、Layer2与Rollup.md)
- [模块化架构、DA与扩容权衡](第五部分：链类型、跨链与模块化区块链/20_模块化架构、DA与扩容权衡.md)

### 第六部分：比特币原理与生态

- [Bitcoin架构、交易与共识流程](第六部分：比特币原理与生态/21_Bitcoin架构、交易与共识流程.md)
- [UTXO、交易费与双花防护](第六部分：比特币原理与生态/22_UTXO、交易费与双花防护.md)
- [比特币脚本、地址与多重签名](第六部分：比特币原理与生态/23_比特币脚本、地址与多重签名.md)
- [Taproot、Schnorr与比特币隐私](第六部分：比特币原理与生态/24_Taproot、Schnorr与比特币隐私.md)
- [比特币生态协议与扩展应用](第六部分：比特币原理与生态/25_比特币生态协议与扩展应用.md)

### 第七部分：以太坊架构、共识与EVM

- [以太坊发展历史与整体架构](第七部分：以太坊架构、共识与EVM/26_以太坊发展历史与整体架构.md)
- [账户模型、交易与状态转换](第七部分：以太坊架构、共识与EVM/27_账户模型、交易与状态转换.md)
- [以太坊PoS共识、验证者与最终性](第七部分：以太坊架构、共识与EVM/28_以太坊PoS共识、验证者与最终性.md)
- [EVM执行模型、Gas与交易回滚](第七部分：以太坊架构、共识与EVM/29_EVM执行模型、Gas与交易回滚.md)
- [以太坊生态协议与客户端实现](第七部分：以太坊架构、共识与EVM/30_以太坊生态协议与客户端实现.md)

### 第八部分：智能合约与安全

- [Solidity智能合约开发基础](第八部分：智能合约与安全/31_Solidity智能合约开发基础.md)
- [智能合约安全：重入、整数溢出与DoS](第八部分：智能合约与安全/32_智能合约安全：重入、整数溢出与DoS.md)
- [钱包、Token、NFT与DeFi安全](第八部分：智能合约与安全/33_钱包、Token、NFT与DeFi安全.md)
- [链上治理、分叉与安全事件分析](第八部分：智能合约与安全/34_链上治理、分叉与安全事件分析.md)

### 第九部分：工程工具与综合实践

- [区块链开发工具、节点与RPC](第九部分：工程工具与综合实践/35_区块链开发工具、节点与RPC.md)
- [Go与Solidity工程规范、测试与审计](第九部分：工程工具与综合实践/36_Go与Solidity工程规范、测试与审计.md)
- [链上数据分析、监控与系统运维](第九部分：工程工具与综合实践/37_链上数据分析、监控与系统运维.md)
- [Web3综合项目设计与技术选型](第九部分：工程工具与综合实践/38_Web3综合项目设计与技术选型.md)

## 附加代码示例

Go 与 Solidity 示例位于 [`附加资料/代码示例/`](附加资料/代码示例/)。运行前请查看对应目录中的依赖和版本配置。

## 主要官方资料入口

- [Bitcoin Core](https://github.com/bitcoin/bitcoin)
- [Bitcoin Developer Documentation](https://developer.bitcoin.org/)
- [Ethereum Developer Documentation](https://ethereum.org/developers/docs/)
- [Ethereum Execution Specs](https://github.com/ethereum/execution-specs)
- [Ethereum Consensus Specs](https://github.com/ethereum/consensus-specs)
- [Ethereum EIPs](https://eips.ethereum.org/)
- [Solidity Documentation](https://docs.soliditylang.org/)
- [OpenZeppelin Contracts](https://github.com/OpenZeppelin/openzeppelin-contracts)
- [Solana Documentation](https://solana.com/docs)
- [Polkadot Developer Docs](https://docs.polkadot.com/)
- [Cosmos Documentation](https://docs.cosmos.network/)
- [Hyperledger Fabric](https://hyperledger-fabric.readthedocs.io/)
