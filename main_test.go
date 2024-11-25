package main

import (
	"AuctionMatch/utils"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestFullAuctionMatchProcess 测试完整的撮合流程
func TestFullAuctionMatchProcess(t *testing.T) {
	tests := []struct {
		name           string
		inputContent   string
		expectedOutput string
	}{
		{
			name: "基本撮合测试",
			inputContent: `IF2412,0,3973.4,3
IF2412,1,3973.2,2
IF2412,0,3973.2,1`,
			expectedOutput: "IF2412,3973.4\n", // 最大成交量价格
		},
		{
			name: "多个合约测试",
			inputContent: `IF2412,0,3973.4,3
IF2306,0,3972.2,2
IF2412,1,3973.2,2
IF2306,1,3972.0,1`,
			expectedOutput: "IF2412,3973.4\nIF2306,3972.2\n", // 修正期望输出
		},
		{
			name: "无法撮合测试",
			inputContent: `IF2412,0,3970.0,3
IF2412,1,3975.0,2`,
			expectedOutput: "IF2412,\n",
		},
		{
			name: "相同最大成交量不同剩余量测试",
			inputContent: `IF2412,0,3973.4,5
IF2412,0,3973.2,3
IF2412,1,3973.0,4
IF2412,1,3973.2,3`,
			expectedOutput: "IF2412,3973.2\n",
		},
		{
			name:           "空文件测试",
			inputContent:   "",
			expectedOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建临时测试文件
			tmpDir := t.TempDir()
			inputFile := filepath.Join(tmpDir, "input.csv")
			err := os.WriteFile(inputFile, []byte(tt.inputContent), 0644)
			if err != nil {
				t.Fatalf("创建测试文件失败: %v", err)
			}

			// 捕获标准输出
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// 运行主程序
			os.Args = []string{"cmd", inputFile}
			main()

			// 恢复标准输出并获取输出内容
			w.Close()
			os.Stdout = oldStdout
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			// 比较结果
			if output != tt.expectedOutput {
				t.Errorf("期望输出 %q, 实际输出 %q", tt.expectedOutput, output)
			}
		})
	}
}

// TestCalculateAuctionPrice 测试价格计算逻辑
func TestCalculateAuctionPrice(t *testing.T) {
	tests := []struct {
		name          string
		orders        []Order
		expectedPrice float64
	}{
		{
			name: "正常撮合",
			orders: []Order{
				{InstrumentID: "IF2412", Direction: 0, Price: 3973.4, Volume: 3}, // 买单
				{InstrumentID: "IF2412", Direction: 1, Price: 3973.2, Volume: 2}, // 卖单
			},
			expectedPrice: 3973.4,
		},
		{
			name: "无买单",
			orders: []Order{
				{InstrumentID: "IF2412", Direction: 1, Price: 3973.2, Volume: 2},
			},
			expectedPrice: 0,
		},
		{
			name: "无卖单",
			orders: []Order{
				{InstrumentID: "IF2412", Direction: 0, Price: 3973.4, Volume: 3},
			},
			expectedPrice: 0,
		},
		{
			name: "价格不交叉",
			orders: []Order{
				{InstrumentID: "IF2412", Direction: 0, Price: 3970.0, Volume: 3},
				{InstrumentID: "IF2412", Direction: 1, Price: 3975.0, Volume: 2},
			},
			expectedPrice: 0,
		},
		{
			name: "多个价格档位测试",
			orders: []Order{
				{InstrumentID: "IF2412", Direction: 0, Price: 3973.4, Volume: 3}, // 买单
				{InstrumentID: "IF2412", Direction: 0, Price: 3973.2, Volume: 2}, // 买单
				{InstrumentID: "IF2412", Direction: 1, Price: 3973.0, Volume: 4}, // 卖单
			},
			expectedPrice: 3973.2,
		},
		{
			name: "相同成交量不同剩余量测试",
			orders: []Order{
				{InstrumentID: "IF2412", Direction: 0, Price: 3973.4, Volume: 3}, // 买单
				{InstrumentID: "IF2412", Direction: 1, Price: 3973.2, Volume: 2}, // 卖单
				{InstrumentID: "IF2412", Direction: 1, Price: 3973.0, Volume: 1}, // 卖单
			},
			expectedPrice: 3973.4, // 在最大成交量和最小剩余量相同时，选择最高价格
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price := calculateAuctionPrice(tt.orders)
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

	// 读取订单
	orders, instruments := readOrders(inputFile)

	// 验证订单数量
	expectedOrderCount := 4 // 不包括无效行
	if len(orders) != expectedOrderCount {
		t.Errorf("期望订单数量 %d, 实际数量 %d", expectedOrderCount, len(orders))
	}

	// 验证合约顺序
	expectedInstruments := []string{"IF2412", "IF2306"}
	if !stringSliceEqual(instruments, expectedInstruments) {
		t.Errorf("期望合约顺序 %v, 实际顺序 %v", expectedInstruments, instruments)
	}
}

// 辅助函数：比较字符串切片
func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
