package config

const (
	ModePro Mode = iota
	ModeStg
	ModeDev
)

type Mode int

type UpdateHandler func()
