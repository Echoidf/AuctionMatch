package order

import (
	"fmt"
	"strings"
)

type SingleProcessor struct {
}

func (p *SingleProcessor) Process(stream *OrderStream) []ProcessResult {
	ordersByInstrument := make(map[string][]Order)

	// 跟踪合约出现顺序
	instrumentOrder := make([]string, 0)
	seenInstruments := make(map[string]bool)
	scale := 0

	for line := range stream.Orders {
		record := strings.Split(line, ",")
		if !IsValidRecord(record) {
			continue
		}
		order, err := ParseOrder(record)
		if err != nil {
			stream.Error <- fmt.Errorf("解析订单出错: %v", err)
			continue
		}
		if !seenInstruments[order.InstrumentID] {
			// 获取价格精度
			if dotIndex := strings.Index(record[2], "."); dotIndex != -1 {
				scale = len(record[2]) - dotIndex - 1
			}
			instrumentOrder = append(instrumentOrder, order.InstrumentID)
			seenInstruments[order.InstrumentID] = true
		}

		ordersByInstrument[order.InstrumentID] = append(
			ordersByInstrument[order.InstrumentID],
			order,
		)
	}

	// 创建有序的结果切片
	results := make([]ProcessResult, len(instrumentOrder))

	// 按照首次出现顺序处理合约
	for i, instrumentID := range instrumentOrder {
		orders := ordersByInstrument[instrumentID]
		price := CalculateAuctionPrice(orders)
		results[i] = ProcessResult{
			InstrumentID: instrumentID,
			Price:        price,
			Scale:        uint(scale),
		}
	}

	return results
}
