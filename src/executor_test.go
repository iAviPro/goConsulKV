package src

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock implementations for consulwrapper functions and helpers
var (
	mockAddKVToConsulCalled            bool
	mockAddKVToConsulArgs              []string
	mockDeleteKVFromConsulCalled       bool
	mockDeleteKVFromConsulArgs         []string
	mockBackupConsulKVCalled           bool
	mockBackupConsulKVArgs             []string
	mockRestoreConsulKVCalled          bool
	mockRestoreConsulKVArgs            []string
	mockSyncConsulKVStoreCalled        bool
	mockSyncConsulKVStoreArgs          []string
	mockExitCalled                     bool
	mockExitCode                       int
	mockCreateConsulDetailsCalled      bool
	mockCreateConsulDetailsArgs        string
	// ... add other mock trackers if needed for other helper functions
)

func resetGlobalMocks() {
	mockAddKVToConsulCalled = false
	mockAddKVToConsulArgs = nil
	mockDeleteKVFromConsulCalled = false
	mockDeleteKVFromConsulArgs = nil
	mockBackupConsulKVCalled = false
	mockBackupConsulKVArgs = nil
	mockRestoreConsulKVCalled = false
	mockRestoreConsulKVArgs = nil
	mockSyncConsulKVStoreCalled = false
	mockSyncConsulKVStoreArgs = nil
	mockExitCalled = false
	mockExitCode = 0
	mockCreateConsulDetailsCalled = false
	mockCreateConsulDetailsArgs = ""
}

