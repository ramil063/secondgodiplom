package dialog

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// ClearScreen функция очистки экрана, применяется для удобства взаимодействия с пользователем
func ClearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// PressEnterToContinue функция задержки работы приложения, до тех пор пока пользователь не нажмет на клавишу Enter
func PressEnterToContinue() {
	fmt.Print("\nНажмите Enter для продолжения...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
