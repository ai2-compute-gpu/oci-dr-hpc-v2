package level1_tests

import (
	"fmt"
	"os/exec"
	"strings"
	
	"github.com/oracle/oci-dr-hpc-v2/internal/logger"
)

func RunPCIeErrorCheck() error {
	logger.Info("=== PCIe Error Check ===")
	
	// Check for PCIe errors in dmesg
	cmd := exec.Command("sudo", "dmesg", "-T")
	output, err := cmd.Output()
	if err != nil {
		logger.Error("Failed to run dmesg (may require sudo):", err)
		logger.Info("PCIe Error Check: SKIP - dmesg requires elevated privileges")
		return nil
	}
	
	outputStr := string(output)
	pciErrors := []string{
		"PCIe Bus Error",
		"AER:",
		"correctable error",
		"uncorrectable error",
		"fatal error",
	}
	
	errorCount := 0
	for _, errorPattern := range pciErrors {
		if strings.Contains(strings.ToLower(outputStr), strings.ToLower(errorPattern)) {
			errorCount++
			logger.Error(fmt.Sprintf("Found PCIe error pattern: %s", errorPattern))
		}
	}
	
	if errorCount > 0 {
		logger.Error(fmt.Sprintf("PCIe Error Check: FAIL - Found %d error patterns", errorCount))
		return fmt.Errorf("found %d PCIe error patterns", errorCount)
	}
	
	logger.Info("PCIe Error Check: PASS - No PCIe errors found")
	return nil
}
