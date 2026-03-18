package domain

type Mode string

const (
	ModeUpdate  Mode = "update"
	ModeInstall Mode = "install"
)

func (s Mode) String() string {
	return string(s)
}
