package dao

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/BOCK-CHAIN/BockChain/types"
	shell "github.com/ipfs/go-ipfs-api"
)

// IPFSClient wraps the IPFS shell client with DAO-specific functionality
type IPFSClient struct {
	shell   *shell.Shell
	timeout time.Duration
}

// NewIPFSClient creates a new IPFS client instance
func NewIPFSClient(nodeURL string) *IPFSClient {
	if nodeURL == "" {
		nodeURL = "localhost:5001" // Default IPFS API endpoint
	}

	return &IPFSClient{
		shell:   shell.NewShell(nodeURL),
		timeout: 30 * time.Second,
	}
}

// ProposalMetadata represents the metadata structure for proposals
type ProposalMetadata struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Details     string              `json:"details,omitempty"`
	Documents   []DocumentReference `json:"documents,omitempty"`
	Links       []LinkReference     `json:"links,omitempty"`
	Tags        []string            `json:"tags,omitempty"`
	Version     string              `json:"version"`
	CreatedAt   int64               `json:"created_at"`
	UpdatedAt   int64               `json:"updated_at,omitempty"`
	Checksum    string              `json:"checksum"`
}

// DocumentReference represents a reference to a document stored on IPFS
type DocumentReference struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Hash        string `json:"hash"`
	Size        int64  `json:"size"`
	MimeType    string `json:"mime_type,omitempty"`
}

// LinkReference represents an external link reference
type LinkReference struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// UploadProposalMetadata uploads proposal metadata to IPFS
func (c *IPFSClient) UploadProposalMetadata(metadata *ProposalMetadata) (types.Hash, error) {

	// Set timestamps
	now := time.Now().Unix()
	metadata.CreatedAt = now
	if metadata.UpdatedAt == 0 {
		metadata.UpdatedAt = now
	}
	if metadata.Version == "" {
		metadata.Version = "1.0"
	}

	// Serialize metadata to JSON
	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return types.Hash{}, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Calculate checksum
	hash := sha256.Sum256(jsonData)
	metadata.Checksum = hex.EncodeToString(hash[:])

	// Re-serialize with checksum
	jsonData, err = json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return types.Hash{}, fmt.Errorf("failed to marshal metadata with checksum: %w", err)
	}

	// Upload to IPFS
	reader := bytes.NewReader(jsonData)
	ipfsHash, err := c.shell.Add(reader)
	if err != nil {
		return types.Hash{}, fmt.Errorf("failed to upload to IPFS: %w", err)
	}

	// Convert IPFS hash to types.Hash
	return c.ipfsHashToTypesHash(ipfsHash), nil
}

// RetrieveProposalMetadata retrieves proposal metadata from IPFS
func (c *IPFSClient) RetrieveProposalMetadata(hash types.Hash) (*ProposalMetadata, error) {

	ipfsHash := c.typesHashToIPFSHash(hash)

	reader, err := c.shell.Cat(ipfsHash)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve from IPFS: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read IPFS data: %w", err)
	}

	var metadata ProposalMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	// Verify checksum
	if err := c.verifyMetadataChecksum(&metadata, data); err != nil {
		return nil, fmt.Errorf("metadata verification failed: %w", err)
	}

	return &metadata, nil
}

// UploadDocument uploads a document to IPFS and returns its reference
func (c *IPFSClient) UploadDocument(name string, data []byte, mimeType string) (*DocumentReference, error) {

	reader := bytes.NewReader(data)
	ipfsHash, err := c.shell.Add(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to upload document to IPFS: %w", err)
	}

	return &DocumentReference{
		Name:     name,
		Hash:     ipfsHash,
		Size:     int64(len(data)),
		MimeType: mimeType,
	}, nil
}

// RetrieveDocument retrieves a document from IPFS
func (c *IPFSClient) RetrieveDocument(docRef *DocumentReference) ([]byte, error) {

	reader, err := c.shell.Cat(docRef.Hash)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve document from IPFS: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read document data: %w", err)
	}

	// Verify size if specified
	if docRef.Size > 0 && int64(len(data)) != docRef.Size {
		return nil, fmt.Errorf("document size mismatch: expected %d, got %d", docRef.Size, len(data))
	}

	return data, nil
}

// PinContent pins content to prevent garbage collection
func (c *IPFSClient) PinContent(hash types.Hash) error {

	ipfsHash := c.typesHashToIPFSHash(hash)
	return c.shell.Pin(ipfsHash)
}

// UnpinContent unpins content to allow garbage collection
func (c *IPFSClient) UnpinContent(hash types.Hash) error {

	ipfsHash := c.typesHashToIPFSHash(hash)
	return c.shell.Unpin(ipfsHash)
}

