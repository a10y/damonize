// daemonize a passed binary
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %v [ARGS] COMMAND\n", os.Args[0])
	flag.PrintDefaults()
}

func parseCmdLine(rawArgs []string, cmd *string, args *[]string) {
	if len(rawArgs) == 0 {
		// print usage
		fmt.Fprintf(os.Stderr, "error: not enough arguments\n")
		flag.Usage()
		os.Exit(1)
	}
	*cmd = rawArgs[0]
	*args = rawArgs
}

var (
	clearEnv bool
	cmd      string
	args     []string
)

func init() {
	flag.BoolVar(&clearEnv, "x", false, "execute the daemon with an empty environment")
	flag.Usage = usage
	flag.Parse()
	parseCmdLine(flag.Args(), &cmd, &args)
}

func main() {
	// Ensure the binary specified can be found on the current PATH
	qualified, err := exec.LookPath(cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "damonize: %v not found on PATH\n", cmd)
		os.Exit(1)
	}

	// Copied from exec_bsd.go in runtime package
	darwin := runtime.GOOS == "darwin"
	pid, ischild, _ := syscall.RawSyscall(syscall.SYS_FORK, 0, 0, 0)
	if darwin && ischild == 1 {
		pid = 0
	}

	if pid > 0 {
		// parent, dies
		os.Exit(0)
	} else {

		// child, create new session and fork subprocess
		syscall.Setsid()
		pid, ischild, _ = syscall.RawSyscall(syscall.SYS_FORK, 0, 0, 0)
		if darwin && ischild == 1 {
			pid = 0
		}

		if pid > 0 {
			// kill the parent
			os.Exit(0)
		} else {
			// Execute the daemon in the new session
			if clearEnv {
				syscall.Exec(qualified, args, []string{})
			} else {
				syscall.Exec(qualified, args, os.Environ())
			}
		}
	}
}
