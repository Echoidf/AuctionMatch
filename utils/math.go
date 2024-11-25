package utils

import "math"

const EPSILON = 1e-10

// 定义一个约束，限制类型为数值类型
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64
}

// FindMin 查找切片中的最小值
func FindMin[T Number](nums []T) T {
	if len(nums) == 0 {
		var zero T
		return zero // 返回对应类型的零值
	}

	min := nums[0]
	for _, num := range nums[1:] {
		if num < min {
			min = num
		}
	}
	return min
}

// FindMax 查找切片中的最大值
func FindMax[T Number](nums []T) T {
	if len(nums) == 0 {
		var zero T
		return zero
	}

	max := nums[0]
	for _, num := range nums[1:] {
		if num > max {
			max = num
		}
	}
	return max
}

// Min 返回两个数中的最小值
func Min[T Number](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Max 返回两个数中的最大值
func Max[T Number](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Abs 返回绝对值
func Abs[T Number](x T) T {
	if x < 0 {
		return -x
	}
	return x
}

// 添加浮点数比较的辅助函数
func FloatEquals(a, b float64) bool {
	return math.Abs(a-b) < EPSILON
}

func FloatLessEqual(a, b float64) bool {
	return a < b || FloatEquals(a, b)
}

func FloatGreaterEqual(a, b float64) bool {
	return a > b || FloatEquals(a, b)
}
