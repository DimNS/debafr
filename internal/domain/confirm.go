package domain

type Confirm string

const (
	ConfirmYes Confirm = "yes"
	ConfirmNo  Confirm = "no"
)

func (s Confirm) String() string {
	return string(s)
}
