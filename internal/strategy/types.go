package strategy

type StrategyType int

const (
	RoundRobinStrategy = iota
	UnknownStrategy
)

func (s StrategyType) String() string {
	switch s {
	case RoundRobinStrategy:
		return "rb"
	default:
		return "unknown"
	}
}

func ParseStrategyType(name string) StrategyType {
	switch name {
	case "rb":
		return RoundRobinStrategy
	case "unknown":
		return UnknownStrategy
	default:
		return UnknownStrategy
	}
}
