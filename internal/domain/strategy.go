package domain

type Strategy string

const (
	StrategyBlue  Strategy = "blue"
	StrategyGreen Strategy = "green"
)

func (s Strategy) String() string {
	return string(s)
}
