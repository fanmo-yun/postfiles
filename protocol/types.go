package protocol

type DataType uint8

const (
	FileMeta DataType = iota + 101
	FileQuantity
	EndOfTransmission
	ConfirmAccept
	AcceptFile
	RejectFile
)
