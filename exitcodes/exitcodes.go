package exitcodes

const (
	ErrServer = iota + 1001
	ErrClient
	ErrFlag
	ErrFileStat
	ErrDirStat
	ErrIPAndPort
	ErrNotTerminal
	ErrReadInput
)
