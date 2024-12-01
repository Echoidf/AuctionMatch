package main

import (
	"AuctionMatch/order"
	"fmt"
	"os"
	"runtime"
	"strings"
)

func checkArgs() {
	if len(os.Args) == 2 && os.Args[1] == "-h" {
		printUsage()
		return
	}

	if len(os.Args) < 2 {
		fmt.Println("参数错误！使用 -h 查看帮助信息")
		return
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

// writeResults 将结果写入标准输出
func writeResults(results []order.ProcessResult, outputFile string) {
	var output strings.Builder

	// 处理空结果的情况
	if len(results) == 0 {
		return
	}

	// 构建输出字符串
	for _, item := range results {
		if item.Price == 0 {
			output.WriteString(fmt.Sprintf("%s,\n", item.InstrumentID))
		} else {
			// 与输入精度保持一致
			output.WriteString(fmt.Sprintf("%s,%.*f\n", item.InstrumentID, item.Scale, item.Price))
		}
	}

	// 写入输出
	outputStr := output.String()
	if outputFile == "" {
		fmt.Print(outputStr)
	} else {
		// 确保使用UTF-8编码并保持原有的换行符
		if err := os.WriteFile(outputFile, []byte(strings.ReplaceAll(outputStr, "\n", "\r\n")), 0644); err != nil {
			fmt.Printf("写入结果时发生错误: %v\n", err)
			os.Exit(1)
		}
	}
}

func main() {
	checkArgs()
	// 创建订单流
	stream := order.StreamOrders(os.Args[1])

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

	// 输出结果
	if len(os.Args) == 4 && os.Args[2] == "-o" {
		writeResults(results, os.Args[3])
	} else {
		writeResults(results, "")
	}
}
