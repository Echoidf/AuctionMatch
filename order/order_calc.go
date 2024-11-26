package order

import (
	"AuctionMatch/utils"
	"math"
)

type PricePoint struct {
	price      int64 // 价格（以tick为单位）
	buyVolume  int   // 该价格的买单量
	sellVolume int   // 该价格的卖单量
}

// 工具函数
func ToInt(price float32, tick float32) int64 {
	return int64(math.Round(float64(price / tick)))
}

func ToFloat(priceInt int64, tick float32) float32 {
	return float32(priceInt) * tick
}

// 集合竞价计算函数
func CalculateAuctionPrice(orders []Order) float32 {
	if len(orders) == 0 {
		return 0
	}

	priceMap := NewPriceLevelMap()

	tick := orders[0].GetTick()

	for _, order := range orders {
		// 转为tick数
		priceInt := ToInt(order.Price, tick)

		if order.Direction == 0 { // 买单
			priceMap.buyLevels[priceInt] += order.Volume
			// 维护最高买价
			if priceMap.highestBid == -1 || order.Price > priceMap.highestBid {
				priceMap.highestBid = order.Price
			}
		} else { // 卖单
			priceMap.sellLevels[priceInt] += order.Volume
			// 维护最低卖价
			if priceMap.lowestAsk == -1 || order.Price < priceMap.lowestAsk {
				priceMap.lowestAsk = order.Price
			}
		}
	}

	// 修改价格点的收集逻辑
	lowestPriceInt := ToInt(priceMap.lowestAsk, tick)
	highestPriceInt := ToInt(priceMap.highestBid, tick)

	// 如果最高买价低于最低卖价，则没有成交
	if priceMap.highestBid < priceMap.lowestAsk ||
		priceMap.highestBid == -1 ||
		priceMap.lowestAsk == -1 {
		return 0
	}

	// 构造完整的分价表
	pricePoints := make([]PricePoint, 0)
	for priceInt := highestPriceInt; priceInt >= lowestPriceInt; priceInt-- {
		pricePoints = append(pricePoints, PricePoint{
			price:      priceInt,
			buyVolume:  priceMap.buyLevels[priceInt],
			sellVolume: priceMap.sellLevels[priceInt],
		})
	}

	maxMatchVolume := -1
	minRemainVolume := math.MaxInt32
	var bestPrice int64

	accumBuy := 0
	accumSell := 0

	// 计算所有卖单总量
	for _, pp := range pricePoints {
		accumSell += pp.sellVolume
	}

	// 从高到低遍历所有价格点
	for _, pp := range pricePoints {
		accumBuy += pp.buyVolume

		matchVolume := utils.Min(accumBuy, accumSell)
		remainVolume := utils.Abs(accumBuy - accumSell)

		if matchVolume > maxMatchVolume ||
			(matchVolume == maxMatchVolume && remainVolume < minRemainVolume) {
			maxMatchVolume = matchVolume
			minRemainVolume = remainVolume
			bestPrice = pp.price
		}

		accumSell -= pp.sellVolume
	}

	if maxMatchVolume <= 0 {
		return 0
	}

	return ToFloat(bestPrice, tick)
}
