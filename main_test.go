package main

import (
	"testing"

	// Adjust the import path based on your actual module structure for the src package
	// Example: "github.com/YourOrg/YourRepo/src"
	// Using the path from previous steps:
	"github.com/ConsulScale/src"
	"github.com/stretchr/testify/assert"
)

func TestMainFunctionCallsExecuteGoConsulKV(t *testing.T) {
	// 1. Save the original function from the src package
	// This variable (ExecuteGoConsulKVFunc) was created in src/executor.go for testability.
	originalExecuteGoConsulKVFunc := src.ExecuteGoConsulKVFunc

	// 2. Setup a flag and the mock function
	var executeCalled bool
	src.ExecuteGoConsulKVFunc = func() {
		executeCalled = true
	}

	// 3. Defer restoration of the original function
	// This ensures it's restored even if the test panics or t.Fatal is called.
	defer func() {
		src.ExecuteGoConsulKVFunc = originalExecuteGoConsulKVFunc
	}()

	// 4. Call the main() function
	// Since this test is in `package main`, we can call `main()` directly.
	main()

	// 5. Assert that the mock function was called
	assert.True(t, executeCalled, "src.ExecuteGoConsulKVFunc should have been called by main()")
}
