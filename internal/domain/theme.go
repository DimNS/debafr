package domain

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	ColorWhite  lipgloss.Color
	ColorGray   lipgloss.Color
	ColorGreen  lipgloss.Color
	ColorRed    lipgloss.Color
	ColorYellow lipgloss.Color
	ColorOrange lipgloss.Color

	StyleGreen lipgloss.Style
	StyleRed   lipgloss.Style

	TextPressEnterToContinue string
}

func NewTheme() *Theme {
	colorWhite := lipgloss.Color("#FFFFFF")
	colorGray := lipgloss.Color("#666666")
	colorGreen := lipgloss.Color("#16AE71")
	colorRed := lipgloss.Color("#FB1C3C")
	colorYellow := lipgloss.Color("#ECBC64")
	colorOrange := lipgloss.Color("#E95420")

	return &Theme{
		ColorWhite:  colorWhite,
		ColorGray:   colorGray,
		ColorGreen:  colorGreen,
		ColorRed:    colorRed,
		ColorYellow: colorYellow,
		ColorOrange: colorOrange,

		StyleGreen: lipgloss.NewStyle().Foreground(colorGreen),
		StyleRed:   lipgloss.NewStyle().Foreground(colorRed),

		TextPressEnterToContinue: lipgloss.NewStyle().Foreground(colorGray).Render("Press enter to continue"),
	}
}
