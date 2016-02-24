package main

import (
	"fmt"
	//"log"
	"os/exec"
)

func main() {

	//out, _ := exec.Command("C:/Windows/SysWOW64/cmd.exe", "/c", "C:/Program Files (x86)/Git/bin/sh.exe", "--login", "-i").Output()
	//out, err1 := exec.Command("C:/Windows/SysWOW64/cmd.exe", "/c").Output()
	//cmd := exec.Command("C:/Program Files (x86)/Git/bin/sh.exe", "--login", "-i")
	//cmd, err := exec.Command("C:/Windows/SysWOW64/cmd.exe", "dir").Output()
	//cmd, err := exec.Command("C:/Program Files (x86)/Git/bin/sh.exe", "ls").Output()
	//cmd, err := exec.Command("C:/Users/jaredwarren/bin/Sencha/Cmd/6.0.1.76/sencha.exe").Output()

	execGit([]string{"add", "."})
	execGit([]string{"commit", "-a", "-m", `TEST`})

	// Build All
	//execSencha([]string{"app", "build", "classic"})

	/*cmd := exec.Command("C:/Program Files (x86)/Git/bin/git.exe", "status")
	cmd.Dir = "C:/data/www/sandbox/ExtBuilder"
	out, err := cmd.Output()
	fmt.Println(":", err, string(out))*/

	// sencha
	/*cmd := exec.Command("C:/Users/jaredwarren/bin/Sencha/Cmd/6.0.1.76/sencha.exe", "app", "build", "classic")
	cmd.Dir = "C:/data/www/sandbox/ExtBuilder"
	fmt.Println(cmd.Dir)
	if out, err := cmd.Output(); err != nil{
		fmt.Println(":", err, out)
	}*/
}

func execGit(args []string) {
	cmd := exec.Command("C:/Program Files (x86)/Git/bin/git.exe", args...)
	cmd.Dir = "C:/data/www/sandbox/ExtBuilder"
	if out, err := cmd.Output(); err != nil {
		fmt.Println(":", err, string(out))
	} else {
		fmt.Println(string(out))
	}
}

func execSencha(args []string) {
	cmd := exec.Command("C:/Users/jaredwarren/bin/Sencha/Cmd/6.0.1.76/sencha.exe", args...)
	cmd.Dir = "C:/data/www/sandbox/ExtBuilder"
	if out, err := cmd.Output(); err != nil {
		fmt.Println(":", err, string(out))
	} else {
		fmt.Println(string(out))
	}
}
