package exitcodes

const (
	ErrServer = iota + 1001
	ErrClient
	ErrJsonEncoding
	ErrJsonUnmarshal
	ErrFlag
	ErrFileStat
	ErrDirStat
	ErrIPAndPort
	ErrNotTerminal
	ErrReadInput
)
