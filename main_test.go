package main

import (
	"AuctionMatch/order"
	"AuctionMatch/utils"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// TestCalculateAuctionPrice 测试价格计算逻辑
func TestCalculateAuctionPrice(t *testing.T) {
	tests := []struct {
		name          string
		orders        []order.Order
		expectedPrice float32
	}{
		{
			name: "正常撮合",
			orders: []order.Order{
				{InstrumentID: "IF2412", Direction: 0, Price: 3973.4, Volume: 3}, // 买单
				{InstrumentID: "IF2412", Direction: 1, Price: 3973.2, Volume: 2}, // 卖单
			},
			expectedPrice: 3973.4,
		},
		{
			name: "无买单",
			orders: []order.Order{
				{InstrumentID: "IF2412", Direction: 1, Price: 3973.2, Volume: 2},
			},
			expectedPrice: 0,
		},
		{
			name: "无卖单",
			orders: []order.Order{
				{InstrumentID: "IF2412", Direction: 0, Price: 3973.4, Volume: 3},
			},
			expectedPrice: 0,
		},
		{
			name: "价格不交叉",
			orders: []order.Order{
				{InstrumentID: "IF2412", Direction: 0, Price: 3970.0, Volume: 3},
				{InstrumentID: "IF2412", Direction: 1, Price: 3975.0, Volume: 2},
			},
			expectedPrice: 0,
		},
		{
			name: "多个价格档位测试",
			orders: []order.Order{
				{InstrumentID: "IF2412", Direction: 0, Price: 3973.4, Volume: 3}, // 买单
				{InstrumentID: "IF2412", Direction: 0, Price: 3973.2, Volume: 2}, // 买单
				{InstrumentID: "IF2412", Direction: 1, Price: 3973.0, Volume: 4}, // 卖单
			},
			expectedPrice: 3973.2,
		},
		{
			name: "相同成交量不同剩余量测试",
			orders: []order.Order{
				{InstrumentID: "IF2412", Direction: 0, Price: 3973.4, Volume: 3}, // 买单
				{InstrumentID: "IF2412", Direction: 1, Price: 3973.2, Volume: 2}, // 卖单
				{InstrumentID: "IF2412", Direction: 1, Price: 3973.0, Volume: 1}, // 卖单
			},
			expectedPrice: 3973.4, // 在最大成交量和最小剩余量相同时，选择最高价格
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price := order.CalculateAuctionPrice(tt.orders)
			if !utils.FloatEquals(price, tt.expectedPrice) {
				t.Errorf("期望价格 %.1f, 实际价格 %.1f", tt.expectedPrice, price)
			}
		})
	}
}