// GetContentSize returns the size of content stored at the given hash
func (c *IPFSClient) GetContentSize(hash types.Hash) (int64, error) {

	ipfsHash := c.typesHashToIPFSHash(hash)
	stat, err := c.shell.ObjectStat(ipfsHash)
	if err != nil {
		return 0, fmt.Errorf("failed to get content size: %w", err)
	}

	return int64(stat.CumulativeSize), nil
}

// VerifyContentExists checks if content exists on IPFS
func (c *IPFSClient) VerifyContentExists(hash types.Hash) (bool, error) {

	ipfsHash := c.typesHashToIPFSHash(hash)
	_, err := c.shell.ObjectStat(ipfsHash)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, fmt.Errorf("failed to verify content existence: %w", err)
	}

	return true, nil
}

// ListPinnedContent returns a list of all pinned content hashes
func (c *IPFSClient) ListPinnedContent() ([]types.Hash, error) {

	pins, err := c.shell.Pins()
	if err != nil {
		return nil, fmt.Errorf("failed to list pinned content: %w", err)
	}

	var hashes []types.Hash
	for ipfsHash := range pins {
		hash := c.ipfsHashToTypesHash(ipfsHash)
		hashes = append(hashes, hash)
	}

	return hashes, nil
}

// GetNodeInfo returns information about the connected IPFS node
func (c *IPFSClient) GetNodeInfo() (map[string]interface{}, error) {

	id, err := c.shell.ID()
	if err != nil {
		return nil, fmt.Errorf("failed to get node info: %w", err)
	}

	return map[string]interface{}{
		"id":               id.ID,
		"public_key":       id.PublicKey,
		"addresses":        id.Addresses,
		"agent_version":    id.AgentVersion,
		"protocol_version": id.ProtocolVersion,
	}, nil
}

// Helper functions

// ipfsHashToTypesHash converts an IPFS hash string to types.Hash
func (c *IPFSClient) ipfsHashToTypesHash(ipfsHash string) types.Hash {
	// For now, we'll use the first 32 bytes of the IPFS hash
	// In a production system, you might want a more sophisticated mapping
	hashBytes := sha256.Sum256([]byte(ipfsHash))
	var hash types.Hash
	copy(hash[:], hashBytes[:])
	return hash
}

// typesHashToIPFSHash converts a types.Hash to an IPFS hash string
func (c *IPFSClient) typesHashToIPFSHash(hash types.Hash) string {
	// This is a simplified conversion - in practice, you'd want to store
	// the actual IPFS hash and use this as a lookup key
	return hex.EncodeToString(hash[:])
}

// verifyMetadataChecksum verifies the checksum of metadata
func (c *IPFSClient) verifyMetadataChecksum(metadata *ProposalMetadata, data []byte) error {
	// Create a copy without checksum for verification
	temp := *metadata
	temp.Checksum = ""

	tempData, err := json.MarshalIndent(&temp, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal for checksum verification: %w", err)
	}

	hash := sha256.Sum256(tempData)
	expectedChecksum := hex.EncodeToString(hash[:])

	if metadata.Checksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, metadata.Checksum)
	}

	return nil
}

// CreateProposalWithIPFS creates a proposal with IPFS metadata storage
func (c *IPFSClient) CreateProposalWithIPFS(title, description, details string, documents []DocumentReference, links []LinkReference, tags []string) (*ProposalMetadata, types.Hash, error) {
	metadata := &ProposalMetadata{
		Title:       title,
		Description: description,
		Details:     details,
		Documents:   documents,
		Links:       links,
		Tags:        tags,
		Version:     "1.0",
	}

	hash, err := c.UploadProposalMetadata(metadata)
	if err != nil {
		return nil, types.Hash{}, err
	}

	return metadata, hash, nil
}

// UpdateProposalMetadata updates existing proposal metadata with a new version
func (c *IPFSClient) UpdateProposalMetadata(existingHash types.Hash, updates *ProposalMetadata) (types.Hash, error) {
	// Retrieve existing metadata
	existing, err := c.RetrieveProposalMetadata(existingHash)
	if err != nil {
		return types.Hash{}, fmt.Errorf("failed to retrieve existing metadata: %w", err)
	}

	// Update fields
	if updates.Title != "" {
		existing.Title = updates.Title
	}
	if updates.Description != "" {
		existing.Description = updates.Description
	}
	if updates.Details != "" {
		existing.Details = updates.Details
	}
	if len(updates.Documents) > 0 {
		existing.Documents = updates.Documents
	}
	if len(updates.Links) > 0 {
		existing.Links = updates.Links
	}
	if len(updates.Tags) > 0 {
		existing.Tags = updates.Tags
	}

	// Increment version
	existing.Version = fmt.Sprintf("%.1f", parseVersion(existing.Version)+0.1)
	existing.UpdatedAt = time.Now().Unix()

	// Upload updated metadata
	return c.UploadProposalMetadata(existing)
}

// parseVersion parses a version string to float64
func parseVersion(version string) float64 {
	var v float64
	fmt.Sscanf(version, "%f", &v)
	return v
}
