package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
	"sync"

	"AuctionMatch/utils"
)

// Order 表示一个订单
type Order struct {
	InstrumentID string
	Direction    int // 0:买, 1:卖
	Price        float64
	Volume       int
}

// PriceLevel 价格档位信息
type PriceLevel struct {
	Price      float64
	BuyVolume  int
	SellVolume int
}

const (
	PRICE_TICK   = 0.2 // 价格最小变动单位
	WORKER_COUNT = 4   // 并发工作协程数
)

// OrderBatch 表示一批需要处理的订单
type OrderBatch struct {
	Orders       []Order
	InstrumentID string
}

// PriceLevelMap 价格档位映射
type PriceLevelMap struct {
	buyLevels  map[int64]int // 买单价格档位
	sellLevels map[int64]int // 卖单价格档位
	highestBid float64       // 最高买单价格
	lowestAsk  float64       // 最低卖单价格
	sync.RWMutex
}

func NewPriceLevelMap() *PriceLevelMap {
	return &PriceLevelMap{
		buyLevels:  make(map[int64]int),
		sellLevels: make(map[int64]int),
		highestBid: -1,
		lowestAsk:  -1,
	}
}

func main() {
	// 检查参数
	if len(os.Args) == 2 && os.Args[1] == "-h" {
		printUsage()
		return
	}

	if len(os.Args) < 2 {
		fmt.Println("参数错误！使用 -h 查看帮助信息")
		return
	}

	// 读取输入文件
	orders, instruments := readOrders(os.Args[1])

	// 如果没有订单或合约，直接返回
	if len(orders) == 0 || len(instruments) == 0 {
		writeResults(make(map[string]float64), instruments, "")
		return
	}

	// 使用缓冲channel进行并发处理
	workerCount := utils.Min(WORKER_COUNT, len(instruments)) // 根据合约数量调整worker数
	batchChan := make(chan OrderBatch, workerCount)
	resultChan := make(chan struct {
		instrumentID string
		price        float64
	}, len(instruments)) // 使用缓冲channel

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for batch := range batchChan {
				price := calculateAuctionPrice(batch.Orders)
				resultChan <- struct {
					instrumentID string
					price        float64
				}{batch.InstrumentID, price}
			}
		}()
	}

	// 确保正确初始化 map
	results := make(map[string]float64)
	ordersByInstrument := make(map[string][]Order)

	// 分发订单批次
	for _, order := range orders {
		if ordersByInstrument[order.InstrumentID] == nil {
			ordersByInstrument[order.InstrumentID] = make([]Order, 0)
		}
		ordersByInstrument[order.InstrumentID] = append(
			ordersByInstrument[order.InstrumentID],
			order,
		)
	}

	// 启动结果收集协程
	done := make(chan bool)
	go func() {
		receivedCount := 0
		for result := range resultChan {
			results[result.instrumentID] = result.price
			receivedCount++
			if receivedCount == len(instruments) {
				close(done)
				return
			}
		}
	}()

	// 发送订单批次
	for _, instrumentID := range instruments {
		batchChan <- OrderBatch{
			Orders:       ordersByInstrument[instrumentID],
			InstrumentID: instrumentID,
		}
	}
	close(batchChan)

	// 等待所有处理完成
	wg.Wait()
	close(resultChan)
	<-done

	// 输出结果
	if len(os.Args) == 4 && os.Args[2] == "-o" {
		writeResults(results, instruments, os.Args[3])
	} else {
		writeResults(results, instruments, "")
	}
}

func printUsage() {
	fmt.Println("集合竞价撮合程序")
	fmt.Println("\n用法:")
	fmt.Println("  ./auctionMatch <input.csv> -o <output.csv>")
	fmt.Println("  ./auctionMatch -h")
	fmt.Println("\n参数:")
	fmt.Println("  input.csv    输入的订单CSV文件")
	fmt.Println("  output.csv   输出的结果CSV文件")
	fmt.Println("  -h          显示帮助信息")
	fmt.Println("\n示例:")
	fmt.Println("  ./auctionMatch orders.csv -o results.csv")
}

