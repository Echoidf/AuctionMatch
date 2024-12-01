package order

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// 处理器接口
type (
	ProcessResult struct {
		InstrumentID string
		Price        float32
		Scale        uint // 精度
	}
	OrderProcessor interface {
		Process(stream *OrderStream) []ProcessResult
	}
)

// 创建处理器工厂函数
func NewOrderProcessor(numCPU int) OrderProcessor {
	if numCPU <= 1 {
		return &SingleProcessor{}
	} else {
		return &ParallelProcessor{numWorkers: numCPU}
	}
}

// 辅助函数：验证记录的有效性
func IsValidRecord(record []string) bool {
	return len(record) == 4
}

// 辅助函数：解析订单数据
func ParseOrder(record []string) (Order, error) {
	direction, err := strconv.Atoi(record[1])
	if err != nil || (direction != 0 && direction != 1) {
		return Order{}, fmt.Errorf("无效的direction值: %s", record[1])
	}

	price, err := strconv.ParseFloat(record[2], 64)
	if err != nil {
		return Order{}, fmt.Errorf("无效的price值: %s", record[2])
	}

	volume, err := strconv.Atoi(record[3])
	if err != nil {
		return Order{}, fmt.Errorf("无效的volume值: %s", record[3])
	}

	return Order{
		InstrumentID: record[0],
		Direction:    int8(direction),
		Price:        float32(price),
		Volume:       int32(volume),
	}, nil
}

// streamOrders 流式读取CSV文件
func StreamOrders(filename string) *OrderStream {
	stream := NewOrderStream()

	go func() {
		defer close(stream.Orders)
		// defer close(stream.Error)
		defer close(stream.Done)

		file, _ := os.Open(filename)
		// if err != nil {
		// 	stream.Error <- fmt.Errorf("无法打开文件: %v", err)
		// 	return
		// }
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			// 发送订单到channel
			stream.Orders <- line
		}
	}()

	return stream
}
