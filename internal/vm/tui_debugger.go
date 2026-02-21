package vm

import (
	"alna-lang/internal/opcode"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TuiDebuggerModel struct {
	VM                *VM
	width             int
	height            int
	paused            bool
	stepCount         int
	showHelp          bool
	bytecodeViewport  viewport.Model
	sourceViewport    viewport.Model
	functionsViewport viewport.Model
	stackViewport     viewport.Model
	variablesViewport viewport.Model
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("63")).
			Padding(0, 1)

	pausedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Bold(true)

	runningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true)

	keyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("228"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	currentLineStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("62")).
				Bold(true)
)

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
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "n":
			m.VM.Step()
			m.stepCount++
			return m, nil
		case "h":
			m.showHelp = true
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		colWidth := msg.Width/2 - 4
		if colWidth < 20 {
			colWidth = 20
		}

		leftPanelCount := 3
		rightPanelCount := 2
		leftHeight := (msg.Height - 5 - leftPanelCount*3) / leftPanelCount
		rightHeight := (msg.Height - 5 - rightPanelCount*3) / rightPanelCount

		m.bytecodeViewport = viewport.New(colWidth, leftHeight)
		m.sourceViewport = viewport.New(colWidth, leftHeight)
		m.functionsViewport = viewport.New(colWidth, leftHeight)
		m.stackViewport = viewport.New(colWidth, rightHeight)
		m.variablesViewport = viewport.New(colWidth, rightHeight)
	}

	return m, nil
}

func (m TuiDebuggerModel) View() string {
	m.bytecodeViewport.SetContent(m.VM.renderBytecodeView())
	m.sourceViewport.SetContent(m.VM.renderSourceView())
	m.functionsViewport.SetContent(m.VM.renderFunctionsView())
	m.stackViewport.SetContent(m.VM.renderStackView())
	m.variablesViewport.SetContent(m.VM.renderVariablesView())

	colWidth := m.width/2 - 4
	if colWidth < 20 {
		colWidth = 20
	}

	bytecodeBox := m.renderTitledBox("Bytecode", m.bytecodeViewport.View(), colWidth)
	sourceBox := m.renderTitledBox("Source", m.sourceViewport.View(), colWidth)
	functionsBox := m.renderTitledBox("Functions", m.functionsViewport.View(), colWidth)
	stackBox := m.renderTitledBox("Stack", m.stackViewport.View(), colWidth)
	variablesBox := m.renderTitledBox("Variables", m.variablesViewport.View(), colWidth)

	leftSide := lipgloss.JoinVertical(lipgloss.Top, bytecodeBox, sourceBox, functionsBox)
	rightSide := lipgloss.JoinVertical(lipgloss.Top, stackBox, variablesBox)

	content := lipgloss.JoinHorizontal(lipgloss.Top, leftSide, rightSide)

	statusBar := m.renderStatusBar()

	mainView := lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.NewStyle().PaddingTop(1).Render(content),
		statusBar,
	)

	if m.showHelp {
		return m.renderHelpOverlay(mainView)
	}

	return mainView
}

func (m TuiDebuggerModel) renderTitledBox(title, content string, width int) string {
	titleText := titleStyle.Render(title)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("63")).
		Width(width)

	border := borderStyle.GetBorderStyle()
	topLeft := border.TopLeft
	topRight := border.TopRight
	left := border.Left
	right := border.Right
	bottomLeft := border.BottomLeft
	bottomRight := border.BottomRight
	horizontal := border.Top

	topBorder := fmt.Sprintf("%s %s ", topLeft, titleText)
	titleWidth := lipgloss.Width(topBorder)
	remainingWidth := width - titleWidth - 1
	if remainingWidth < 1 {
		remainingWidth = 1
	}
	topBorder += strings.Repeat(horizontal, remainingWidth) + topRight

	lines := strings.Split(content, "\n")
	var innerContent string
	contentWidth := width - 2
	for _, line := range lines {
		truncated := line
		if len(line) > contentWidth {
			truncated = line[:contentWidth]
		}
		innerContent += left + truncated + strings.Repeat(" ", contentWidth-len(truncated)) + right + "\n"
	}
	innerContent = strings.TrimSuffix(innerContent, "\n")

	bottomBorder := bottomLeft + strings.Repeat(horizontal, width-2) + bottomRight

	return topBorder + "\n" + innerContent + "\n" + bottomBorder
}

func (m TuiDebuggerModel) renderStatusBar() string {
	stateText := pausedStyle.Render("⏸ PAUSED")
	pcText := fmt.Sprintf("PC: %04d", m.VM.Pc)
	stepText := fmt.Sprintf("Step: %d", m.stepCount)

	keys := fmt.Sprintf("%s %s %s",
		keyStyle.Render("[n]ext"),
		keyStyle.Render("[q]uit"),
		keyStyle.Render("[h]elp"),
	)

	leftSide := fmt.Sprintf("%s │ %s │ %s", stateText, dimStyle.Render(pcText), dimStyle.Render(stepText))
	rightSide := keys

	spacing := m.width - lipgloss.Width(leftSide) - lipgloss.Width(rightSide) - 4
	if spacing < 1 {
		spacing = 1
	}

	statusContent := fmt.Sprintf("%s%s%s", leftSide, strings.Repeat(" ", spacing), rightSide)

	return statusBarStyle.Width(m.width).Render(statusContent)
}

