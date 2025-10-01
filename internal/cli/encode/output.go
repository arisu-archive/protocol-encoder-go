package encode

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/arisu-archive/protocol-encoder-go/internal/emulator"
)

func outputResultsJSON(result *emulator.InvokeResponse) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func outputResultsText(result *emulator.InvokeResponse) error {
	fmt.Printf("Execution Results:\n")
	fmt.Printf("  Function Address: %#x\n", result.FunctionAddress)
	fmt.Printf("  Return Value:     %d (%#x)\n", result.ReturnValue, result.ReturnValue)
	fmt.Printf("  Execution Time:   %v\n", result.ExecutionTime)
	fmt.Printf("  Success:          %t\n", result.Success)
	if result.Error != "" {
		fmt.Printf("  Error:            %s\n", result.Error)
	}
	return nil
}
