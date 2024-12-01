package order

import (
	"AuctionMatch/utils"
	"fmt"
	"strings"
)

type SingleProcessor struct {
}

func (p *SingleProcessor) Process(stream *OrderStream) []ProcessResult {
	ordersByInstrument := make(map[string][]Order)

	// 跟踪合约出现顺序
	instrumentOrder := make([]string, 0)
	seenInstruments := make(map[string]uint) // 记录每个合约的价格精度
	var scale uint = 0

	// 收集订单
	for line := range stream.Orders {
		record := utils.CustomSplit(line)
		if !IsValidRecord(record) {
			continue
		}
		order, err := ParseOrder(record)
		if err != nil {
			stream.Error <- fmt.Errorf("解析订单出错: %v", err)
			continue
		}
		if seenInstruments[order.InstrumentID] == 0 {
			// 获取价格精度
			if dotIndex := strings.Index(record[2], "."); dotIndex != -1 {
				scale = uint(len(record[2]) - dotIndex - 1)
			}
			instrumentOrder = append(instrumentOrder, order.InstrumentID)
			seenInstruments[order.InstrumentID] = scale
			scale = 0
		}

		ordersByInstrument[order.InstrumentID] = append(
			ordersByInstrument[order.InstrumentID],
			order,
		)
	}

	results := make([]ProcessResult, len(instrumentOrder))

	// 按照顺序计算集合竞价价格
	for i, instrumentID := range instrumentOrder {
		orders := ordersByInstrument[instrumentID]
		price := CalculateAuctionPrice(orders)
		results[i] = ProcessResult{
			InstrumentID: instrumentID,
			Price:        price,
			Scale:        seenInstruments[instrumentID],
		}
	}

	return results
}
