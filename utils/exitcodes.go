package utils

const (
	ErrServer = iota + 1001
	ErrServerClose
	ErrClient
	ErrFlag
	ErrFileStat
	ErrDirStat
	ErrIPAndPort
	ErrNotTerminal
	ErrReadInput
)
