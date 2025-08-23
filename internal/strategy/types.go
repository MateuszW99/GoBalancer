package strategy

type StrategyType int

const (
	UnknownStrategy    = iota
	RoundRobinStrategy = iota
	LeastConnections   = iota
)

func (s StrategyType) String() string {
	switch s {
	case RoundRobinStrategy:
		return "rb"
	case LeastConnections:
		return "lc"
	default:
		return "unknown"
	}
}

func ParseStrategyType(name string) StrategyType {
	switch name {
	case "rb":
		return RoundRobinStrategy
	case "lc":
		return LeastConnections
	case "unknown":
		return UnknownStrategy
	default:
		return UnknownStrategy
	}
}
