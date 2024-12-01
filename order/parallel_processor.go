package order

import (
	"AuctionMatch/utils"
	"fmt"
	"strings"
	"sync"
)

type ParallelProcessor struct {
	numWorkers int
}

func (p *ParallelProcessor) Process(stream *OrderStream) []ProcessResult {
	ordersByInstrument := make(map[string][]Order)
	var instrumentsMutex sync.Mutex

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
		instrumentsMutex.Lock()
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
		instrumentsMutex.Unlock()
	}

	results := make([]ProcessResult, len(instrumentOrder))
	var wg sync.WaitGroup

	// 将工作分配给workers
	for i := 0; i < p.numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// 每个worker处理一部分instruments
			for j := workerID; j < len(instrumentOrder); j += p.numWorkers {
				instrumentID := instrumentOrder[j]
				orders := ordersByInstrument[instrumentID]
				price := CalculateAuctionPrice(orders)

				results[j] = ProcessResult{
					InstrumentID: instrumentID,
					Price:        price,
					Scale:        seenInstruments[instrumentID],
				}
			}
		}(i)
	}

	wg.Wait()
	return results
}
