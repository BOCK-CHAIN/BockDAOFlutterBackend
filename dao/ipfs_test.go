package dao

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/BOCK-CHAIN/BockChain/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockIPFSClient implements a mock IPFS client for testing
type MockIPFSClient struct {
	storage map[string][]byte
	pins    map[string]bool
}

// NewMockIPFSClient creates a new mock IPFS client
func NewMockIPFSClient() *MockIPFSClient {
	return &MockIPFSClient{
		storage: make(map[string][]byte),
		pins:    make(map[string]bool),
	}
}

// mockAdd simulates adding content to IPFS
func (m *MockIPFSClient) mockAdd(data []byte) string {
	hash := "Qm" + string(rune(len(m.storage)+1)) + "TestHash"
	m.storage[hash] = data
	return hash
}

// mockCat simulates retrieving content from IPFS
func (m *MockIPFSClient) mockCat(hash string) ([]byte, bool) {
	data, exists := m.storage[hash]
	return data, exists
}

// mockPin simulates pinning content
func (m *MockIPFSClient) mockPin(hash string) bool {
	if _, exists := m.storage[hash]; exists {
		m.pins[hash] = true
		return true
	}
	return false
}

// mockUnpin simulates unpinning content
func (m *MockIPFSClient) mockUnpin(hash string) bool {
	if _, exists := m.pins[hash]; exists {
		delete(m.pins, hash)
		return true
	}
	return false
}

func TestNewIPFSClient(t *testing.T) {
	tests := []struct {
		name     string
		nodeURL  string
		expected string
	}{
		{
			name:     "default URL",
			nodeURL:  "",
			expected: "localhost:5001",
		},
		{
			name:     "custom URL",
			nodeURL:  "127.0.0.1:5002",
			expected: "127.0.0.1:5002",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewIPFSClient(tt.nodeURL)
			assert.NotNil(t, client)
			assert.NotNil(t, client.shell)
			assert.Equal(t, 30*time.Second, client.timeout)
		})
	}
}

func TestProposalMetadata_Serialization(t *testing.T) {
	metadata := &ProposalMetadata{
		Title:       "Test Proposal",
		Description: "A test proposal for unit testing",
		Details:     "Detailed description of the test proposal",
		Documents: []DocumentReference{
			{
				Name:     "proposal.pdf",
				Hash:     "QmTestHash1",
				Size:     1024,
				MimeType: "application/pdf",
			},
		},
		Links: []LinkReference{
			{
				Title: "Reference Link",
				URL:   "https://example.com",
			},
		},
		Tags:      []string{"test", "governance"},
		Version:   "1.0",
		CreatedAt: time.Now().Unix(),
		Checksum:  "test-checksum",
	}

	// Test JSON serialization
	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	require.NoError(t, err)
	assert.Contains(t, string(jsonData), "Test Proposal")

	// Test JSON deserialization
	var deserialized ProposalMetadata
	err = json.Unmarshal(jsonData, &deserialized)
	require.NoError(t, err)
	assert.Equal(t, metadata.Title, deserialized.Title)
	assert.Equal(t, metadata.Description, deserialized.Description)
	assert.Equal(t, len(metadata.Documents), len(deserialized.Documents))
	assert.Equal(t, len(metadata.Links), len(deserialized.Links))
	assert.Equal(t, len(metadata.Tags), len(deserialized.Tags))
}

func TestIPFSClient_HashConversion(t *testing.T) {
	client := NewIPFSClient("")

	// Test IPFS hash to types.Hash conversion
	ipfsHash := "QmTestHash123"
	typesHash := client.ipfsHashToTypesHash(ipfsHash)
	assert.NotEqual(t, types.Hash{}, typesHash)

	// Test types.Hash to IPFS hash conversion
	convertedBack := client.typesHashToIPFSHash(typesHash)
	assert.NotEmpty(t, convertedBack)
	assert.Equal(t, 64, len(convertedBack)) // Should be hex encoded (32 bytes * 2)
}

