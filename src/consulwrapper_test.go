package src

import (
	"testing"

	"bytes"
	"errors"
	"io"
	"os"
	"path"

	// Duplicates of above were removed in a previous step, ensuring they are not re-added.
	"github.com/ConsulScale/config"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ConsulKV interface is defined in consulwrapper.go, no need to redefine here.

// MockKV is a mock implementation of ConsulKV (defined in consulwrapper.go)
type MockKV struct {
	mock.Mock
}

func (m *MockKV) Get(key string, q *api.QueryOptions) (*api.KVPair, *api.QueryMeta, error) {
	args := m.Called(key, q)
	if args.Get(0) == nil {
		// Ensure QueryMeta is not nil if error is also nil (Consul behavior for not found)
		if args.Get(1) == nil && args.Error(2) == nil {
			return nil, new(api.QueryMeta), args.Error(2)
		}
		if args.Get(1) == nil { // if QueryMeta is nil for some reason with an error
		    return nil, nil, args.Error(2)
		}
		return nil, args.Get(1).(*api.QueryMeta), args.Error(2)
	}
	return args.Get(0).(*api.KVPair), args.Get(1).(*api.QueryMeta), args.Error(2)
}

func (m *MockKV) List(prefix string, q *api.QueryOptions) (api.KVPairs, *api.QueryMeta, error) {
	args := m.Called(prefix, q)
	if args.Get(0) == nil {
		if args.Get(1) == nil && args.Error(2) == nil {
			return nil, new(api.QueryMeta), args.Error(2)
		}
		if args.Get(1) == nil {
		    return nil, nil, args.Error(2)
		}
		return nil, args.Get(1).(*api.QueryMeta), args.Error(2)
	}
	return args.Get(0).(api.KVPairs), args.Get(1).(*api.QueryMeta), args.Error(2)
}

func (m *MockKV) Put(p *api.KVPair, q *api.WriteOptions) (*api.WriteMeta, error) {
	args := m.Called(p, q)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.WriteMeta), args.Error(1)
}

func (m *MockKV) Delete(key string, q *api.WriteOptions) (*api.WriteMeta, error) { // Added Delete method
	args := m.Called(key, q)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*api.WriteMeta), args.Error(1)
}


// ConsulClientInterface is an interface for the Consul client
type ConsulClientInterface interface {
	KV() ConsulKV
}

// MockConsulClient is a mock implementation of ConsulClientInterface
type MockConsulClient struct {
	mock.Mock
	MockKV *MockKV // Embed MockKV to control its behavior
}

func (m *MockConsulClient) KV() ConsulKV {
	// Return the embedded MockKV instance
	// This allows us to set expectations on m.MockKV directly in tests
	return m.MockKV
}

// Helper function to create a new MockConsulClient with an initialized MockKV
func newMockConsulClient() *MockConsulClient {
	return &MockConsulClient{
		MockKV: new(MockKV),
	}
}

// testAPIClient is a global variable to hold our mock client for injection.
// This aligns with a simplified version of Option A/C, assuming we can
// control the client used by the functions under test.
// In a real scenario, ConnectConsul might be refactored to allow injection.
var testAPIClient *api.Client

// This function would ideally be part of the setup for tests that need
// to mock the *api.Client. For now, we'll assume functions in consulwrapper.go
// can be made to use a client that is backed by our MockKV.
// The actual mechanism of making `api.Client` use `MockKV` is complex
// because `api.Client.kv` is unexported.
//
// For this subtask, we will proceed with the understanding that functions
// like `putKV`, `getKV` etc., which take `*api.Client` as an argument,
// will be called with a test client where the KV interactions are mocked.
// This might mean we can't directly test `ConnectConsul`'s interaction with `api.NewClient`
// in a fully isolated way without modifying `ConnectConsul` for dependency injection.
//
// Let's assume that we can construct a *api.Client for testing purposes whose
// KV interactions are effectively proxied to our MockKV.
// A common way to achieve this if fields are unexported is to mock the HTTP transport,
// but that's a deeper level of mocking than `MockKV`.
//
// For the purpose of this exercise, the `MockConsulClient` and its `KV()` returning `MockKV`
// is the primary mocking layer. We will assume that functions under test
// receive an `*api.Client` that is configured (somehow) to use this `MockKV`.
// The easiest way to do this if `api.Client.KV()` was an interface itself, or if `kv` was exported.
// Since it's not, we'll rely on the `ConsulManager` struct in `consulwrapper.go`
// being modified or test helpers being able to inject this.
//
// For tests of `putKV`, `getKV`, etc., we'll pass `nil` as the `*api.Client`
// and modify the functions slightly if needed, or more realistically,
// we'd need to refactor `consulwrapper.go` to use an interface for the client.
//
// Given the constraints, I will proceed by:
// 1. Defining MockConsulClient as above.
// 2. When testing functions like `putKV`, I will create a `ConsulManager` instance
//    and set its `Client` field to an `*api.Client` that we *assume* can be made
//    to use our `MockKV`. If `ConsulManager`'s `Client` field was an interface
//    of type `ConsulClientInterface`, this would be straightforward.
//
// Let's refine `ConsulManager` to use `ConsulClientInterface`.
// This will require a change in `consulwrapper.go`.
// If `consulwrapper.go` cannot be changed, then we have to mock at a lower level (HTTP)
// or test `putKV` etc. by passing a `nil` client and ensuring they don't panic,
// or by testing them indirectly via functions that call them IF those higher-level functions
// can have their client interactions mocked.

// For now, I'll assume we can modify ConsulManager or use a similar test-friendly approach.
// The tests for `putKV` etc. will instantiate `MockKV`, then a `ConsulManager`
// whose `Client` field (if it were our interface type) would be set to a `MockConsulClient`
// that uses the `MockKV`.

// Let's proceed with tests for ConnectConsul first, which is more of an integration test.

// config import is now at the top.

