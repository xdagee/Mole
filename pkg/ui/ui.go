package ui

import "fmt"

var (
	IconArrow   = "->"
	IconSuccess = "[OK]"
	IconWarning = "[!]"
	IconClean   = "[C]"
	IconTrash   = "[T]"
	IconFolder  = "[F]"
	IconList    = "[L]"
)

func Banner() string {
	return "Mole"
}

type Style struct{}
func (s Style) Render(str ...string) string {
	if len(str) > 0 {
		return str[0]
	}
	return ""
}

var CyanStyle = Style{}
var GrayStyle = Style{}

func FormatWarning(s string) string { return s }
func FormatInfo(s string) string { return s }

type MenuOption struct {
	Label       string
	Description string
	Icon        string
	Action      func() error
}

type Menu struct {}
func NewMenu(title string, options []MenuOption) *Menu { return &Menu{} }
func (m *Menu) Run() error { return nil }
func (m *Menu) View() string { return "" }

func FormatGray(s string) string {
	return s // Stub
}

func FormatBytes(b int64) string {
	return fmt.Sprintf("%d bytes", b) // Stub
}