func (m TuiDebuggerModel) renderHelpOverlay(underlying string) string {
	helpStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("86")).
		Padding(1, 2).
		Background(lipgloss.Color("235"))

	title := titleStyle.Render("Keyboard Shortcuts")
	divider := dimStyle.Render(strings.Repeat("─", 30))

	helpContent := fmt.Sprintf(`%s
%s

  %s  Execute next instruction
  %s  Toggle this help
  %s  Exit debugger

%s`,
		title,
		divider,
		keyStyle.Render("[n]"),
		keyStyle.Render("[h]"),
		keyStyle.Render("[q]"),
		dimStyle.Render("Press any key to close"),
	)

	helpBox := helpStyle.Render(helpContent)

	overlayStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	return overlayStyle.Render(helpBox)
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

	pos := vm.PcOffset

	for pos < len(vm.program) {
		indicator := "  "
		if pos == vm.Pc {
			indicator = "> "
		}

		op := opcode.Opcode(vm.program[pos])
		instruction := fmt.Sprintf("%s%04d: %s", indicator, pos, op.String())

		if op.HasOperand() && pos+1 < len(vm.program) {
			operand := int(vm.program[pos+1])
			instruction += fmt.Sprintf(" %d", operand)

			if op == opcode.LOAD_CONST && operand < len(vm.constants) {
				instruction += fmt.Sprintf("  ; %v", vm.constants[operand])
			} else if op == opcode.LOAD_VAR || op == opcode.STORE_VAR {
				relativePc := pos - vm.PcOffset
				if srcPos, ok := vm.SourceMap[relativePc]; ok && srcPos.VarName != "" {
					instruction += fmt.Sprintf("  ; %s", srcPos.VarName)
				} else if name, ok := vm.VariableNames[operand]; ok {
					instruction += fmt.Sprintf("  ; %s", name)
				} else if operand < len(vm.Variables) {
					instruction += fmt.Sprintf("  ; %v", vm.Variables[operand])
				}
			}

			pos += 2
		} else {
			pos++
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

func (vm *VM) renderSourceView() string {
	if vm.rawCode == nil || len(vm.rawCode) == 0 {
		return "No source available\n"
	}

	relativePc := vm.Pc - vm.PcOffset
	var currentPos *SourcePosition
	if pos, ok := vm.SourceMap[relativePc]; ok {
		currentPos = &pos
	}

	var sourceContent string
	for i, line := range vm.rawCode {
		lineNum := i + 1

		if currentPos != nil && i == currentPos.Line && currentPos.Column >= 0 && currentPos.EndColumn > currentPos.Column {
			col := currentPos.Column
			endCol := currentPos.EndColumn
			if endCol > len(line) {
				endCol = len(line)
			}
			if col > len(line) {
				col = len(line)
			}

			before := ""
			if col > 0 {
				before = line[:col]
			}
			highlighted := line[col:endCol]
			after := ""
			if endCol < len(line) {
				after = line[endCol:]
			}

			highlightedStyled := currentLineStyle.Render(highlighted)
			sourceContent += fmt.Sprintf("> %3d: %s%s%s\n", lineNum, before, highlightedStyled, after)
		} else if currentPos != nil && i == currentPos.Line {
			sourceContent += fmt.Sprintf("> %3d: %s\n", lineNum, line)
		} else {
			sourceContent += fmt.Sprintf("  %3d: %s\n", lineNum, line)
		}
	}
	return sourceContent
}

func (vm *VM) getVariableNamesAtCurrentPc() map[int]string {
	names := make(map[int]string)
	relativePc := vm.Pc - vm.PcOffset

	// Scan source map for LOAD_VAR/STORE_VAR instructions up to current PC
	for pc := 0; pc <= relativePc; pc++ {
		if srcPos, ok := vm.SourceMap[pc]; ok && srcPos.VarName != "" {
			// Get the operand (variable index) from the bytecode
			absPc := pc + vm.PcOffset
			if absPc+1 < len(vm.program) {
				op := opcode.Opcode(vm.program[absPc])
				if op == opcode.LOAD_VAR || op == opcode.STORE_VAR {
					varIdx := int(vm.program[absPc+1])
					names[varIdx] = srcPos.VarName
				}
			}
		}
	}

	// Also include global variable names
	for idx, name := range vm.VariableNames {
		if _, exists := names[idx]; !exists {
			names[idx] = name
		}
	}

	return names
}

func (vm *VM) renderVariablesView() string {
	var variablesContent string
	if len(vm.Variables) == 0 {
		variablesContent = "No variables yet\n"
	} else {
		varNames := vm.getVariableNamesAtCurrentPc()
		for i, v := range vm.Variables {
			if name, ok := varNames[i]; ok {
				variablesContent += fmt.Sprintf("%s: %v\n", name, v)
			} else {
				variablesContent += fmt.Sprintf("var[%d]: %v\n", i, v)
			}
		}
	}
	return variablesContent
}

func (vm *VM) renderFunctionsView() string {
	var functionsContent string
	if len(vm.Functions) == 0 {
		functionsContent = "No functions\n"
	} else {
		for i, fn := range vm.Functions {
			typeLabel := "compiled"
			if fn.Type == FunctionTypeBuiltin {
				typeLabel = "builtin"
			}
			functionsContent += fmt.Sprintf("%d: %s [%s]\n", i, fn.Name, typeLabel)
		}
	}
	return functionsContent
}
