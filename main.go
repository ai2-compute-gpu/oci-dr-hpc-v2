// =============================================================================
// main.go - Application Entry Point
package main

import "github.com/oracle/oci-dr-hpc-v2/cmd"

// Version is set by the build system
var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
