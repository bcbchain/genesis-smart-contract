# Changelog

## v2.2

*March 18, 2020*



### BUG FIXES:

- [跨链] [\#1646](http://114.242.31.175:85/zentao/bug-view-1646.html) UpdateChain() 中修改 yy、jiuj 侧链对应 peerChainBalance值。tokenbasic_ibc.go 中，修改侧链之间转账时保存 peerChainBalance 方法。（@李明磊）

### IMPROVEMENTS
- [代码结构] 添加 tokenbasic_updatechain.go 文件。（@李明磊）
- [跨链] 将RecalAddress替换为RecalAddressEx，以使用新的侧链地址格式。

*Apr 9, 2020*

### BUG FIXES:

- [跨链] [\#1646](https://dc.giblockchain.cn/zentao/bug-view-1646.html) 修复跨链转账目标地址为合约地址时 peerChainBalance 错误 bug。（@李明磊）
- [跨链] [\#1646](https://dc.giblockchain.cn/zentao/bug-view-1646.html) 旧版侧链不能转账到主链。（@李明磊）