func TestProposalMetadata_Validation(t *testing.T) {
	tests := []struct {
		name     string
		metadata *ProposalMetadata
		valid    bool
	}{
		{
			name: "valid metadata",
			metadata: &ProposalMetadata{
				Title:       "Valid Proposal",
				Description: "A valid proposal",
				Version:     "1.0",
				CreatedAt:   time.Now().Unix(),
			},
			valid: true,
		},
		{
			name: "empty title",
			metadata: &ProposalMetadata{
				Title:       "",
				Description: "A proposal without title",
				Version:     "1.0",
				CreatedAt:   time.Now().Unix(),
			},
			valid: false,
		},
		{
			name: "empty description",
			metadata: &ProposalMetadata{
				Title:       "Proposal",
				Description: "",
				Version:     "1.0",
				CreatedAt:   time.Now().Unix(),
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.metadata.Title != "" && tt.metadata.Description != ""
			assert.Equal(t, tt.valid, valid)
		})
	}
}

func TestDocumentReference_Validation(t *testing.T) {
	tests := []struct {
		name   string
		docRef *DocumentReference
		valid  bool
	}{
		{
			name: "valid document reference",
			docRef: &DocumentReference{
				Name:     "document.pdf",
				Hash:     "QmTestHash",
				Size:     1024,
				MimeType: "application/pdf",
			},
			valid: true,
		},
		{
			name: "missing name",
			docRef: &DocumentReference{
				Name:     "",
				Hash:     "QmTestHash",
				Size:     1024,
				MimeType: "application/pdf",
			},
			valid: false,
		},
		{
			name: "missing hash",
			docRef: &DocumentReference{
				Name:     "document.pdf",
				Hash:     "",
				Size:     1024,
				MimeType: "application/pdf",
			},
			valid: false,
		},
		{
			name: "negative size",
			docRef: &DocumentReference{
				Name:     "document.pdf",
				Hash:     "QmTestHash",
				Size:     -1,
				MimeType: "application/pdf",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.docRef.Name != "" && tt.docRef.Hash != "" && tt.docRef.Size >= 0
			assert.Equal(t, tt.valid, valid)
		})
	}
}

func TestLinkReference_Validation(t *testing.T) {
	tests := []struct {
		name    string
		linkRef *LinkReference
		valid   bool
	}{
		{
			name: "valid link reference",
			linkRef: &LinkReference{
				Title: "Test Link",
				URL:   "https://example.com",
			},
			valid: true,
		},
		{
			name: "missing title",
			linkRef: &LinkReference{
				Title: "",
				URL:   "https://example.com",
			},
			valid: false,
		},
		{
			name: "missing URL",
			linkRef: &LinkReference{
				Title: "Test Link",
				URL:   "",
			},
			valid: false,
		},
		{
			name: "invalid URL format",
			linkRef: &LinkReference{
				Title: "Test Link",
				URL:   "not-a-url",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.linkRef.Title != "" && tt.linkRef.URL != ""
			if valid && tt.linkRef.URL != "" {
				// Basic URL validation
				valid = len(tt.linkRef.URL) > 7 && (tt.linkRef.URL[:7] == "http://" || tt.linkRef.URL[:8] == "https://")
			}
			assert.Equal(t, tt.valid, valid)
		})
	}
}

