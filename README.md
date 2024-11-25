# 集合竞价撮合

给定一批报单请求，请按“最大成交量、最小剩余量”的集合竞价撮合规则，输出每个合约的
集合竞价撮合价格。
 
## 输入输出说明：
1. 假设程序名为 auctionMatch，以 C++语言举例，请支持按 ./auctionMatch input.csv > 
output.csv 进行运行，其中 input.csv 为输入报单请求文件，output.csv 为结果输出文
件；(其他语言请按照类似格式支持运行)
2. input.csv 的格式为"instrumentID,direction,price,volume"，无表头，instrumentID 最大
支持长度为 30 的字符串，direction 为枚举值（0 为买，1 为卖），price 为浮点数（tick
值满足我所生产上市合约的约束要求），volume 为整数（满足我所生产所有上市合约对
MaxOrderVolume 的约束）例如"IF2412,0,3973.4,3"表示，以 3973.4 的价格买入 IF2412 合
约 3 手；
3. output.csv 的输出格式为"instrumentID,price"，即按序输出 input.csv 中所有合约的开盘
价格（若无开盘价格，则输出空""），输出顺序以合约在 input.csv 中首次出现的顺序为基
准，例如"IF2412,3973.4"；
4. 会在测试环境提供 3 个简单的 input.csv 和 output.csv 测试示例；
5. input.csv 中合约数量不固定，每合约的报单量也不固定，分配无规律；
集合竞价算法流程如下：
注：本流程仅供理解参考
1. 没有买单或卖单，则无成交机会，返回；
2. 取最高买卖价格，若价格不交叉（最高买价<最低卖价），则无成交机会，返回；
3. 根据合约的 tick 值，构造分价表（从最低卖到最高买），即间隔值为 pricetick，大小为
highestBidPrice/Pricetick - lowestAskPrice/priceTick +1 的数组，分价表元素可包
含：买入量、卖出量、累计买入量、累计卖出量、成交量、剩余量；
4. 遍历买单和卖单，更新分价表中各 tick 档位价格对应的买入和卖出量（即对相同价格的买价
量进行汇总，卖同理）；
5. 计算各价格档位上的累计买卖量，即更低买价的累计买量应该包含所有相对更高买价的买入
量（因为低价买能成交，则高价的买一定能成交），累计卖量的算法亦同；
6. 计算各价格档位上的成交量和剩余量，成交量即为累计买量和累计卖量的较小者，剩余量即
为两者差值的绝对值；
7. 计算最大成交量，比较各价格档位上的最大成交量，若只有唯一的价格，则此价格即为最终
集合竞价的价格；
8. 若满足最大成交量的价格有多个，则在这个范围内选取对应剩余量最小的价格，若只有唯一
的价格，则此价格即为最终集合竞价的价格；
9. 如还存在多个价格满足要求，输出“最高价格”即可；
注: 实际不是“最大成交量、最小剩余量、最高价格“的算法，我们在第 9 步中使用”最高价格”
进行了简化，请不要因此误解实际业务。

## 评分规则：

- 正确率优先（5 个用例）；

- 正确率相同的情况下时间优先（按时间排序，如果 t1 <t2 <t3 ，假设(t2-t1) < t1*0.05，则在评分上认为 t1 = t2）；

- 时间相同的情况下按内存占用进行排序（按内存占用排序，如果 m1 <m2 <m3 ，假设(m2-
m1)<m1*0.05 ，则在评分上认为 m1= m2）；

- 在内存占用相等的情况下按代码风格（代码结构清晰简洁，可读性高易理解）进行主观排序；