package order

import "testing"

func TestGetTick(t *testing.T) {
	// 定义测试用例结构体
	type test struct {
		input string
		want  float32
	}

	// 初始化测试用例
	tests := []test{
		{input: "IF2306", want: 0.2},
		{input: "IH2306", want: 0.2},
		{input: "IC2306", want: 0.2},
		{input: "TS2306", want: 0.002},
	}

	// 运行测试用例
	for _, tt := range tests {
		order := &Order{InstrumentID: tt.input}
		if got := order.GetTick(); got != tt.want {
			t.Errorf("GetTick() = %v, want %v", got, tt.want)
		}
	}
}
