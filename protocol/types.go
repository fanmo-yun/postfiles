package protocol

type DataType uint8

const (
	FileMeta DataType = iota
	FileQuantity
	EndOfTransmission
	Confirm
)
