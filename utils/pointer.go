package utils

// Ptr 返回传入值的指针。
func Ptr[T any](value T) *T {
	return &value
}
