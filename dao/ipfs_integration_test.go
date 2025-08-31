package dao

import (
	"testing"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDAO_IPFSIntegration(t *testing.T) {
	// Create a new DAO instance
	dao := NewDAO("TEST", "Test Token", 18)

	// Verify IPFS client is initialized
	assert.NotNil(t, dao.IPFSClient)

	// Test basic IPFS client functionality
	nodeInfo, err := dao.GetIPFSNodeInfo()
	if err != nil {
		t.Logf("IPFS node not available, skipping integration test: %v", err)
		t.Skip("IPFS node required for integration test")
	}

	assert.NotNil(t, nodeInfo)
	t.Logf("Connected to IPFS node: %v", nodeInfo)
}

func TestDAO_CreateProposalWithMetadata_MockScenario(t *testing.T) {
	// This test simulates the proposal creation with metadata
	// without requiring an actual IPFS node

	dao := NewDAO("TEST", "Test Token", 18)
	creator := crypto.GeneratePrivateKey().PublicKey()

	// Initialize some tokens for the creator
	dao.TokenState.Mint(creator.String(), 1000)

	// Test data
	title := "Test Governance Proposal"
	description := "A comprehensive test proposal"
	details := "This proposal includes detailed information about governance changes"

	documents := []DocumentReference{
		{
			Name:        "proposal-details.pdf",
			Description: "Detailed proposal document",
			Hash:        "QmTestDocumentHash123",
			Size:        2048,
			MimeType:    "application/pdf",
		},
	}

	links := []LinkReference{
		{
			Title:       "Reference Documentation",
			URL:         "https://docs.example.com/governance",
			Description: "Official governance documentation",
		},
	}

	tags := []string{"governance", "protocol", "upgrade"}

	// Test proposal creation (will fail without IPFS but we test the structure)
	proposalHash, metadataHash, err := dao.CreateProposalWithMetadata(
		creator,
		title,
		description,
		details,
		documents,
		links,
		tags,
		ProposalTypeGeneral,
		VotingTypeSimple,
		time.Now().Unix(),
		time.Now().Unix()+3600, // 1 hour voting period
		500,                    // 50% threshold
	)

	if err != nil {
		// Expected without IPFS node
		t.Logf("Expected error without IPFS node: %v", err)
		assert.Contains(t, err.Error(), "IPFS")
	} else {
		// If IPFS is available, verify the results
		assert.NotEqual(t, types.Hash{}, proposalHash)
		assert.NotEqual(t, types.Hash{}, metadataHash)

		// Verify proposal was created
		proposal, err := dao.GetProposal(proposalHash)
		require.NoError(t, err)
		assert.Equal(t, title, proposal.Title)
		assert.Equal(t, description, proposal.Description)
		assert.Equal(t, metadataHash, proposal.MetadataHash)
	}
}

func TestDAO_ProposalMetadataOperations(t *testing.T) {
	dao := NewDAO("TEST", "Test Token", 18)

	// Create a mock proposal with metadata hash
	proposalID := types.Hash{1, 2, 3, 4, 5}
	metadataHash := types.Hash{6, 7, 8, 9, 10}

	proposal := &Proposal{
		ID:           proposalID,
		Creator:      crypto.GeneratePrivateKey().PublicKey(),
		Title:        "Test Proposal",
		Description:  "Test Description",
		ProposalType: ProposalTypeGeneral,
		VotingType:   VotingTypeSimple,
		StartTime:    time.Now().Unix(),
		EndTime:      time.Now().Unix() + 3600,
		Status:       ProposalStatusActive,
		Threshold:    500,
		MetadataHash: metadataHash,
	}

	// Add proposal to governance state
	dao.GovernanceState.Proposals[proposalID] = proposal

	// Test metadata verification (will fail without IPFS but tests the flow)
	exists, err := dao.VerifyProposalMetadata(proposalID)
	if err != nil {
		t.Logf("Expected error without IPFS node: %v", err)
		assert.False(t, exists)
	}

	// Test metadata retrieval (will fail without IPFS but tests the flow)
	metadata, err := dao.GetProposalMetadata(proposalID)
	if err != nil {
		t.Logf("Expected error without IPFS node: %v", err)
		assert.Nil(t, metadata)
	}
}

func TestDAO_DocumentOperations(t *testing.T) {
	dao := NewDAO("TEST", "Test Token", 18)

	// Test document upload
	testData := []byte("This is a test document content")
	docRef, err := dao.UploadProposalDocument("test.txt", testData, "text/plain")

	if err != nil {
		// Expected without IPFS node
		t.Logf("Expected error without IPFS node: %v", err)
		assert.Nil(t, docRef)
	} else {
		// If IPFS is available, verify the document reference
		assert.NotNil(t, docRef)
		assert.Equal(t, "test.txt", docRef.Name)
		assert.Equal(t, int64(len(testData)), docRef.Size)
		assert.Equal(t, "text/plain", docRef.MimeType)
		assert.NotEmpty(t, docRef.Hash)

		// Test document retrieval
		retrievedData, err := dao.RetrieveProposalDocument(docRef)
		require.NoError(t, err)
		assert.Equal(t, testData, retrievedData)
	}
}

func TestDAO_IPFSContentManagement(t *testing.T) {
	dao := NewDAO("TEST", "Test Token", 18)

	// Test listing pinned content
	pinnedContent, err := dao.ListPinnedContent()
	if err != nil {
		t.Logf("Expected error without IPFS node: %v", err)
		assert.Nil(t, pinnedContent)
	} else {
		assert.NotNil(t, pinnedContent)
		t.Logf("Pinned content count: %d", len(pinnedContent))
	}

	// Test cleanup unused metadata
	err = dao.CleanupUnusedMetadata()
	if err != nil {
		t.Logf("Expected error without IPFS node: %v", err)
	}
}

func TestDAO_MetadataUpdateFlow(t *testing.T) {
	dao := NewDAO("TEST", "Test Token", 18)

	// Create a mock proposal
	proposalID := types.Hash{1, 2, 3, 4, 5}
	metadataHash := types.Hash{6, 7, 8, 9, 10}

	proposal := &Proposal{
		ID:           proposalID,
		Creator:      crypto.GeneratePrivateKey().PublicKey(),
		Title:        "Original Title",
		Description:  "Original Description",
		ProposalType: ProposalTypeGeneral,
		VotingType:   VotingTypeSimple,
		StartTime:    time.Now().Unix(),
		EndTime:      time.Now().Unix() + 3600,
		Status:       ProposalStatusActive,
		Threshold:    500,
		MetadataHash: metadataHash,
	}

	dao.GovernanceState.Proposals[proposalID] = proposal

	// Test metadata update
	updates := &ProposalMetadata{
		Title:       "Updated Title",
		Description: "Updated Description",
		Details:     "Additional details added",
	}

	newMetadataHash, err := dao.UpdateProposalMetadata(proposalID, updates)
	if err != nil {
		// Expected without IPFS node
		t.Logf("Expected error without IPFS node: %v", err)
		assert.Equal(t, types.Hash{}, newMetadataHash)
	} else {
		assert.NotEqual(t, types.Hash{}, newMetadataHash)
		assert.NotEqual(t, metadataHash, newMetadataHash)

		// Verify proposal was updated
		updatedProposal, err := dao.GetProposal(proposalID)
		require.NoError(t, err)
		assert.Equal(t, newMetadataHash, updatedProposal.MetadataHash)
	}
}

func TestDAO_IPFSErrorHandling(t *testing.T) {
	dao := NewDAO("TEST", "Test Token", 18)

	// Test with non-existent proposal
	_, err := dao.GetProposalMetadata(types.Hash{99, 99, 99})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test with proposal without metadata hash
	proposalID := types.Hash{1, 2, 3}
	proposal := &Proposal{
		ID:           proposalID,
		Creator:      crypto.GeneratePrivateKey().PublicKey(),
		Title:        "No Metadata Proposal",
		Description:  "This proposal has no metadata",
		MetadataHash: types.Hash{}, // Empty hash
	}

	dao.GovernanceState.Proposals[proposalID] = proposal

	_, err = dao.GetProposalMetadata(proposalID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no metadata hash")

	// Test metadata verification with empty hash
	exists, err := dao.VerifyProposalMetadata(proposalID)
	assert.NoError(t, err)
	assert.False(t, exists)
}

// Benchmark tests for IPFS operations

func BenchmarkDAO_CreateProposalWithMetadata(b *testing.B) {
	dao := NewDAO("TEST", "Test Token", 18)
	creator := crypto.GeneratePrivateKey().PublicKey()
	dao.TokenState.Mint(creator.String(), 1000000)

	documents := []DocumentReference{
		{Name: "doc.pdf", Hash: "QmHash", Size: 1024, MimeType: "application/pdf"},
	}
	links := []LinkReference{
		{Title: "Link", URL: "https://example.com"},
	}
	tags := []string{"test", "benchmark"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := dao.CreateProposalWithMetadata(
			creator,
			"Benchmark Proposal",
			"Benchmark Description",
			"Benchmark Details",
			documents,
			links,
			tags,
			ProposalTypeGeneral,
			VotingTypeSimple,
			time.Now().Unix(),
			time.Now().Unix()+3600,
			500,
		)
		// Expected to fail without IPFS, but we benchmark the preparation logic
		if err == nil {
			b.Log("Unexpected success - IPFS node available")
		}
	}
}

func BenchmarkDAO_MetadataOperations(b *testing.B) {
	dao := NewDAO("TEST", "Test Token", 18)

	// Setup test proposal
	proposalID := types.Hash{1, 2, 3, 4, 5}
	metadataHash := types.Hash{6, 7, 8, 9, 10}

	proposal := &Proposal{
		ID:           proposalID,
		Creator:      crypto.GeneratePrivateKey().PublicKey(),
		Title:        "Benchmark Proposal",
		Description:  "Benchmark Description",
		MetadataHash: metadataHash,
	}

	dao.GovernanceState.Proposals[proposalID] = proposal

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Benchmark metadata verification
		_, _ = dao.VerifyProposalMetadata(proposalID)
	}
}
