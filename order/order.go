package order

import (
	"AuctionMatch/consts"
	"sync"
)

type (
	// OrderStream 订单流
	OrderStream struct {
		Orders   chan string
		Error    chan error
		Done     chan struct{}
		ChunkNum uint
	}
	// Order 订单
	Order struct {
		InstrumentID string
		Direction    int8 // 0:买, 1:卖
		Price        float32
		Volume       int32
	}
	// PriceLevel 价格档位信息
	PriceLevel struct {
		Price      float64
		BuyVolume  int
		SellVolume int
	}

	// OrderBatch 表示一批需要处理的订单
	OrderBatch struct {
		Orders       []Order
		InstrumentID string
	}

	// PriceLevelMap 价格档位映射
	PriceLevelMap struct {
		buyLevels  map[int64]int32 // 买单价格档位
		sellLevels map[int64]int32 // 卖单价格档位
		highestBid float32         // 最高买单价格
		lowestAsk  float32         // 最低卖单价格
		sync.RWMutex
	}
)

const (
	PRICE_TICK   = 0.2 // 价格最小变动单位
	WORKER_COUNT = 4   // 并发工作协程数
)

func NewOrderStream() *OrderStream {
	return &OrderStream{
		Orders: make(chan string, 1000),
		Error:  make(chan error, 1),
		Done:   make(chan struct{}),
	}
}

func NewPriceLevelMap() *PriceLevelMap {
	return &PriceLevelMap{
		buyLevels:  make(map[int64]int32),
		sellLevels: make(map[int64]int32),
		highestBid: -1,
		lowestAsk:  -1,
	}
}

func (order *Order) GetTick() float32 {
	// 从合约ID中提取品种代码（例如从"IF2306"中提取"IF"）
	var productCode string
	for i, c := range order.InstrumentID {
		if i >= 2 || !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')) {
			break
		}
		productCode += string(c)
	}

	// 从CFE_PRODUCT_TICK中获取对应的tick值
	if tick, ok := consts.CFE_PRODUCT_TICK[productCode]; ok {
		return tick
	}

	// 如果找不到对应的tick值，返回默认值0.2
	return PRICE_TICK
}
