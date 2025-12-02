package vm

import (
	"alna-lang/internal/opcode"
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TuiDebuggerModel struct {
	VM                *VM
	width             int
	height            int
	paused            bool
	bytecodeViewport  viewport.Model
	constantsViewport viewport.Model
	stackViewport     viewport.Model
	variablesViewport viewport.Model
}

func TeaInitialModel(vm *VM) TuiDebuggerModel {
	return TuiDebuggerModel{
		VM:     vm,
		paused: true,
	}
}

func (m TuiDebuggerModel) Init() tea.Cmd {
	return nil
}

func (m TuiDebuggerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "n":
			m.VM.Step()
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Account for borders (2 lines per border) and top padding (1 line)
		viewportHeight := (msg.Height - 7) / 2
		m.bytecodeViewport = viewport.New(msg.Width/2-4, viewportHeight)
		m.constantsViewport = viewport.New(msg.Width/2-4, viewportHeight)
		m.stackViewport = viewport.New(msg.Width/2-4, viewportHeight)
		m.variablesViewport = viewport.New(msg.Width/2-4, viewportHeight)
	}

	return m, nil
}

func (m TuiDebuggerModel) View() string {
	m.bytecodeViewport.SetContent(m.VM.renderBytecodeView())
	m.constantsViewport.SetContent(m.VM.renderConstantsView())
	m.stackViewport.SetContent(m.VM.renderStackView())
	m.variablesViewport.SetContent(m.VM.renderVariablesView())

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	bytecodeBox := borderStyle.Render(m.bytecodeViewport.View())
	constantsBox := borderStyle.Render(m.constantsViewport.View())
	stackBox := borderStyle.Render(m.stackViewport.View())
	variablesBox := borderStyle.Render(m.variablesViewport.View())

	leftSide := lipgloss.JoinVertical(lipgloss.Top, bytecodeBox, constantsBox)
	rightSide := lipgloss.JoinVertical(lipgloss.Top, stackBox, variablesBox)

	content := lipgloss.JoinHorizontal(lipgloss.Top, leftSide, rightSide)

	// Add top padding to account for tmux header
	return lipgloss.NewStyle().PaddingTop(1).Render(content)
}

func (vm *VM) StartTuiDebugger() error {
	p := tea.NewProgram(
		TeaInitialModel(vm),
		tea.WithAltScreen(),       // Use alternate screen to prevent output interference
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to start TUI debugger: %w", err)
	}

	return nil
}

func (vm *VM) renderStackView() string {
	var stackContent string
	for i := len(vm.stack) - 1; i >= 0; i-- {
		stackContent += fmt.Sprintf("%d: %v\n", i, vm.stack[i])
	}
	return stackContent
}

func (vm *VM) renderBytecodeView() string {
	var bytecodeContent string

	// Start from PcOffset (where instructions begin after header and constants)
	pos := vm.PcOffset

	for pos < len(vm.program) {
		indicator := "  "
		if pos == vm.Pc {
			indicator = "> "
		}

		op := opcode.Opcode(vm.program[pos])
		instruction := fmt.Sprintf("%s%04d: %s", indicator, pos, op.String())

		// Check if this opcode has an operand
		if op.HasOperand() && pos+1 < len(vm.program) {
			operand := int(vm.program[pos+1])
			instruction += fmt.Sprintf(" %d", operand)

			// Add helpful comments for certain opcodes
			if op == opcode.LOAD_CONST && operand < len(vm.constants) {
				instruction += fmt.Sprintf("  ; %v", vm.constants[operand])
			} else if op == opcode.LOAD_VAR || op == opcode.STORE_VAR {
				if operand < len(vm.Variables) {
					instruction += fmt.Sprintf("  ; %v", vm.Variables[operand])
				}
			}

			pos += 2 // opcode + operand
		} else {
			pos++ // just opcode
		}

		bytecodeContent += instruction + "\n"
	}

	return bytecodeContent
}

func (vm *VM) renderConstantsView() string {
	var constantsContent string
	for i, c := range vm.constants {
		constantsContent += fmt.Sprintf("%d: %v\n", i, c)
	}
	return constantsContent
}

func (vm *VM) renderVariablesView() string {
	var variablesContent string
	if len(vm.Variables) == 0 {
		variablesContent = "No variables yet\n"
	} else {
		for i, v := range vm.Variables {
			variablesContent += fmt.Sprintf("%d: %v\n", i, v)
		}
	}
	return variablesContent
}
