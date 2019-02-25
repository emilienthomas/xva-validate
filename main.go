package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/emilienthomas/xva-validate/xva"
)

// Will be set by goreleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Tests integrity of the xva file passed through --xva parameter.
// Verbosity levels are:
// * 0: Only prints errors and "xva file is invalid" when needed
// * 1: Also prints "xva file is valid" when needed
// * 2: Prints each individual validation, this might create a lot of output.
func main() {
	// Program parameters
	xvaPtr := flag.String("xva", "backup.xva", "xva file")
	verbosityPtr := flag.Uint("verbose", 0, "Verbosity level")
	versionPtr := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *versionPtr {
		fmt.Printf("xva-validate %v, commit %v, built at %v", version, commit, date)
		os.Exit(0)
	}

	isValid, validationIssue, err := xva.Validate(*xvaPtr, *verbosityPtr)
	if err != nil {
		log.Println(fmt.Errorf("%v", err))
		os.Exit(2)
	}
	if !isValid {
		log.Println(fmt.Sprintf("xva file is invalid, reason: %s", validationIssue))
		os.Exit(1)
	} else {
		if *verbosityPtr >= 1 {
			log.Println("xva file is valid")
		}
		os.Exit(0)
	}
}
