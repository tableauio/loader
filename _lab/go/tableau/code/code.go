package code

type Code int

const (
	Success Code = iota
	NotFound
	Nil
	Unknown
)
