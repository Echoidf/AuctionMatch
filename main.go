package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"

	"AuctionMatch/utils"
)

// Order 表示一个订单
type Order struct {
	InstrumentID string
	Direction    int // 0:买, 1:卖
	Price        float64
	Volume       int
}

// PriceLevel 表示价格档位信息
type PriceLevel struct {
	Price           float64
	BuyVolume       int // 买入量
	SellVolume      int // 卖出量
	AccumBuyVolume  int // 累计买入量
	AccumSellVolume int // 累计卖出量
	MatchVolume     int // 成交量
	RemainVolume    int // 剩余量
}

func main() {
	// 检查是否请求帮助信息
	if len(os.Args) == 2 && os.Args[1] == "-h" {
		printUsage()
		return
	}

	// 检查参数
	if len(os.Args) < 2 {
		fmt.Println("参数错误！使用 -h 查看帮助信息")
		return
	}

	// 读取输入文件
	orders, instruments := readOrders(os.Args[1])

	// 将订单按合约ID分组
	ordersByInstrument := make(map[string][]Order)
	for _, order := range orders {
		ordersByInstrument[order.InstrumentID] = append(ordersByInstrument[order.InstrumentID], order)
	}

	// 处理每个合约
	results := make(map[string]float64)
	for _, instrumentID := range instruments {
		price := calculateAuctionPrice(ordersByInstrument[instrumentID])
		results[instrumentID] = price
	}

	// 根据参数决定输出方式
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

		// 解析volume
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

const PRICE_TICK = 0.2 // 价格最小变动单位

// calculateAuctionPrice 计算某个合约的集合竞价价格
func calculateAuctionPrice(orders []Order) float64 {
	if len(orders) == 0 {
		return 0
	}

	// 1. 分离买卖订单并找出最高买价和最低卖价
	var buyOrders, sellOrders []Order
	var highestBid float64 = -1
	var lowestAsk float64 = -1

	for _, order := range orders {
		if order.Direction == 0 { // 买单
			buyOrders = append(buyOrders, order)
			if highestBid == -1 || order.Price > highestBid {
				highestBid = order.Price
			}
		} else { // 卖单
			sellOrders = append(sellOrders, order)
			if lowestAsk == -1 || order.Price < lowestAsk {
				lowestAsk = order.Price
			}
		}
	}

	// 2. 检查是否有买卖订单
	if len(buyOrders) == 0 || len(sellOrders) == 0 {
		return 0
	}

	// 3. 检查价格是否交叉
	if highestBid < lowestAsk {
		return 0
	}

	// 4. 构造价格档位表
	priceLevels := make(map[int64]*PriceLevel)
	// 将价格转换为整数表示（乘以一个足够大的数以保持精度）
	toInt := func(price float64) int64 {
		return int64(math.Round(price / PRICE_TICK))
	}
	toFloat := func(priceInt int64) float64 {
		return float64(priceInt) * PRICE_TICK
	}
	// 从最低卖价到最高买价，按PRICE_TICK递增
	for priceInt := toInt(lowestAsk); priceInt <= toInt(highestBid); priceInt++ {
		price := toFloat(priceInt)
		priceLevels[priceInt] = &PriceLevel{Price: price}
	}

	// 5. 统计各价格档位的买卖量
	for _, order := range buyOrders {
		orderPriceInt := toInt(order.Price)
		if level, exists := priceLevels[orderPriceInt]; exists {
			level.BuyVolume += order.Volume
		}
	}

	for _, order := range sellOrders {
		orderPriceInt := toInt(order.Price)
		if level, exists := priceLevels[orderPriceInt]; exists {
			level.SellVolume += order.Volume
		}
	}

	// 5.1 计算累计买卖量
	var prevBuyVolume, prevSellVolume int
	// 从高到低计算累计买量
	for priceInt := toInt(highestBid); priceInt >= toInt(lowestAsk); priceInt-- {
		if level, exists := priceLevels[priceInt]; exists {
			level.AccumBuyVolume = level.BuyVolume + prevBuyVolume
			prevBuyVolume = level.AccumBuyVolume
		}
	}
	// 从低到高计算累计卖量
	for priceInt := toInt(lowestAsk); priceInt <= toInt(highestBid); priceInt++ {
		if level, exists := priceLevels[priceInt]; exists {
			level.AccumSellVolume = level.SellVolume + prevSellVolume
			prevSellVolume = level.AccumSellVolume
		}
	}

	// 6. 计算每个价格档位的成交量和剩余量
	var maxMatchVolume int = -1
	var minRemainVolume int = math.MaxInt32
	var candidatePriceInts []int64

	// 第一轮：找出最大成交量
	for _, level := range priceLevels {
		level.MatchVolume = utils.Min(level.AccumBuyVolume, level.AccumSellVolume)
		level.RemainVolume = utils.Abs(level.AccumBuyVolume - level.AccumSellVolume)
		if level.MatchVolume > maxMatchVolume {
			maxMatchVolume = level.MatchVolume
		}
	}

	// 第二轮：在最大成交量的价位中找出最小剩余量
	for priceInt, level := range priceLevels {
		if level.MatchVolume == maxMatchVolume {
			if level.RemainVolume < minRemainVolume {
				minRemainVolume = level.RemainVolume
				candidatePriceInts = []int64{priceInt}
			} else if level.RemainVolume == minRemainVolume {
				candidatePriceInts = append(candidatePriceInts, priceInt)
			}
		}
	}

	// 7. 如果没有找到合适的价格，返回0
	if len(candidatePriceInts) == 0 {
		return 0
	}

	// 8. 返回候选价格中的最高价格
	maxPriceInt := utils.FindMax(candidatePriceInts)
	return toFloat(maxPriceInt)
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
