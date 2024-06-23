package utils

const unit = 1024 * 1024

func ToMB(size int64) float64 {
	return (float64(size) / unit)
}
