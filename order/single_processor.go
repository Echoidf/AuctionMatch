package order

type SingleProcessor struct {
	batchSize int
}

func (p *SingleProcessor) Process(stream *OrderStream) []ProcessResult {
	ordersByInstrument := make(map[string][]Order)

	// 跟踪合约出现顺序
	instrumentOrder := make([]string, 0)
	seenInstruments := make(map[string]bool)

	for order := range stream.Orders {
		if !seenInstruments[order.InstrumentID] {
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
		results[i] = struct {
			InstrumentID string
			Price        float32
		}{
			InstrumentID: instrumentID,
			Price:        price,
		}
	}

	return results
}