// captureOutput helper
func captureStdOutput(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// TestMain to set up and tear down mocks for package-level variables
func TestMain(m *testing.M) {
	// Save original functions (no src. prefix as executor_test.go is in package src)
	origAddKVToConsul := AddKVToConsulFunc
	origDeleteKVFromConsul := DeleteKVFromConsulFunc
	origBackupConsulKV := BackupConsulKVFunc
	origRestoreConsulKV := RestoreConsulKVFunc
	origSyncConsulKVStore := SyncConsulKVStoreFunc
	origHelpersExitFunc := ExitFunc // from helpers.go (package src)

	// Setup Mocks (no src. prefix)
	AddKVToConsulFunc = func(sn, cn, props, config, replace string) {
		mockAddKVToConsulCalled = true
		mockAddKVToConsulArgs = []string{sn, cn, props, config, replace}
	}
	DeleteKVFromConsulFunc = func(sn, cn, props, config string) {
		mockDeleteKVFromConsulCalled = true
		mockDeleteKVFromConsulArgs = []string{sn, cn, props, config}
	}
	BackupConsulKVFunc = func(cn, cp, fp string) {
		mockBackupConsulKVCalled = true
		mockBackupConsulKVArgs = []string{cn, cp, fp}
	}
	RestoreConsulKVFunc = func(cn, cp, fp, sn string) {
		mockRestoreConsulKVCalled = true
		mockRestoreConsulKVArgs = []string{cn, cp, fp, sn}
	}
	SyncConsulKVStoreFunc = func(source, target, sn, configName, replace string) {
		mockSyncConsulKVStoreCalled = true
		mockSyncConsulKVStoreArgs = []string{source, target, sn, configName, replace}
	}
	ExitFunc = func(code int) { // Mock for helpers.ExitFunc (package src)
		mockExitCalled = true
		mockExitCode = code
		// Using panic to halt execution in the test, similar to os.Exit
		// This helps ensure assertions after an expected exit are not missed if the code continued.
		panic("exit_called_with_" + string(rune(code+'0')))
	}

	// Run tests
	exitVal := m.Run()

	// Restore original functions (no src. prefix)
	AddKVToConsulFunc = origAddKVToConsul
	DeleteKVFromConsulFunc = origDeleteKVFromConsul
	BackupConsulKVFunc = origBackupConsulKV
	RestoreConsulKVFunc = origRestoreConsulKV
	SyncConsulKVStoreFunc = origSyncConsulKVStore
	ExitFunc = origHelpersExitFunc

	os.Exit(exitVal)
}

func TestExecuteGoConsulKV(t *testing.T) {
	originalOsArgs := os.Args // Save original os.Args
	defer func() { os.Args = originalOsArgs }() // Restore os.Args after all tests

	// --- Test No Command ---
	t.Run("NoCommand", func(t *testing.T) {
		resetGlobalMocks()
		os.Args = []string{"cmd"} // Simulates calling with no command
		output := captureStdOutput(func() {
			ExecuteGoConsulKV()
		})
		assert.Contains(t, output, "Usage: goConsulKV [command]")
		assert.Contains(t, output, "Available commands are:")
		assert.False(t, mockExitCalled)
	})

	// --- Test Invalid Command ---
	t.Run("InvalidCommand", func(t *testing.T) {
		resetGlobalMocks()
		os.Args = []string{"cmd", "unknown"}
		output := captureStdOutput(func() {
			ExecuteGoConsulKV()
		})
		assert.Contains(t, output, "Usage: goConsulKV [command]")
		assert.Contains(t, output, "Unknown command: unknown")
		assert.False(t, mockExitCalled)
	})

	// --- Test "add" command ---
	t.Run("AddCommand", func(t *testing.T) {
		tests := []struct {
			name          string
			args          []string
			expectExit    bool
			exitCode      int
			expectCall    bool
			expectedArgs  []string
			expectedPanic string
			stdOutCheck   func(t *testing.T, output string)
		}{
			{"Success", []string{"cmd", "add", "-s", "s1", "-p", "k=v", "-config", "c.yml", "-n", "n1", "-replace", "true"}, false, 0, true, []string{"s1", "n1", "k=v", "c.yml", "true"}, "", nil},
			{"DefaultReplace", []string{"cmd", "add", "-s", "s1", "-p", "k=v"}, false, 0, true, []string{"s1", "", "k=v", "", "false"}, "", nil},
			{"MissingServiceName", []string{"cmd", "add", "-p", "k=v"}, true, 1, false, nil, "exit_called_with_1", func(t *testing.T, out string) { assert.Contains(t, out, "Missing critical arguments for command: add") }},
			{"MissingProperties", []string{"cmd", "add", "-s", "s1"}, true, 1, false, nil, "exit_called_with_1", func(t *testing.T, out string) { assert.Contains(t, out, "Missing critical arguments for command: add") }},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resetGlobalMocks()
				os.Args = tt.args
				
				runFunc := func() { ExecuteGoConsulKV() }
				
				if tt.expectPanic != "" {
					var output string
					assert.PanicsWithValue(t, tt.expectPanic, func() {
						output = captureStdOutput(runFunc)
					}, "Expected panic with value "+tt.expectPanic)
					if tt.stdOutCheck != nil { // Check output even if panic (e.g. if Print then Exit)
						tt.stdOutCheck(t, output)
					}
				} else {
					output := captureStdOutput(runFunc)
					if tt.stdOutCheck != nil {
						tt.stdOutCheck(t, output)
					}
				}

				assert.Equal(t, tt.expectExit, mockExitCalled, "mockExitCalled mismatch")
				if tt.expectExit {
					assert.Equal(t, tt.exitCode, mockExitCode, "mockExitCode mismatch")
				}
				assert.Equal(t, tt.expectCall, mockAddKVToConsulCalled, "mockAddKVToConsulCalled mismatch")
				if tt.expectCall {
					assert.Equal(t, tt.expectedArgs, mockAddKVToConsulArgs, "mockAddKVToConsulArgs mismatch")
				}
			})
		}
	})

	// --- Test "delete" command ---
	t.Run("DeleteCommand", func(t *testing.T) {
		tests := []struct {
			name          string
			args          []string
			expectExit    bool
			exitCode      int
			expectCall    bool
			expectedArgs  []string
			expectedPanic string
			stdOutCheck   func(t *testing.T, output string)
		}{
			{"Success", []string{"cmd", "delete", "-s", "s1", "-p", "k1", "-config", "c.yml", "-n", "n1"}, false, 0, true, []string{"s1", "n1", "k1", "c.yml"}, "", nil},
			{"MissingServiceName", []string{"cmd", "delete", "-p", "k1"}, true, 1, false, nil, "exit_called_with_1", func(t *testing.T, out string) { assert.Contains(t, out, "Missing critical arguments for command: delete") }},
			{"MissingProperties", []string{"cmd", "delete", "-s", "s1"}, true, 1, false, nil, "exit_called_with_1", func(t *testing.T, out string) { assert.Contains(t, out, "Missing critical arguments for command: delete") }},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resetGlobalMocks()
				os.Args = tt.args
				runFunc := func() { ExecuteGoConsulKV() }

				if tt.expectedPanic != "" {
					var output string
					assert.PanicsWithValue(t, tt.expectedPanic, func() {
						output = captureStdOutput(runFunc)
					}, "Expected panic with value "+tt.expectedPanic)
					if tt.stdOutCheck != nil { tt.stdOutCheck(t, output) }
				} else {
					output := captureStdOutput(runFunc)
					if tt.stdOutCheck != nil { tt.stdOutCheck(t, output) }
				}
				
				assert.Equal(t, tt.expectExit, mockExitCalled)
				if tt.expectExit { assert.Equal(t, tt.exitCode, mockExitCode) }
				assert.Equal(t, tt.expectCall, mockDeleteKVFromConsulCalled)
				if tt.expectCall { assert.Equal(t, tt.expectedArgs, mockDeleteKVFromConsulArgs) }
			})
		}
	})

	// --- Test "backup" command ---
	t.Run("BackupCommand", func(t *testing.T) {
		tests := []struct {
			name         string
			args         []string
			expectCall   bool
			expectedArgs []string
			stdOutChecks []string // Substrings to check in stdout
		}{
			{"Success_AllArgs", []string{"cmd", "backup", "-n", "n1", "-config", "c.yml", "-save", "/p"}, true, []string{"n1", "c.yml", "/p"}, nil},
			{"Success_DefaultArgs", []string{"cmd", "backup"}, true, []string{"", "", ""}, []string{"Consul Env Config File Path not specified.", "Consul Name not specified.", "Backup file path not specified."}},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resetGlobalMocks()
				os.Args = tt.args
				output := captureStdOutput(func() { ExecuteGoConsulKV() })

				assert.False(t, mockExitCalled) // Backup has no mandatory args that cause exit
				assert.Equal(t, tt.expectCall, mockBackupConsulKVCalled)
				if tt.expectCall { assert.Equal(t, tt.expectedArgs, mockBackupConsulKVArgs) }
				for _, check := range tt.stdOutChecks {
					assert.Contains(t, output, check)
				}
			})
		}
	})

	// --- Test "restore" command ---
	t.Run("RestoreCommand", func(t *testing.T) {
		tests := []struct {
			name          string
			args          []string
			expectExit    bool
			exitCode      int
			expectCall    bool
			expectedArgs  []string
			expectedPanic string
			stdOutCheck   func(t *testing.T, output string)
		}{
			{"Success", []string{"cmd", "restore", "-n", "n1", "-config", "c.yml", "-s", "s1", "-file", "f.json"}, false, 0, true, []string{"n1", "c.yml", "f.json", "s1"},"", nil},
			{"MissingConsulName", []string{"cmd", "restore", "-file", "f.json"}, true, 1, false, nil, "exit_called_with_1", func(t *testing.T, out string){ assert.Contains(t, out, "Missing critical arguments for command: restore") }},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resetGlobalMocks()
				os.Args = tt.args
				runFunc := func() { ExecuteGoConsulKV() }

				if tt.expectedPanic != "" {
					var output string
					assert.PanicsWithValue(t, tt.expectedPanic, func() {
						output = captureStdOutput(runFunc)
					}, "Expected panic with value "+tt.expectedPanic)
					if tt.stdOutCheck != nil { tt.stdOutCheck(t, output) }
				} else {
					output := captureStdOutput(runFunc)
					if tt.stdOutCheck != nil { tt.stdOutCheck(t, output) }
				}

				assert.Equal(t, tt.expectExit, mockExitCalled)
				if tt.expectExit { assert.Equal(t, tt.exitCode, mockExitCode) }
				assert.Equal(t, tt.expectCall, mockRestoreConsulKVCalled)
				if tt.expectCall { assert.Equal(t, tt.expectedArgs, mockRestoreConsulKVArgs) }
			})
		}
	})
	
	// --- Test "sync" command ---
	t.Run("SyncCommand", func(t *testing.T) {
		tests := []struct {
			name          string
			args          []string
			expectExit    bool
			exitCode      int
			expectCall    bool
			expectedArgs  []string
			expectedPanic string
			stdOutCheck   func(t *testing.T, output string)
		}{
			{"Success", []string{"cmd", "sync", "-source", "src", "-target", "tgt", "-config", "c.yml", "-s", "s1", "-replace", "true"}, false, 0, true, []string{"src", "tgt", "s1", "c.yml", "true"},"", nil},
			{"MissingSource", []string{"cmd", "sync", "-target", "tgt"}, true, 1, false, nil, "exit_called_with_1", func(t *testing.T, out string){ assert.Contains(t, out, "Missing critical arguments for command: sync")}},
			{"MissingTarget", []string{"cmd", "sync", "-source", "src"}, true, 1, false, nil, "exit_called_with_1", func(t *testing.T, out string){ assert.Contains(t, out, "Missing critical arguments for command: sync")}},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resetGlobalMocks()
				os.Args = tt.args
				runFunc := func() { ExecuteGoConsulKV() }

				if tt.expectedPanic != "" {
					var output string
					assert.PanicsWithValue(t, tt.expectedPanic, func() {
						output = captureStdOutput(runFunc)
					}, "Expected panic with value "+tt.expectedPanic)
					if tt.stdOutCheck != nil { tt.stdOutCheck(t, output) }
				} else {
					output := captureStdOutput(runFunc)
					if tt.stdOutCheck != nil { tt.stdOutCheck(t, output) }
				}

				assert.Equal(t, tt.expectExit, mockExitCalled)
				if tt.expectExit { assert.Equal(t, tt.exitCode, mockExitCode) }
				assert.Equal(t, tt.expectCall, mockSyncConsulKVStoreCalled)
				if tt.expectCall { assert.Equal(t, tt.expectedArgs, mockSyncConsulKVStoreArgs) }
			})
		}
	})
}