// TestReadOrders 测试订单读取逻辑
func TestReadOrders(t *testing.T) {
	// 创建测试数据
	inputContent := `IF2412,0,3973.4,3
IF2306,0,3972.2,2
IF2412,1,3973.2,2
IF2306,1,3972.0,1`

	// 创建临时文件
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test_input.csv")
	err := os.WriteFile(inputFile, []byte(inputContent), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 启动流式读取
	stream := order.StreamOrders(inputFile)

	// 准备收集数据的容器
	var orders []order.Order

	// 收集错误
	var errors []error
	go func() {
		for err := range stream.Error {
			errors = append(errors, err)
		}
	}()

	// 收集订单
	for order := range stream.Orders {
		orders = append(orders, order)
	}

	// 等待处理完成
	<-stream.Done

	// 验证测试结果
	t.Run("验证订单数量", func(t *testing.T) {
		expectedOrderCount := 4
		if len(orders) != expectedOrderCount {
			t.Errorf("期望订单数量 %d, 实际数量 %d", expectedOrderCount, len(orders))
		}
	})
	t.Run("验证订单内容", func(t *testing.T) {
		// 验证第一个订单
		if len(orders) > 0 {
			firstOrder := orders[0]
			expectedOrder := order.Order{
				InstrumentID: "IF2412",
				Direction:    0,
				Price:        3973.4,
				Volume:       3,
			}
			if !orderEqual(firstOrder, expectedOrder) {
				t.Errorf("第一个订单不匹配\n期望: %+v\n实际: %+v", expectedOrder, firstOrder)
			}
		}
	})

	t.Run("验证错误处理", func(t *testing.T) {
		if len(errors) > 0 {
			t.Errorf("读取过程中出现意外错误: %v", errors)
		}
	})
}

// 辅助函数：比较两个订单是否相等
func orderEqual(a, b order.Order) bool {
	return a.InstrumentID == b.InstrumentID &&
		a.Direction == b.Direction &&
		utils.FloatEquals(a.Price, b.Price) &&
		a.Volume == b.Volume
}

// generateLargeTestData 生成大批量测试数据
func generateLargeTestData(numOrders int) string {
	var buffer bytes.Buffer
	instruments := []string{"IF2412", "IF2306", "IF2403", "IF2406"}

	for i := 0; i < numOrders; i++ {
		// 随机选择合约
		instrument := instruments[i%len(instruments)]
		// 随机生成买卖方向 (0或1)
		direction := i % 2
		// 生成基准价格 3900.0 到 4000.0 之间的随机价格
		basePrice := 3900.0 + float64(i%100)
		// 添加随机小数位
		price := basePrice + float64(i%10)*0.2
		// 生成 1-10 之间的随机数量
		volume := 1 + (i % 10)

		buffer.WriteString(fmt.Sprintf("%s,%d,%.1f,%d\n",
			instrument, direction, price, volume))
	}
	return buffer.String()
}

func TestGenerateLargeTestData(t *testing.T) {
	data := generateLargeTestData(1000000)
	// write to file
	os.WriteFile("large_test_data.csv", []byte(data), 0644)
}

// TestStreamOrders 测试流式读取订单
func TestStreamOrders(t *testing.T) {
	// 创建测试数据
	inputContent := `IF2412,0,3973.4,3
IF2306,0,3972.2,2
IF2412,1,3973.2,2
IF2306,1,3972.0,1`

	// 创建临时文件
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test_input.csv")
	err := os.WriteFile(inputFile, []byte(inputContent), 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}

	// 启动流式读取
	stream := order.StreamOrders(inputFile)

	// 收集结果
	var orders []order.Order

	// 处理错误
	go func() {
		for err := range stream.Error {
			t.Errorf("读取过程中出现错误: %v", err)
		}
	}()

	// 收集订单
	for order := range stream.Orders {
		orders = append(orders, order)
	}

	// 等待处理完成
	<-stream.Done

	// 验证订单数量
	expectedOrderCount := 4
	if len(orders) != expectedOrderCount {
		t.Errorf("期望订单数量 %d, 实际数量 %d", expectedOrderCount, len(orders))
	}
}

// TestLargeScaleAuctionMatch 压力测试
func TestLargeScaleAuctionMatch(t *testing.T) {
	tests := []struct {
		name      string
		numOrders int
	}{
		{"1000个订单测试", 1000},
		{"100000个订单测试", 100000},
		{"1000000个订单测试", 1000000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 记录初始内存状态
			var initialMemStats, finalMemStats runtime.MemStats
			runtime.ReadMemStats(&initialMemStats)

			// 生成测试数据
			inputContent := generateLargeTestData(tt.numOrders)

			// 创建临时测试文件
			tmpDir := t.TempDir()
			inputFile := filepath.Join(tmpDir, "large_input.csv")
			err := os.WriteFile(inputFile, []byte(inputContent), 0644)
			if err != nil {
				t.Fatalf("创建大规模测试文件失败: %v", err)
			}

			// 记录开始时间
			start := time.Now()

			// 启动流式处理
			stream := order.StreamOrders(inputFile)
			// 创建合适的处理器
			processor := order.NewOrderProcessor(runtime.NumCPU())

			// 处理错误
			go func() {
				for err := range stream.Error {
					fmt.Println(err)
				}
			}()

			// 等待所有数据处理完成
			results := processor.Process(stream)
			<-stream.Done
			t.Logf("处理结果长度: %v", len(results))

			// 记录最终内存状态
			runtime.ReadMemStats(&finalMemStats)

			duration := time.Since(start)
			allocatedMemory := float64(finalMemStats.TotalAlloc-initialMemStats.TotalAlloc) / 1024 / 1024
			t.Logf("处理 %d 个订单耗时: %v", tt.numOrders, duration)
			t.Logf("内存分配: %.2f MB", allocatedMemory)
		})
	}
}

func TestExample(t *testing.T) {
	testFiles := []string{
		"input/example1_normal_small.csv",
		"input/example2_new_decision_algorithm_in_email.csv",
		"input/example3_only_sells.csv",
	}

	testOutputs := []string{
		"output/example1_output.csv",
		"output/example2_output.csv",
		"output/example3_output.csv",
	}

	for index, tt := range testFiles {
		t.Run(tt, func(t *testing.T) {
			// 捕获标准输出
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// 运行主程序
			os.Args = []string{"cmd", tt}
			main()

			// 恢复标准输出
			w.Close()
			os.Stdout = oldStdout
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			// 比较输出和预期输出文件内容是否一致
			expectedOutput, _ := os.ReadFile(testOutputs[index])
			if output != string(expectedOutput) {
				t.Errorf("输出不匹配")
			}
		})
	}
}
