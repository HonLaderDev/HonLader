package utils

// Mod 返回非负取模结果。
func Mod(value, divisor int) int {
	result := value % divisor
	if result < 0 {
		result += divisor
	}
	return result
}

// ClampIndex 将索引约束到 [0, count-1] 范围内。
func ClampIndex(index, count int) int {
	return max(0, min(index, count-1))
}

// CeilQuotient 返回 n/d 向上取整后的结果。
func CeilQuotient(n, d int) int {
	if d <= 0 {
		return 0
	}
	return (n + d - 1) / d
}

// MinInt 返回两个整数中较小的一个。
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
