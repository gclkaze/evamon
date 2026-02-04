package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

type ConsolePrinter struct {
	onVerboseMode bool
}

func NewConsolePrinter(verbose bool) *ConsolePrinter {
	return &ConsolePrinter{onVerboseMode: verbose}
}

func (p *ConsolePrinter) GetVerboseFlagPointer() *bool {
	return &p.onVerboseMode
}

func (p *ConsolePrinter) Info(msg string) {
	fmt.Println(msg)
}

func (p *ConsolePrinter) Warn(msg string) {
	warningColor := color.New(color.FgHiYellow, color.Bold).SprintFunc()
	fmt.Fprintln(os.Stderr, warningColor(msg))
}
func (p *ConsolePrinter) ErrorWithMessage(msg string, err error) {
	p.Error(fmt.Errorf("%s%s", msg, err.Error()))
}

func (p *ConsolePrinter) VerboseInfo(msg string) {
	if p.onVerboseMode {
		fmt.Println(msg)
	}
}
func (p *ConsolePrinter) VerboseWarn(msg string) {
	if p.onVerboseMode {
		warningColor := color.New(color.FgHiYellow, color.Bold).SprintFunc()
		fmt.Fprintln(os.Stderr, warningColor(msg))
	}
}

func (p *ConsolePrinter) Error(err error) {
	errorColor := color.New(color.FgHiRed, color.Bold).SprintFunc()
	theError := err.Error()

	if strings.Contains(theError, "because the target machine actively refused it") {
		theError = "Couldn't connect to the Module Repository Server..check your internet connection"
	}

	if strings.Contains(theError, "An existing connection was forcibly closed by the remote host.") {
		theError = "The Module Repository Server went away..check your internet connection"
	}
	fmt.Fprintln(os.Stderr, errorColor("Error: "+theError))
}

func (p *ConsolePrinter) Success(msg string) {
	successColor := color.New(color.FgHiGreen, color.Bold).SprintFunc()
	fmt.Fprintln(os.Stderr, successColor(msg))
}