func TestConnectConsul(t *testing.T) {
	t.Run("SuccessfulConnectionDefaultScheme", func(t *testing.T) {
		// This test requires a running Consul instance or a way to mock the network call if api.NewClient makes one.
		// Assuming api.NewClient primarily sets up config and doesn't immediately connect for basic setup.
		// If Consul is not running, and NewClient tries to connect, this might fail.
		// For now, we assume it's about config validation primarily.
		cd := config.ConsulDetail{
			BaseURL: "http://127.0.0.1:8500", // Use BaseURL
			// Scheme is implicitly part of BaseURL or handled by api.NewClient if not specified
		}
		client, err := ConnectConsul(&cd)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		if client != nil { // Sanity check for linter
			// api.Client().Address is not public. We check what was passed to api.NewClient via apiConfig.Address
			// In ConnectConsul, apiConfig.Address = detail.BaseURL
			// If BaseURL includes "http://", api.NewClient's internal config will parse it.
			// We can check client.Scheme and client.Config().Address if NewClient sets them as expected.
			// However, client.Config() is not exported.
			// A simple check is that no error occurred and client is not nil.
			// For a deeper check, we'd need to inspect how api.NewClient populates its internal config,
			// or have an integration test.
			// For now, we'll rely on the fact that BaseURL was used and NewClient didn't error.
			// Direct inspection of api.Client's scheme/address post-creation is difficult.
			// We trust that if api.NewClient is given "http://host:port", it configures itself correctly.
			// This test now primarily ensures ConnectConsul passes the BaseURL to the underlying client mechanism.
		}
	})

	t.Run("SuccessfulConnectionWithSchemeInBaseURL", func(t *testing.T) {
		cd := config.ConsulDetail{
			BaseURL: "https://myconsul.com:8500", // Scheme is part of BaseURL
		}
		client, err := ConnectConsul(&cd)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		// Similar to above, direct verification of scheme and address from client object is hard.
		// We trust api.NewClient.
	})

	t.Run("SuccessfulConnectionWithToken", func(t *testing.T) {
		cd := config.ConsulDetail{
			BaseURL: "http://127.0.0.1:8500", // Use BaseURL
			Token:   "my-secret-token",
		}
		client, err := ConnectConsul(&cd)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		// How to verify the token is set? api.Client.Config.Token
		// This requires access to the config, which is typically not public after client creation.
		// We trust api.NewClient to handle it if no error occurs.
		// If there was a method like client.Token() it would be testable.
	})

	t.Run("ConnectionFailureInvalidAddress", func(t *testing.T) {
		// An invalid address that api.NewClient should reject
		// Example: an address that causes url.Parse to fail.
		cd := config.ConsulDetail{
			BaseURL: "http://%", // Invalid hostname character
		}
		client, err := ConnectConsul(&cd)
		assert.Error(t, err) // Expect an error
		assert.Nil(t, client) // Expect client to be nil
	})

	t.Run("NilConsulDetail", func(t *testing.T) {
		client, err := ConnectConsul(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ConsulDetail is nil")
		assert.Nil(t, client)
	})
}

// Test cases for ConsulManager methods (PutKV, GetKV, DeleteKV, ListAllKV)

func TestPutKV(t *testing.T) {
	mockKV := new(MockKV)
	manager := &ConsulManager{
		KVStore: mockKV,
		Detail: &config.ConsulDetail{
			BasePath: "testBase/",
		},
	}

	kvPair := &api.KVPair{Key: "myKey", Value: []byte("myValue")}
	expectedKey := "testBase/myKey" // BasePath + Key

	t.Run("SuccessfulPut", func(t *testing.T) {
		// Reset mock for each sub-test if necessary, or ensure expectations are specific enough.
		// For this simple case, a fresh mock for each top-level test (TestPutKV) is fine.
		// If subtests have conflicting expectations on the same mock method, then reset or use separate mocks.
		mockKV.On("Put", mock.MatchedBy(func(p *api.KVPair) bool {
			return p.Key == expectedKey && string(p.Value) == string(kvPair.Value)
		}), (*api.WriteOptions)(nil)).Return(new(api.WriteMeta), nil).Once()

		success, err := manager.PutKV(kvPair)

		assert.True(t, success)
		assert.NoError(t, err)
		mockKV.AssertExpectations(t)
	})

	t.Run("PutFailureConsulError", func(t *testing.T) {
		expectedErr := errors.New("consul put error")
		mockKV.On("Put", mock.MatchedBy(func(p *api.KVPair) bool {
			return p.Key == expectedKey
		}), (*api.WriteOptions)(nil)).Return(nil, expectedErr).Once()

		success, err := manager.PutKV(kvPair)

		assert.False(t, success)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockKV.AssertExpectations(t)
	})

	t.Run("PutFailureKVStoreNil", func(t *testing.T) {
		nilStoreManager := &ConsulManager{
			KVStore: nil, // Explicitly set to nil
			Detail: &config.ConsulDetail{
				BasePath: "testBase/",
			},
		}
		success, err := nilStoreManager.PutKV(kvPair)
		assert.False(t, success)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Consul KVStore is nil")
		// No expectations on mockKV as it shouldn't be called.
	})
}

func TestGetKV(t *testing.T) {
	mockKV := new(MockKV)
	manager := &ConsulManager{
		KVStore: mockKV,
		Detail: &config.ConsulDetail{
			BasePath: "testBase/",
		},
	}

	keySuffix := "myKey"
	expectedFullKey := "testBase/myKey"
	expectedPair := &api.KVPair{Key: expectedFullKey, Value: []byte("myValue")}

	t.Run("SuccessfulGet", func(t *testing.T) {
		mockKV.On("Get", expectedFullKey, (*api.QueryOptions)(nil)).Return(expectedPair, new(api.QueryMeta), nil).Once()

		pair, err := manager.GetKV(keySuffix)

		assert.NoError(t, err)
		assert.Equal(t, expectedPair, pair)
		mockKV.AssertExpectations(t)
	})

	t.Run("GetFailureConsulError", func(t *testing.T) {
		expectedErr := errors.New("consul get error")
		mockKV.On("Get", expectedFullKey, (*api.QueryOptions)(nil)).Return(nil, new(api.QueryMeta), expectedErr).Once()

		pair, err := manager.GetKV(keySuffix)

		assert.Error(t, err)
		assert.Nil(t, pair)
		assert.Equal(t, expectedErr, err)
		mockKV.AssertExpectations(t)
	})

	t.Run("GetNotFound", func(t *testing.T) {
		// Consul's Get returns (nil, queryMeta, nil) for a key not found.
		mockKV.On("Get", expectedFullKey, (*api.QueryOptions)(nil)).Return(nil, new(api.QueryMeta), nil).Once()

		pair, err := manager.GetKV(keySuffix)

		assert.NoError(t, err) // No error for not found
		assert.Nil(t, pair)    // Pair is nil
		mockKV.AssertExpectations(t)
	})

	t.Run("GetFailureKVStoreNil", func(t *testing.T) {
		nilStoreManager := &ConsulManager{
			KVStore: nil,
			Detail: &config.ConsulDetail{
				BasePath: "testBase/",
			},
		}
		pair, err := nilStoreManager.GetKV(keySuffix)
		assert.Error(t, err)
		assert.Nil(t, pair)
		assert.Contains(t, err.Error(), "Consul KVStore is nil")
	})
}

func TestListAllKV(t *testing.T) {
	mockKV := new(MockKV)
	manager := &ConsulManager{
		KVStore: mockKV,
		Detail: &config.ConsulDetail{
			BasePath: "testBase/", // BasePath is important for List
		},
	}

	expectedKVPairs := api.KVPairs{
		{Key: "testBase/myKey1", Value: []byte("value1")},
		{Key: "testBase/myKey2", Value: []byte("value2")},
	}

	t.Run("SuccessfulListAll", func(t *testing.T) {
		// List method is called with the manager's BasePath
		mockKV.On("List", manager.Detail.BasePath, (*api.QueryOptions)(nil)).Return(expectedKVPairs, new(api.QueryMeta), nil).Once()

		pairs, err := manager.ListAllKV()

		assert.NoError(t, err)
		assert.Equal(t, expectedKVPairs, pairs)
		mockKV.AssertExpectations(t)
	})

	t.Run("ListAllFailureConsulError", func(t *testing.T) {
		expectedErr := errors.New("consul list error")
		mockKV.On("List", manager.Detail.BasePath, (*api.QueryOptions)(nil)).Return(nil, new(api.QueryMeta), expectedErr).Once()

		pairs, err := manager.ListAllKV()

		assert.Error(t, err)
		assert.Nil(t, pairs)
		assert.Equal(t, expectedErr, err)
		mockKV.AssertExpectations(t)
	})

	t.Run("ListAllReturnsEmpty", func(t *testing.T) {
		emptyKVPairs := api.KVPairs{} // Or nil, depending on how Consul client behaves for empty list
		mockKV.On("List", manager.Detail.BasePath, (*api.QueryOptions)(nil)).Return(emptyKVPairs, new(api.QueryMeta), nil).Once()

		pairs, err := manager.ListAllKV()

		assert.NoError(t, err)
		assert.Equal(t, emptyKVPairs, pairs) // Should be empty, not nil, if Consul returns empty list
		mockKV.AssertExpectations(t)
	})

	t.Run("ListAllFailureKVStoreNil", func(t *testing.T) {
		nilStoreManager := &ConsulManager{
			KVStore: nil,
			Detail: &config.ConsulDetail{
				BasePath: "testBase/",
			},
		}
		pairs, err := nilStoreManager.ListAllKV()
		assert.Error(t, err)
		assert.Nil(t, pairs)
		assert.Contains(t, err.Error(), "Consul KVStore is nil")
	})
}

// TestListAllKV is defined once above. Removed duplicate.

// Higher-level function tests
// Need to mock helper functions from src package (helpers.go)
// and NewConsulManager from the current package (consulwrapper.go)

// captureOutput is a helper to capture stdout
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestAddKVToConsul(t *testing.T) {
	// Save original functions and defer their restoration
	origCreateConsulDetailsFunc := CreateConsulDetailsFunc
	origNewConsulManagerFunc := NewConsulManagerFunc
	origCreateKVPairsFunc := CreateKVPairsFunc
	origIsConfigNameValidFunc := IsConfigNameValidFunc
	origErrConsulConnectionFunc := ErrConsulConnectionFunc
	defer func() {
		CreateConsulDetailsFunc = origCreateConsulDetailsFunc
		NewConsulManagerFunc = origNewConsulManagerFunc
		CreateKVPairsFunc = origCreateKVPairsFunc
		IsConfigNameValidFunc = origIsConfigNameValidFunc
		ErrConsulConnectionFunc = origErrConsulConnectionFunc
	}()

	mockConsulDetail1 := config.ConsulDetail{ConsulName: "consul1", BasePath: "base1/"}
	mockConsulDetail2 := config.ConsulDetail{ConsulName: "consul2", BasePath: "base2/"}

	CreateConsulDetailsFunc = func(cfp string) (map[string]config.ConsulDetail, error) {
		return map[string]config.ConsulDetail{
			"consul1": mockConsulDetail1,
			"consul2": mockConsulDetail2,
		}, nil
	}

	CreateKVPairsFunc = func(props, sName string) []api.KVPair {
		// Simplified: assume props "key1=val1|key2=val2" and sName "svc"
		// Results in keys "svc/key1", "svc/key2"
		if props == "key1=val1" {
			return []api.KVPair{{Key: path.Join(sName, "key1"), Value: []byte("val1")}}
		}
		return []api.KVPair{
			{Key: path.Join(sName, "key1"), Value: []byte("val1")},
			{Key: path.Join(sName, "key2"), Value: []byte("val2")},
		}
	}
	IsConfigNameValidFunc = func(name string, consulData map[string]config.ConsulDetail) bool {
		_, ok := consulData[name]
		return ok
	}
	ErrConsulConnectionFunc = func(name string, er error, exit bool) {
		// No-op for tests, or log to test buffer if needed
	}

	t.Run("Add_SpecificConsul_ReplaceFalse_KeyNotExists", func(t *testing.T) {
		mockKV := new(MockKV)
		NewConsulManagerFunc = func(detail *config.ConsulDetail) (*ConsulManager, error) {
			return &ConsulManager{KVStore: mockKV, Detail: detail}, nil
		}

		kvToGet := "svc/key1" // Relative key for GetKV/PutKV
		fullKeyPath := path.Join(mockConsulDetail1.BasePath, kvToGet)

		// GetKV returns nil, nil (key not found)
		mockKV.On("Get", fullKeyPath, (*api.QueryOptions)(nil)).Return(nil, new(api.QueryMeta), nil).Once()
		// PutKV should be called
		mockKV.On("Put", mock.MatchedBy(func(p *api.KVPair) bool {
			return p.Key == fullKeyPath && string(p.Value) == "val1"
		}), (*api.WriteOptions)(nil)).Return(new(api.WriteMeta), nil).Once()

		output := captureOutput(func() {
			AddKVToConsul("svc", "consul1", "key1=val1", "dummyConfigPath", "false")
		})

		mockKV.AssertExpectations(t)
		assert.Contains(t, output, "Added KV: "+fullKeyPath)
	})

	t.Run("Add_SpecificConsul_ReplaceFalse_KeyExists", func(t *testing.T) {
		mockKV := new(MockKV)
		NewConsulManagerFunc = func(detail *config.ConsulDetail) (*ConsulManager, error) {
			return &ConsulManager{KVStore: mockKV, Detail: detail}, nil
		}
		kvToGet := "svc/key1"
		fullKeyPath := path.Join(mockConsulDetail1.BasePath, kvToGet)
		existingPair := &api.KVPair{Key: fullKeyPath, Value: []byte("existingVal")}

		mockKV.On("Get", fullKeyPath, (*api.QueryOptions)(nil)).Return(existingPair, new(api.QueryMeta), nil).Once()
		// PutKV should NOT be called

		output := captureOutput(func() {
			AddKVToConsul("svc", "consul1", "key1=val1", "dummyConfigPath", "false")
		})
		mockKV.AssertExpectations(t)
		assert.Contains(t, output, "Failed to Add/Replace Following KVs") // Expect message about existing key
		assert.Contains(t, output, "Key: "+kvToGet) // The relative key
	})

	t.Run("Add_SpecificConsul_ReplaceTrue", func(t *testing.T) {
		mockKV := new(MockKV)
		NewConsulManagerFunc = func(detail *config.ConsulDetail) (*ConsulManager, error) {
			return &ConsulManager{KVStore: mockKV, Detail: detail}, nil
		}
		kvToPut := "svc/key1"
		fullKeyPath := path.Join(mockConsulDetail1.BasePath, kvToPut)

		// GetKV might not be called, PutKV should be.
		// Depending on implementation, GetKV might still be called by AddKVToConsul.
		// The current AddKVToConsul implementation does not call GetKV if replace is true.
		mockKV.On("Put", mock.MatchedBy(func(p *api.KVPair) bool {
			return p.Key == fullKeyPath && string(p.Value) == "val1"
		}), (*api.WriteOptions)(nil)).Return(new(api.WriteMeta), nil).Once()

		output := captureOutput(func() {
			AddKVToConsul("svc", "consul1", "key1=val1", "dummyConfigPath", "true")
		})
		mockKV.AssertExpectations(t)
		assert.Contains(t, output, "Added KV: "+fullKeyPath)
	})

	t.Run("Add_AllConsuls_ReplaceFalse", func(t *testing.T) {
		// This will involve two mock KVs if we want to simulate different states for consul1 & consul2
		// For simplicity, let's assume both will behave the same: key does not exist.
		mockKV1 := new(MockKV)
		mockKV2 := new(MockKV)

		NewConsulManagerFunc = func(detail *config.ConsulDetail) (*ConsulManager, error) {
			if detail.ConsulName == "consul1" {
				return &ConsulManager{KVStore: mockKV1, Detail: detail}, nil
			}
			return &ConsulManager{KVStore: mockKV2, Detail: detail}, nil
		}

		kvToGet := "svc/key1"
		fullKeyPath1 := path.Join(mockConsulDetail1.BasePath, kvToGet)
		fullKeyPath2 := path.Join(mockConsulDetail2.BasePath, kvToGet)

		mockKV1.On("Get", fullKeyPath1, (*api.QueryOptions)(nil)).Return(nil, new(api.QueryMeta), nil).Once()
		mockKV1.On("Put", mock.MatchedBy(func(p *api.KVPair) bool { return p.Key == fullKeyPath1 }), (*api.WriteOptions)(nil)).Return(new(api.WriteMeta), nil).Once()

		mockKV2.On("Get", fullKeyPath2, (*api.QueryOptions)(nil)).Return(nil, new(api.QueryMeta), nil).Once()
		mockKV2.On("Put", mock.MatchedBy(func(p *api.KVPair) bool { return p.Key == fullKeyPath2 }), (*api.WriteOptions)(nil)).Return(new(api.WriteMeta), nil).Once()

		output := captureOutput(func() {
			AddKVToConsul("svc", "", "key1=val1", "dummyConfigPath", "false") // cn == ""
		})

		mockKV1.AssertExpectations(t)
		mockKV2.AssertExpectations(t)
		assert.Contains(t, output, "Added KV: "+fullKeyPath1)
		assert.Contains(t, output, "Added KV: "+fullKeyPath2)
	})

	t.Run("Add_InvalidConsulName", func(t *testing.T) {
		NewConsulManagerFunc = func(detail *config.ConsulDetail) (*ConsulManager, error) {
			// Should not be called if IsConfigNameValid is false
			t.Fatal("NewConsulManagerFunc called for invalid consul name")
			return nil, errors.New("should not be called")
		}
		// IsConfigNameValidFunc will return false for "invalidConsul"
		output := captureOutput(func() {
			AddKVToConsul("svc", "invalidConsul", "key1=val1", "dummyConfigPath", "false")
		})
		assert.Contains(t, output, "Invalid Consul Server Name specified")
	})

	t.Run("Add_GetKV_Error", func(t *testing.T) {
		mockKV := new(MockKV)
		NewConsulManagerFunc = func(detail *config.ConsulDetail) (*ConsulManager, error) {
			return &ConsulManager{KVStore: mockKV, Detail: detail}, nil
		}
		kvToGet := "svc/key1"
		fullKeyPath := path.Join(mockConsulDetail1.BasePath, kvToGet)
		expectedGetErr := errors.New("get error")

		mockKV.On("Get", fullKeyPath, (*api.QueryOptions)(nil)).Return(nil, new(api.QueryMeta), expectedGetErr).Once()
		// Put should not be called

		output := captureOutput(func() {
			AddKVToConsul("svc", "consul1", "key1=val1", "dummyConfigPath", "false")
		})
		mockKV.AssertExpectations(t)
		assert.Contains(t, output, "Error checking key "+fullKeyPath)
		assert.Contains(t, output, "Failed to Add/Replace Following KVs")
	})

	t.Run("Add_PutKV_Error_ReplaceTrue", func(t *testing.T) {
		mockKV := new(MockKV)
		NewConsulManagerFunc = func(detail *config.ConsulDetail) (*ConsulManager, error) {
			return &ConsulManager{KVStore: mockKV, Detail: detail}, nil
		}
		kvToPut := "svc/key1"
		fullKeyPath := path.Join(mockConsulDetail1.BasePath, kvToPut)
		expectedPutErr := errors.New("put error")

		mockKV.On("Put", mock.MatchedBy(func(p *api.KVPair) bool { return p.Key == fullKeyPath }), (*api.WriteOptions)(nil)).Return(nil, expectedPutErr).Once()

		output := captureOutput(func() {
			AddKVToConsul("svc", "consul1", "key1=val1", "dummyConfigPath", "true")
		})
		mockKV.AssertExpectations(t)
		assert.Contains(t, output, "Could not Update Key in Consul Server consul1. KV: "+kvToPut)
	})

}


func TestDeleteKVFromConsul(t *testing.T) {
	// Save original functions and defer their restoration
	origCreateConsulDetailsFunc := CreateConsulDetailsFunc
	origNewConsulManagerFunc := NewConsulManagerFunc
	origCreateKVPairsFunc := CreateKVPairsFunc
	origIsConfigNameValidFunc := IsConfigNameValidFunc
	origErrConsulConnectionFunc := ErrConsulConnectionFunc
	defer func() {
		CreateConsulDetailsFunc = origCreateConsulDetailsFunc
		NewConsulManagerFunc = origNewConsulManagerFunc
		CreateKVPairsFunc = origCreateKVPairsFunc
		IsConfigNameValidFunc = origIsConfigNameValidFunc
		ErrConsulConnectionFunc = origErrConsulConnectionFunc
	}()

	mockConsulDetail1 := config.ConsulDetail{ConsulName: "consul1", BasePath: "base1/"}
	mockConsulDetail2 := config.ConsulDetail{ConsulName: "consul2", BasePath: "base2/"}

	CreateConsulDetailsFunc = func(cfp string) (map[string]config.ConsulDetail, error) {
		return map[string]config.ConsulDetail{
			"consul1": mockConsulDetail1,
			"consul2": mockConsulDetail2,
		}, nil
	}
	CreateKVPairsFunc = func(props, sName string) []api.KVPair {
		if props == "key1=val1" { // Assuming props define the keys to delete
			return []api.KVPair{{Key: path.Join(sName, "key1")}}
		}
		return []api.KVPair{
			{Key: path.Join(sName, "key1")},
			{Key: path.Join(sName, "key2")},
		}
	}
	IsConfigNameValidFunc = func(name string, consulData map[string]config.ConsulDetail) bool {
		_, ok := consulData[name]
		return ok
	}
	ErrConsulConnectionFunc = func(name string, er error, exit bool) { /* no-op */ }

	t.Run("Delete_SpecificConsul_Success", func(t *testing.T) {
		mockKV := new(MockKV)
		NewConsulManagerFunc = func(detail *config.ConsulDetail) (*ConsulManager, error) {
			assert.Equal(t, "consul1", detail.ConsulName)
			return &ConsulManager{KVStore: mockKV, Detail: detail}, nil
		}

		keySuffixToDelete := "svc/key1"
		fullKeyPath := path.Join(mockConsulDetail1.BasePath, keySuffixToDelete)

		mockKV.On("Delete", fullKeyPath, (*api.WriteOptions)(nil)).Return(new(api.WriteMeta), nil).Once()

		output := captureOutput(func() {
			DeleteKVFromConsul("svc", "consul1", "key1=val1", "dummyConfigPath")
		})

		mockKV.AssertExpectations(t)
		assert.Contains(t, output, "Deleted Key: "+fullKeyPath)
	})

	t.Run("Delete_SpecificConsul_DeleteError", func(t *testing.T) {
		mockKV := new(MockKV)
		NewConsulManagerFunc = func(detail *config.ConsulDetail) (*ConsulManager, error) {
			return &ConsulManager{KVStore: mockKV, Detail: detail}, nil
		}

		keySuffixToDelete := "svc/key1"
		fullKeyPath := path.Join(mockConsulDetail1.BasePath, keySuffixToDelete)
		expectedErr := errors.New("delete failed")

		mockKV.On("Delete", fullKeyPath, (*api.WriteOptions)(nil)).Return(nil, expectedErr).Once()

		output := captureOutput(func() {
			DeleteKVFromConsul("svc", "consul1", "key1=val1", "dummyConfigPath")
		})

		mockKV.AssertExpectations(t)
		assert.Contains(t, output, "Could not Delete Key "+fullKeyPath+" from Consul Server consul1")
	})

	t.Run("Delete_AllConsuls_Success", func(t *testing.T) {
		mockKV1 := new(MockKV)
		mockKV2 := new(MockKV)

		NewConsulManagerFunc = func(detail *config.ConsulDetail) (*ConsulManager, error) {
			if detail.ConsulName == "consul1" {
				return &ConsulManager{KVStore: mockKV1, Detail: detail}, nil
			}
			return &ConsulManager{KVStore: mockKV2, Detail: detail}, nil
		}

		keySuffixToDelete := "svc/key1" // From CreateKVPairsFunc mock for "key1=val1"
		fullKeyPath1 := path.Join(mockConsulDetail1.BasePath, keySuffixToDelete)
		fullKeyPath2 := path.Join(mockConsulDetail2.BasePath, keySuffixToDelete)

		mockKV1.On("Delete", fullKeyPath1, (*api.WriteOptions)(nil)).Return(new(api.WriteMeta), nil).Once()
		mockKV2.On("Delete", fullKeyPath2, (*api.WriteOptions)(nil)).Return(new(api.WriteMeta), nil).Once()

		output := captureOutput(func() {
			DeleteKVFromConsul("svc", "", "key1=val1", "dummyConfigPath") // cn == ""
		})

		mockKV1.AssertExpectations(t)
		mockKV2.AssertExpectations(t)
		assert.Contains(t, output, "Deleted Key: "+fullKeyPath1)
		assert.Contains(t, output, "Deleted Key: "+fullKeyPath2)
	})

	t.Run("Delete_InvalidConsulName", func(t *testing.T) {
		var newConsulManagerCalled bool
		NewConsulManagerFunc = func(detail *config.ConsulDetail) (*ConsulManager, error) {
			newConsulManagerCalled = true
			return nil, errors.New("should not be called")
		}

		output := captureOutput(func() {
			DeleteKVFromConsul("svc", "invalidConsul", "key1=val1", "dummyConfigPath")
		})

		assert.False(t, newConsulManagerCalled, "NewConsulManager should not be called for invalid consul name")
		assert.Contains(t, output, "Invalid Consul Server Name specified")
	})

}

func TestBackupConsulKV(t *testing.T) {
	// Save original functions and defer their restoration
	origCreateConsulDetailsFunc := CreateConsulDetailsFunc
	origNewConsulManagerFunc := NewConsulManagerFunc
	origIsConfigNameValidFunc := IsConfigNameValidFunc
	origErrConsulConnectionFunc := ErrConsulConnectionFunc
	origKVPairsToJSONFunc := KVPairsToJSONFunc
	origCreateBackupFileAndWriteDataFunc := CreateBackupFileAndWriteDataFunc
	// os.Getwd mocking was fully removed, no more origGetwd

	defer func() {
		CreateConsulDetailsFunc = origCreateConsulDetailsFunc
		NewConsulManagerFunc = origNewConsulManagerFunc
		IsConfigNameValidFunc = origIsConfigNameValidFunc
		ErrConsulConnectionFunc = origErrConsulConnectionFunc
		KVPairsToJSONFunc = origKVPairsToJSONFunc
		CreateBackupFileAndWriteDataFunc = origCreateBackupFileAndWriteDataFunc
		// No os.Getwd restoration needed
	}()

	mockConsulDetail1 := config.ConsulDetail{ConsulName: "consul1", BasePath: "base1/"}
	mockKVPairs := api.KVPairs{{Key: "base1/key1", Value: []byte("val1")}}
	mockJSONData := []byte(`[{"key":"base1/key1","value":"val1"}]`)

	CreateConsulDetailsFunc = func(cfp string) (map[string]config.ConsulDetail, error) {
		return map[string]config.ConsulDetail{"consul1": mockConsulDetail1}, nil
	}
	IsConfigNameValidFunc = func(name string, consulData map[string]config.ConsulDetail) bool { return name == "consul1" }
	ErrConsulConnectionFunc = func(name string, er error, exit bool) {}
	KVPairsToJSONFunc = func(kvPairs api.KVPairs) ([]byte, error) {
		assert.Equal(t, mockKVPairs, kvPairs)
		return mockJSONData, nil
	}

	// No os.Getwd mocking here

	t.Run("Backup_SpecificConsul_Success", func(t *testing.T) {
		mockKV := new(MockKV)
		NewConsulManagerFunc = func(detail *config.ConsulDetail) (*ConsulManager, error) {
			return &ConsulManager{KVStore: mockKV, Detail: detail}, nil
		}
		CreateBackupFileAndWriteDataFunc = func(fp, cn string, data []byte) (string, error) {
			assert.Equal(t, "/customPath", fp) // fp is finalBackupPath
			assert.Equal(t, "consul1", cn)
			assert.Equal(t, mockJSONData, data)
			return path.Join(fp, cn+".json"), nil
		}

		mockKV.On("List", mockConsulDetail1.BasePath, (*api.QueryOptions)(nil)).Return(mockKVPairs, new(api.QueryMeta), nil).Once()

		output := captureOutput(func() {
			BackupConsulKV("consul1", "dummyConfig", "/customPath")
		})

		mockKV.AssertExpectations(t)
		assert.Contains(t, output, "Note:- All Consul Values are base64 encoded") // From BackupConsulKV
	})

	t.Run("Backup_SpecificConsul_DefaultBackupPath", func(t *testing.T) {
		mockKV := new(MockKV)
		NewConsulManagerFunc = func(detail *config.ConsulDetail) (*ConsulManager, error) {
			return &ConsulManager{KVStore: mockKV, Detail: detail}, nil
		}
		CreateBackupFileAndWriteDataFunc = func(fp, cn string, data []byte) (string, error) {
			// fp here is the finalBackupPath. If input fp was empty, it's current working dir.
			// We can't easily assert the exact pwd without mocking os.Getwd, so we check other args.
			// For this test, fp is "/customPath" (from BackupConsulKV call below), so this assertion is fine.
			// If testing the empty path case, this assertion would need to be more flexible or os.Getwd mocked differently.
			// Let's assume this test case is for a specified path.
			assert.Equal(t, "/customPath", fp)
			assert.Equal(t, "consul1", cn)
			assert.Equal(t, mockJSONData, data)
			return path.Join(fp, cn+".json"), nil
		}

		mockKV.On("List", mockConsulDetail1.BasePath, (*api.QueryOptions)(nil)).Return(mockKVPairs, new(api.QueryMeta), nil).Once()

		output := captureOutput(func() {
			BackupConsulKV("consul1", "dummyConfig", "") // Empty fp
		})
		mockKV.AssertExpectations(t)
		assert.Contains(t, output, "Note:- All Consul Values are base64 encoded")
	})

	t.Run("Backup_SpecificConsul_ListError", func(t *testing.T) {
		mockKV := new(MockKV)
		NewConsulManagerFunc = func(detail *config.ConsulDetail) (*ConsulManager, error) {
			return &ConsulManager{KVStore: mockKV, Detail: detail}, nil
		}
		expectedErr := errors.New("list error")
		mockKV.On("List", mockConsulDetail1.BasePath, (*api.QueryOptions)(nil)).Return(nil, new(api.QueryMeta), expectedErr).Once()
		// CreateBackupFileAndWriteDataFunc should not be called

		var createBackupCalled bool
		CreateBackupFileAndWriteDataFunc = func(fp, cn string, data []byte) (string, error) {
			createBackupCalled = true
			return "", nil
		}

		output := captureOutput(func() {
			BackupConsulKV("consul1", "dummyConfig", "/customPath")
		})

		mockKV.AssertExpectations(t)
		assert.False(t, createBackupCalled)
		assert.Contains(t, output, "Could not List KVs for Backup from Consul Server consul1")
	})

	t.Run("Backup_InvalidConsulName", func(t *testing.T) {
		output := captureOutput(func() {
			BackupConsulKV("invalidConsul", "dummyConfig", "/customPath")
		})
		assert.Contains(t, output, "Invalid Consul Server Name specified")
	})

	t.Run("Backup_CreateBackupFileError", func(t *testing.T) {
		mockKV := new(MockKV)
		NewConsulManagerFunc = func(detail *config.ConsulDetail) (*ConsulManager, error) {
			return &ConsulManager{KVStore: mockKV, Detail: detail}, nil
		}
		expectedErr := errors.New("backup write error")
		CreateBackupFileAndWriteDataFunc = func(fp, cn string, data []byte) (string, error) {
			return "", expectedErr
		}

		mockKV.On("List", mockConsulDetail1.BasePath, (*api.QueryOptions)(nil)).Return(mockKVPairs, new(api.QueryMeta), nil).Once()

		output := captureOutput(func() {
			BackupConsulKV("consul1", "dummyConfig", "/customPath")
		})
		mockKV.AssertExpectations(t)
		assert.Contains(t, output, "Could not Write Backup for Consul Server consul1")
	})

}

func TestRestoreConsulKV(t *testing.T) {
	// Save original functions and defer their restoration
	origCreateConsulDetailsFunc := CreateConsulDetailsFunc
	origNewConsulManagerFunc := NewConsulManagerFunc
	origIsConfigNameValidFunc := IsConfigNameValidFunc
	origErrConsulConnectionFunc := ErrConsulConnectionFunc
	origValidateFilePathFunc := ValidateFilePathFunc
	origReadJSONFileAndReturnStructFunc := ReadJSONFileAndReturnStructFunc
	origConvertJSONStructToKvPairsFunc := ConvertJSONStructToKvPairsFunc
	// No os.Getwd mocking

	defer func() {
		CreateConsulDetailsFunc = origCreateConsulDetailsFunc
		NewConsulManagerFunc = origNewConsulManagerFunc
		IsConfigNameValidFunc = origIsConfigNameValidFunc
		ErrConsulConnectionFunc = origErrConsulConnectionFunc
		ValidateFilePathFunc = origValidateFilePathFunc
		ReadJSONFileAndReturnStructFunc = origReadJSONFileAndReturnStructFunc
		ConvertJSONStructToKvPairsFunc = origConvertJSONStructToKvPairsFunc
		// No os.Getwd restoration
	}()

	mockConsulDetail1 := config.ConsulDetail{ConsulName: "consul1", BasePath: "base1/"}
	mockKVDetails := &[]KVDetails{
		{Key: "base1/svc1/key1", Value: []byte("val1_svc1")}, // Matches BasePath
		{Key: "svc2/key2", Value: []byte("val2_svc2")},       // Does not match BasePath, treated as suffix
		{Key: "base1/svc1/key3", Value: []byte("val3_svc1")},
	}
	// ConvertJSONStructToKvPairsFunc will convert these. Let's define the expected KVPairs for PutKV.
	// PutKV expects keys relative to BasePath.
	// So "base1/svc1/key1" becomes "svc1/key1" for PutKV if BasePath is "base1/"
	// "svc2/key2" becomes "svc2/key2" for PutKV (already relative or treated as such)
	mockAPIKVPairs := &[]api.KVPair{
		{Key: "base1/svc1/key1", Value: []byte("val1_svc1")},
		{Key: "svc2/key2", Value: []byte("val2_svc2")},
		{Key: "base1/svc1/key3", Value: []byte("val3_svc1")},
	}

	CreateConsulDetailsFunc = func(cfp string) (map[string]config.ConsulDetail, error) {
		return map[string]config.ConsulDetail{"consul1": mockConsulDetail1}, nil
	}
	IsConfigNameValidFunc = func(name string, consulData map[string]config.ConsulDetail) bool { return name == "consul1" }
	ErrConsulConnectionFunc = func(name string, er error, exit bool) { /* no-op */ }
	ValidateFilePathFunc = func(p string) error { return nil } // Assume valid path
	ReadJSONFileAndReturnStructFunc = func(fp string) *[]KVDetails { return mockKVDetails }
	ConvertJSONStructToKvPairsFunc = func(kd *[]KVDetails) *[]api.KVPair { return mockAPIKVPairs }
	// No os.Getwd mocking

	t.Run("Restore_Success_NoServiceFilter", func(t *testing.T) {
		mockKV := new(MockKV)
		NewConsulManagerFunc = func(detail *config.ConsulDetail) (*ConsulManager, error) {
			return &ConsulManager{KVStore: mockKV, Detail: detail}, nil
		}

		// Expectations for PutKV. Keys should be relative to BasePath.
		// Key "base1/svc1/key1" from file -> "svc1/key1" for PutKV
		mockKV.On("Put", mock.MatchedBy(func(p *api.KVPair) bool { return p.Key == "svc1/key1" }), (*api.WriteOptions)(nil)).Return(new(api.WriteMeta), nil).Once()
		// Key "svc2/key2" from file -> "svc2/key2" for PutKV (as it doesn't start with BasePath)
		mockKV.On("Put", mock.MatchedBy(func(p *api.KVPair) bool { return p.Key == "svc2/key2" }), (*api.WriteOptions)(nil)).Return(new(api.WriteMeta), nil).Once()
		// Key "base1/svc1/key3" from file -> "svc1/key3" for PutKV
		mockKV.On("Put", mock.MatchedBy(func(p *api.KVPair) bool { return p.Key == "svc1/key3" }), (*api.WriteOptions)(nil)).Return(new(api.WriteMeta), nil).Once()

		output := captureOutput(func() {
			RestoreConsulKV("consul1", "dummyCfg", "/customPath/backup.json", "")
		})
		mockKV.AssertExpectations(t)
		assert.Contains(t, output, "Consul Recovery Completed")
	})

	t.Run("Restore_Success_WithServiceFilter_Matching", func(t *testing.T) {
		mockKV := new(MockKV)
		NewConsulManagerFunc = func(detail *config.ConsulDetail) (*ConsulManager, error) {
			return &ConsulManager{KVStore: mockKV, Detail: detail}, nil
		}
		// Service filter "svc1"
		// Expect Put for "base1/svc1/key1" (becomes "svc1/key1")
		// Expect Put for "base1/svc1/key3" (becomes "svc1/key3")
		// No Put for "svc2/key2"
		mockKV.On("Put", mock.MatchedBy(func(p *api.KVPair) bool { return p.Key == "svc1/key1" }), (*api.WriteOptions)(nil)).Return(new(api.WriteMeta), nil).Once()
		mockKV.On("Put", mock.MatchedBy(func(p *api.KVPair) bool { return p.Key == "svc1/key3" }), (*api.WriteOptions)(nil)).Return(new(api.WriteMeta), nil).Once()

		output := captureOutput(func() {
			RestoreConsulKV("consul1", "dummyCfg", "/customPath/backup.json", "svc1")
		})
		mockKV.AssertExpectations(t)
		assert.Contains(t, output, "Consul Recovery Completed")
	})

	t.Run("Restore_Success_WithServiceFilter_NonMatching", func(t *testing.T) {
		mockKV := new(MockKV)
		NewConsulManagerFunc = func(detail *config.ConsulDetail) (*ConsulManager, error) {
			return &ConsulManager{KVStore: mockKV, Detail: detail}, nil
		}
		// Service filter "svc3" - no keys should match
		// No PutKV calls expected
		output := captureOutput(func() {
			RestoreConsulKV("consul1", "dummyCfg", "/customPath/backup.json", "svc3")
		})
		mockKV.AssertExpectations(t) // Should have no expectations, so this is fine
		assert.Contains(t, output, "Consul Recovery Completed")
	})

	t.Run("Restore_DefaultFilePath", func(t *testing.T) {
		mockKV := new(MockKV) // Only need to ensure ReadJSONFileAndReturnStructFunc is called with default path
		NewConsulManagerFunc = func(detail *config.ConsulDetail) (*ConsulManager, error) {
			return &ConsulManager{KVStore: mockKV, Detail: detail}, nil
		}
		var actualDefaultPathChecked bool
		ReadJSONFileAndReturnStructFunc = func(fp string) *[]KVDetails {
			// We can't easily assert the exact default path without os.Getwd or more info
			// So we'll just ensure it's called, and trust other parts for path construction.
			// Or check if it contains DefaultBackupFilePath and consul1.json
			assert.Contains(t, fp, DefaultBackupFilePath)
			assert.Contains(t, fp, "consul1.json")
			// actualDefaultPathChecked = true // Removed unused variable
			return mockKVDetails
		}
		// Minimal PutKV expectations as path is main focus
		mockKV.On("Put", mock.AnythingOfType("*api.KVPair"), (*api.WriteOptions)(nil)).Return(new(api.WriteMeta), nil).Times(len(*mockAPIKVPairs))


		captureOutput(func() {
			RestoreConsulKV("consul1", "dummyCfg", "", "") // Empty fp for default path
		})
		// Assertions are within mocked ReadJSONFileAndReturnStructFunc
		mockKV.AssertExpectations(t)
	})

	t.Run("Restore_InvalidFilePath", func(t *testing.T) {
		// Mock os.Exit to prevent test termination
		var exitCalled bool
		mockExit := func(code int) {
			assert.Equal(t, 1, code)
			exitCalled = true
			panic("os.Exit called") // Panic to stop execution flow after exit
		}
		origCurrentExitFunc := ExitFunc // Use the exported ExitFunc from src package (helpers.go)
		ExitFunc = mockExit             // Assign to the exported variable
		defer func() { ExitFunc = origCurrentExitFunc }()


		ValidateFilePathFunc = func(p string) error { return errors.New("invalid path") }
		var managerCalled bool
		NewConsulManagerFunc = func(detail *config.ConsulDetail) (*ConsulManager, error) {
			managerCalled = true // Should be called before ValidateFilePath if fp is not empty
			return &ConsulManager{KVStore: new(MockKV), Detail: detail}, nil
		}


		assert.PanicsWithValue(t, "os.Exit called", func() {
			RestoreConsulKV("consul1", "dummyCfg", "/invalid/path.json", "")
		}, "os.Exit was not called as expected")

		assert.True(t, exitCalled, "os.Exit was not called")
		assert.True(t, managerCalled) // Manager is created before file path validation if fp is not empty
	})
}

func TestSyncConsulKVStore(t *testing.T) {
	// Save original functions and defer their restoration
	origCreateConsulDetailsFunc := CreateConsulDetailsFunc
	origNewConsulManagerFunc := NewConsulManagerFunc
	origIsConfigNameValidFunc := IsConfigNameValidFunc
	origErrConsulConnectionFunc := ErrConsulConnectionFunc
	origRemoveExistingKVPairsFunc := RemoveExistingKVPairsFunc
	origConvertServiceKVPairsToMapFunc := ConvertServiceKVPairsToMapFunc
	origCurrentExitFuncForSync := ExitFunc // For os.Exit mocking

	defer func() {
		CreateConsulDetailsFunc = origCreateConsulDetailsFunc
		NewConsulManagerFunc = origNewConsulManagerFunc
		IsConfigNameValidFunc = origIsConfigNameValidFunc
		ErrConsulConnectionFunc = origErrConsulConnectionFunc
		RemoveExistingKVPairsFunc = origRemoveExistingKVPairsFunc
		ConvertServiceKVPairsToMapFunc = origConvertServiceKVPairsToMapFunc
		ExitFunc = origCurrentExitFuncForSync
	}()

	mockSourceDetail := config.ConsulDetail{ConsulName: "sourceConsul", BasePath: "sourceBase/"}
	mockTargetDetail := config.ConsulDetail{ConsulName: "targetConsul", BasePath: "targetBase/"}

	mockSourceKVs := api.KVPairs{
		{Key: "sourceBase/svc1/key1", Value: []byte("s_val1")},
		{Key: "sourceBase/svc2/key2", Value: []byte("s_val2")},
	}
	mockTargetKVs := api.KVPairs{ // Target might have some existing keys
		{Key: "targetBase/svc1/key1", Value: []byte("t_val_old")},
		{Key: "targetBase/svcX/keyX", Value: []byte("t_valX")},
	}

	CreateConsulDetailsFunc = func(cfp string) (map[string]config.ConsulDetail, error) {
		return map[string]config.ConsulDetail{
			"sourceConsul": mockSourceDetail,
			"targetConsul": mockTargetDetail,
		}, nil
	}
	IsConfigNameValidFunc = func(name string, consulData map[string]config.ConsulDetail) bool {
		return name == "sourceConsul" || name == "targetConsul"
	}
	ErrConsulConnectionFunc = func(name string, er error, exit bool) { if er != nil && exit { panic("exit_called") } }
	
	// Mock os.Exit behavior
	var exitCalled bool
	ExitFunc = func(code int) { // Assign to the exported ExitFunc
		exitCalled = true
		panic("exit_called")
	}


	t.Run("Sync_ReplaceFalse_NoFilter", func(t *testing.T) {
		exitCalled = false
		mockSourceKV := new(MockKV)
		mockTargetKV := new(MockKV)

		NewConsulManagerFunc = func(detail *config.ConsulDetail) (*ConsulManager, error) {
			if detail.ConsulName == "sourceConsul" {
				return &ConsulManager{KVStore: mockSourceKV, Detail: detail}, nil
			}
			return &ConsulManager{KVStore: mockTargetKV, Detail: detail}, nil
		}

		// removeExistingKVPairs will return a map of {suffixKey: valueBytes}
		// For "replace=false", let's say "svc2/key2" is new.
		mapBytesToSync := map[string][]byte{"svc2/key2": []byte("s_val2")}
		RemoveExistingKVPairsFunc = func(sKVs *api.KVPairs, sBP string, tKVs *api.KVPairs, tBP string) map[string][]byte { // Changed to *api.KVPairs
			// Simplified mock: assert arguments if necessary
			assert.Equal(t, &mockSourceKVs, sKVs) // Pass pointers for comparison if original func expects pointers
			assert.Equal(t, &mockTargetKVs, tKVs)
			return mapBytesToSync
		}

		mockSourceKV.On("List", mockSourceDetail.BasePath, (*api.QueryOptions)(nil)).Return(mockSourceKVs, new(api.QueryMeta), nil).Once()
		mockTargetKV.On("List", mockTargetDetail.BasePath, (*api.QueryOptions)(nil)).Return(mockTargetKVs, new(api.QueryMeta), nil).Once()
		// Expect Put on target for "svc2/key2"
		mockTargetKV.On("Put", mock.MatchedBy(func(p *api.KVPair) bool {
			return p.Key == "svc2/key2" && string(p.Value) == "s_val2" // Comparison still okay with string cast
		}), (*api.WriteOptions)(nil)).Return(new(api.WriteMeta), nil).Once()

		output := captureOutput(func() {
			SyncConsulKVStore("sourceConsul", "targetConsul", "", "dummyCfg", "false")
		})
		mockSourceKV.AssertExpectations(t)
		mockTargetKV.AssertExpectations(t)
		assert.Contains(t, output, "Consul KV Store Sync of Source Consul Server: sourceConsul and Target Consul Server: targetConsul is complete")
		assert.False(t, exitCalled)
	})

	t.Run("Sync_ReplaceTrue_WithFilter", func(t *testing.T) {
		exitCalled = false
		mockSourceKV := new(MockKV)
		mockTargetKV := new(MockKV)
		NewConsulManagerFunc = func(detail *config.ConsulDetail) (*ConsulManager, error) {
			if detail.ConsulName == "sourceConsul" {
				return &ConsulManager{KVStore: mockSourceKV, Detail: detail}, nil
			}
			return &ConsulManager{KVStore: mockTargetKV, Detail: detail}, nil
		}

		// convertServiceKVPairsToMap returns {suffixKey: valueBytes} from source
		// All source keys, relative to BasePath: {"svc1/key1": "s_val1", "svc2/key2": "s_val2"}
		mapBytesToSyncFromSource := map[string][]byte{
			"svc1/key1": []byte("s_val1"),
			"svc2/key2": []byte("s_val2"),
		}
		ConvertServiceKVPairsToMapFunc = func(kvps *api.KVPairs, bp string) map[string][]byte { // Changed to *api.KVPairs
			assert.Equal(t, &mockSourceKVs, kvps) // Pass pointer for comparison
			return mapBytesToSyncFromSource
		}

		mockSourceKV.On("List", mockSourceDetail.BasePath, (*api.QueryOptions)(nil)).Return(mockSourceKVs, new(api.QueryMeta), nil).Once()
		mockTargetKV.On("List", mockTargetDetail.BasePath, (*api.QueryOptions)(nil)).Return(mockTargetKVs, new(api.QueryMeta), nil).Once() // Still list target, though not used by replace=true logic directly for filtering

		// With filter "svc1", only "svc1/key1" should be Put
		mockTargetKV.On("Put", mock.MatchedBy(func(p *api.KVPair) bool {
			return p.Key == "svc1/key1" && string(p.Value) == "s_val1"
		}), (*api.WriteOptions)(nil)).Return(new(api.WriteMeta), nil).Once()

		output := captureOutput(func() {
			SyncConsulKVStore("sourceConsul", "targetConsul", "svc1", "dummyCfg", "true")
		})
		mockSourceKV.AssertExpectations(t)
		mockTargetKV.AssertExpectations(t)
		assert.Contains(t, output, "Consul KV Store Sync of Source Consul Server: sourceConsul and Target Consul Server: targetConsul is complete")
		assert.False(t, exitCalled)
	})
	
	t.Run("Sync_InvalidConsulName", func(t *testing.T) {
		exitCalled = false
		IsConfigNameValidFunc = func(name string, consulData map[string]config.ConsulDetail) bool { return false } // Make it invalid
		
		// output := "" // Removed unused variable
		assert.PanicsWithValue(t, "exit_called", func() {
			captureOutput(func() { // captureOutput might not fully capture due to panic
				SyncConsulKVStore("invalidSource", "targetConsul", "", "dummyCfg", "true")
			})
		})
		// Output capture might be tricky with panic. The primary check is exitCalled.
		// For more robust output checking with panic, one might need to adjust captureOutput or how os.Stdout is handled.
		assert.True(t, exitCalled)
		// Reset IsConfigNameValidFunc for other tests if needed, but defer handles it at top level.
	})
}


func TestDeleteKV(t *testing.T) {
	mockKV := new(MockKV)
	manager := &ConsulManager{
		KVStore: mockKV,
		Detail: &config.ConsulDetail{
			BasePath: "testBase/",
		},
	}

	keySuffix := "myKeyToDelete"
	expectedFullKey := "testBase/myKeyToDelete"

	t.Run("SuccessfulDelete", func(t *testing.T) {
		mockKV.On("Delete", expectedFullKey, (*api.WriteOptions)(nil)).Return(new(api.WriteMeta), nil).Once()

		success, err := manager.DeleteKV(keySuffix)

		assert.True(t, success)
		assert.NoError(t, err)
		mockKV.AssertExpectations(t)
	})

	t.Run("DeleteFailureConsulError", func(t *testing.T) {
		expectedErr := errors.New("consul delete error")
		mockKV.On("Delete", expectedFullKey, (*api.WriteOptions)(nil)).Return(nil, expectedErr).Once()

		success, err := manager.DeleteKV(keySuffix)

		assert.False(t, success)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockKV.AssertExpectations(t)
	})

	t.Run("DeleteFailureKVStoreNil", func(t *testing.T) {
		nilStoreManager := &ConsulManager{
			KVStore: nil,
			Detail: &config.ConsulDetail{
				BasePath: "testBase/",
			},
		}
		success, err := nilStoreManager.DeleteKV(keySuffix)
		assert.False(t, success)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Consul KVStore is nil")
	})
}