// readOrders 读取CSV文件并解析订单
func readOrders(filename string) ([]Order, []string) {
	// 打开文件
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("无法打开文件: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// 创建CSV reader
	reader := csv.NewReader(file)

	var orders []Order
	instrumentSet := make(map[string]struct{}) // 用于去重
	var instruments []string                   // 保持合约的首次出现顺序

	// 逐行读取CSV
	for {
		record, err := reader.Read()
		if err != nil {
			break // 文件结束或发生错误时退出
		}

		// 确保每行有4个字段
		if len(record) != 4 {
			fmt.Printf("无效的CSV行: %v\n", record)
			continue
		}

		// 解析direction
		direction, err := strconv.Atoi(record[1])
		if err != nil || (direction != 0 && direction != 1) {
			fmt.Printf("无效的direction值: %s\n", record[1])
			continue
		}

		// 解析price
		price, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			fmt.Printf("无效的price值: %s\n", record[2])
			continue
		}

		// 析volume
		volume, err := strconv.Atoi(record[3])
		if err != nil {
			fmt.Printf("无效的volume值: %s\n", record[3])
			continue
		}

		// 创建订单对象
		order := Order{
			InstrumentID: record[0],
			Direction:    direction,
			Price:        price,
			Volume:       volume,
		}

		// 添加订单到列表
		orders = append(orders, order)

		// 如果是新的合约ID，添加到instruments列表
		if _, exists := instrumentSet[order.InstrumentID]; !exists {
			instrumentSet[order.InstrumentID] = struct{}{}
			instruments = append(instruments, order.InstrumentID)
		}
	}

	return orders, instruments
}

// 价格计算函数
func calculateAuctionPrice(orders []Order) float64 {
	if len(orders) == 0 {
		return 0
	}

	priceMap := NewPriceLevelMap()
	prices := utils.NewOrderedSet()

	// 第一次遍历：收集价格点和统计量
	for _, order := range orders {
		priceInt := toInt(order.Price)

		prices.Add(priceInt)

		if order.Direction == 0 { // 买单
			priceMap.buyLevels[priceInt] += order.Volume
			if priceMap.highestBid == -1 || order.Price > priceMap.highestBid {
				priceMap.highestBid = order.Price
			}
		} else { // 卖单
			priceMap.sellLevels[priceInt] += order.Volume
			if priceMap.lowestAsk == -1 || order.Price < priceMap.lowestAsk {
				priceMap.lowestAsk = order.Price
			}
		}
	}

	if priceMap.highestBid < priceMap.lowestAsk ||
		priceMap.highestBid == -1 ||
		priceMap.lowestAsk == -1 {
		return 0
	}

	// 预先计算每个价格点的买卖量
	type PricePoint struct {
		price      int64
		buyVolume  int
		sellVolume int
	}

	maxMatchVolume := -1             // 最大成交量
	minRemainVolume := math.MaxInt32 // 最小剩余量
	var bestPrice int64              // 最佳竞价

	accumBuy := 0  // 当前价格及以上的累计买量
	accumSell := 0 // 所有卖单总量

	// 将所有价格点从高到低排序并预处理数据
	pricePoints := make([]PricePoint, 0, prices.Len())
	for _, priceInt := range prices.GetSorted(true) {
		sellVolume := priceMap.sellLevels[priceInt]
		pricePoints = append(pricePoints, PricePoint{
			price:      priceInt,
			buyVolume:  priceMap.buyLevels[priceInt],
			sellVolume: sellVolume,
		})
		accumSell += sellVolume
	}

	// 只需要遍历一次价格点
	for _, pp := range pricePoints {
		accumBuy += pp.buyVolume

		// 当前价格以下的累计卖量
		currentSell := accumSell

		matchVolume := utils.Min(accumBuy, currentSell)
		remainVolume := utils.Abs(accumBuy - currentSell)

		if matchVolume > maxMatchVolume ||
			(matchVolume == maxMatchVolume && remainVolume < minRemainVolume) {
			maxMatchVolume = matchVolume
			minRemainVolume = remainVolume
			bestPrice = pp.price
		}

		// 为下一个价格点更新累计卖量
		accumSell -= pp.sellVolume
	}

	if maxMatchVolume <= 0 {
		return 0
	}

	return toFloat(bestPrice)
}

// 工具函数
func toInt(price float64) int64 {
	return int64(math.Round(price / PRICE_TICK))
}

func toFloat(priceInt int64) float64 {
	return float64(priceInt) * PRICE_TICK
}

// writeResults 将结果写入标准输出
func writeResults(results map[string]float64, instruments []string, outputFile string) {
	// 创建CSV writer
	var writer *csv.Writer
	if outputFile == "" {
		writer = csv.NewWriter(os.Stdout)
	} else {
		file, err := os.Create(outputFile)
		if err != nil {
			fmt.Printf("无法创建输出文件: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		writer = csv.NewWriter(file)
	}
	defer writer.Flush()

	// 按照合约首次出现顺序输出结果
	for _, instrumentID := range instruments {
		price := results[instrumentID]
		var record []string

		if price == 0 {
			// 如果价格为0（无法撮合），输出空字符串
			record = []string{instrumentID, ""}
		} else {
			// 格式化价格，保留一位小数
			record = []string{instrumentID, fmt.Sprintf("%.1f", price)}
		}

		if err := writer.Write(record); err != nil {
			fmt.Printf("写入结果时发生错误: %v\n", err)
			os.Exit(1)
		}
	}
}