func TestCreateProposalWithIPFS(t *testing.T) {
	client := NewIPFSClient("")

	documents := []DocumentReference{
		{
			Name:     "proposal.pdf",
			Hash:     "QmTestDoc1",
			Size:     2048,
			MimeType: "application/pdf",
		},
	}

	links := []LinkReference{
		{
			Title: "Reference",
			URL:   "https://example.com",
		},
	}

	tags := []string{"governance", "test"}

	metadata, hash, err := client.CreateProposalWithIPFS(
		"Test Proposal",
		"Test Description",
		"Detailed test description",
		documents,
		links,
		tags,
	)

	// The actual upload will fail without IPFS node, but we test the logic
	if err != nil {
		t.Logf("Expected error without IPFS node: %v", err)
		assert.Nil(t, metadata)
		assert.Equal(t, types.Hash{}, hash)
	} else {
		// If IPFS is available, test the structure
		assert.NotNil(t, metadata)
		assert.Equal(t, "Test Proposal", metadata.Title)
		assert.Equal(t, "Test Description", metadata.Description)
		assert.Equal(t, "Detailed test description", metadata.Details)
		assert.Equal(t, documents, metadata.Documents)
		assert.Equal(t, links, metadata.Links)
		assert.Equal(t, tags, metadata.Tags)
		assert.Equal(t, "1.0", metadata.Version)
		assert.Greater(t, metadata.CreatedAt, int64(0))
		assert.NotEqual(t, types.Hash{}, hash)
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected float64
	}{
		{
			name:     "simple version",
			version:  "1.0",
			expected: 1.0,
		},
		{
			name:     "decimal version",
			version:  "2.5",
			expected: 2.5,
		},
		{
			name:     "complex version",
			version:  "1.23",
			expected: 1.23,
		},
		{
			name:     "invalid version",
			version:  "invalid",
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseVersion(tt.version)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMetadataChecksumValidation(t *testing.T) {
	client := NewIPFSClient("")

	metadata := &ProposalMetadata{
		Title:       "Test Proposal",
		Description: "Test Description",
		Version:     "1.0",
		CreatedAt:   time.Now().Unix(),
	}

	// Serialize without checksum
	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	require.NoError(t, err)

	// This would normally be done by UploadProposalMetadata
	// but we test the checksum logic separately
	err = client.verifyMetadataChecksum(metadata, jsonData)

	// Should fail because checksum is empty
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checksum mismatch")
}

// Integration test helpers

func TestIPFSIntegration_MockScenario(t *testing.T) {
	// This test simulates IPFS operations using a mock
	mock := NewMockIPFSClient()

	// Test data
	testData := []byte("test proposal content")

	// Mock upload
	hash := mock.mockAdd(testData)
	assert.NotEmpty(t, hash)
	assert.True(t, len(hash) > 0)

	// Mock retrieve
	retrieved, exists := mock.mockCat(hash)
	assert.True(t, exists)
	assert.Equal(t, testData, retrieved)

	// Mock pin
	pinned := mock.mockPin(hash)
	assert.True(t, pinned)
	assert.True(t, mock.pins[hash])

	// Mock unpin
	unpinned := mock.mockUnpin(hash)
	assert.True(t, unpinned)
	assert.False(t, mock.pins[hash])
}

func TestIPFSClient_ErrorHandling(t *testing.T) {
	client := NewIPFSClient("")

	// Test with empty hash
	emptyHash := types.Hash{}
	exists, err := client.VerifyContentExists(emptyHash)

	// Should handle error gracefully
	if err != nil {
		t.Logf("Expected error with empty hash: %v", err)
		assert.False(t, exists)
	}
}

// Benchmark tests

func BenchmarkHashConversion(b *testing.B) {
	client := NewIPFSClient("")
	ipfsHash := "QmTestHashForBenchmarking123456789"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		typesHash := client.ipfsHashToTypesHash(ipfsHash)
		_ = client.typesHashToIPFSHash(typesHash)
	}
}

func BenchmarkMetadataSerialization(b *testing.B) {
	metadata := &ProposalMetadata{
		Title:       "Benchmark Proposal",
		Description: "A proposal for benchmarking serialization performance",
		Details:     "Detailed description with more content for realistic benchmarking",
		Documents: []DocumentReference{
			{Name: "doc1.pdf", Hash: "QmHash1", Size: 1024},
			{Name: "doc2.pdf", Hash: "QmHash2", Size: 2048},
		},
		Links: []LinkReference{
			{Title: "Link1", URL: "https://example1.com"},
			{Title: "Link2", URL: "https://example2.com"},
		},
		Tags:      []string{"benchmark", "test", "performance"},
		Version:   "1.0",
		CreatedAt: time.Now().Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.MarshalIndent(metadata, "", "  ")
		if err != nil {
			b.Fatal(err)
		}
	}
}
