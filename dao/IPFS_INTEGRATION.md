# IPFS Integration for ProjectX DAO

This document describes the IPFS (InterPlanetary File System) integration implemented for the ProjectX DAO system, enabling decentralized storage of governance metadata and documents.

## Overview

The IPFS integration provides:
- **Decentralized metadata storage** for governance proposals
- **Document management** for proposal attachments
- **Content addressing and verification** for data integrity
- **Pin management** for content persistence
- **Seamless integration** with the existing DAO system

## Features

### 1. Proposal Metadata Storage
- Rich metadata structure with title, description, details, documents, links, and tags
- Automatic versioning and checksums for data integrity
- JSON serialization with proper formatting
- IPFS hash integration with ProjectX's type system

### 2. Document Management
- Upload and retrieve documents of any type
- MIME type support and size validation
- Document references with metadata
- Batch document operations

### 3. Content Verification
- Checksum validation for metadata integrity
- Content existence verification
- Size validation for documents
- Hash conversion between IPFS and ProjectX formats

### 4. Pin Management
- Automatic pinning of important content
- Cleanup of unused metadata
- Pin listing and management
- Garbage collection prevention

## Architecture

### Core Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   DAO System   │────│  IPFS Client    │────│   IPFS Node     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
    ┌────▼────┐              ┌───▼───┐              ┌────▼────┐
    │Proposals│              │Metadata│              │Documents│
    │ Manager │              │Storage │              │ Storage │
    └─────────┘              └───────┘              └─────────┘
```

### Data Structures

#### ProposalMetadata
```go
type ProposalMetadata struct {
    Title       string                 `json:"title"`
    Description string                 `json:"description"`
    Details     string                 `json:"details,omitempty"`
    Documents   []DocumentReference    `json:"documents,omitempty"`
    Links       []LinkReference        `json:"links,omitempty"`
    Tags        []string               `json:"tags,omitempty"`
    Version     string                 `json:"version"`
    CreatedAt   int64                  `json:"created_at"`
    UpdatedAt   int64                  `json:"updated_at,omitempty"`
    Checksum    string                 `json:"checksum"`
}
```

#### DocumentReference
```go
type DocumentReference struct {
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
    Hash        string `json:"hash"`
    Size        int64  `json:"size"`
    MimeType    string `json:"mime_type,omitempty"`
}
```

#### LinkReference
```go
type LinkReference struct {
    Title       string `json:"title"`
    URL         string `json:"url"`
    Description string `json:"description,omitempty"`
}
```

## Usage Examples

### 1. Creating a Proposal with Metadata

```go
// Create DAO instance
dao := NewDAO("GOVTOKEN", "Governance Token", 18)

// Prepare documents
documents := []DocumentReference{
    {
        Name:        "proposal-spec.pdf",
        Description: "Technical specification",
        Hash:        "QmExampleHash1",
        Size:        2048,
        MimeType:    "application/pdf",
    },
}

// Prepare links
links := []LinkReference{
    {
        Title:       "Community Discussion",
        URL:         "https://forum.example.com/proposal-123",
        Description: "Community discussion thread",
    },
}

// Create proposal with metadata
proposalHash, metadataHash, err := dao.CreateProposalWithMetadata(
    creator,
    "Protocol Upgrade Proposal",
    "Upgrade to version 2.0",
    "Detailed description...",
    documents,
    links,
    []string{"protocol", "upgrade"},
    ProposalTypeTechnical,
    VotingTypeQuadratic,
    startTime,
    endTime,
    6000, // 60% threshold
)
```

### 2. Uploading Documents

```go
// Upload a document
documentContent := []byte("Document content here...")
docRef, err := dao.UploadProposalDocument(
    "technical-spec.md",
    documentContent,
    "text/markdown",
)

// Retrieve the document
retrievedContent, err := dao.RetrieveProposalDocument(docRef)
```

### 3. Managing Metadata

```go
// Retrieve proposal metadata
metadata, err := dao.GetProposalMetadata(proposalID)

// Update metadata
updates := &ProposalMetadata{
    Details: "Updated details with additional information",
    Tags:    []string{"protocol", "upgrade", "audited"},
}
newMetadataHash, err := dao.UpdateProposalMetadata(proposalID, updates)

// Verify metadata exists
exists, err := dao.VerifyProposalMetadata(proposalID)
```

### 4. Content Management

```go
// List pinned content
pinnedHashes, err := dao.ListPinnedContent()

// Cleanup unused metadata
err = dao.CleanupUnusedMetadata()

// Get IPFS node info
nodeInfo, err := dao.GetIPFSNodeInfo()
```

## Configuration

### IPFS Node Setup

1. **Install IPFS**:
   ```bash
   # Download and install IPFS
   wget https://dist.ipfs.io/go-ipfs/v0.14.0/go-ipfs_v0.14.0_linux-amd64.tar.gz
   tar -xvzf go-ipfs_v0.14.0_linux-amd64.tar.gz
   cd go-ipfs
   sudo bash install.sh
   ```

2. **Initialize IPFS**:
   ```bash
   ipfs init
   ```

3. **Start IPFS Daemon**:
   ```bash
   ipfs daemon
   ```

### DAO Configuration

```go
// Default configuration (localhost:5001)
dao := NewDAO("TOKEN", "Token Name", 18)

// Custom IPFS node
dao.IPFSClient = NewIPFSClient("192.168.1.100:5001")
```

## API Reference

### IPFSClient Methods

#### Core Operations
- `UploadProposalMetadata(metadata *ProposalMetadata) (types.Hash, error)`
- `RetrieveProposalMetadata(hash types.Hash) (*ProposalMetadata, error)`
- `UploadDocument(name string, data []byte, mimeType string) (*DocumentReference, error)`
- `RetrieveDocument(docRef *DocumentReference) ([]byte, error)`

#### Content Management
- `PinContent(hash types.Hash) error`
- `UnpinContent(hash types.Hash) error`
- `VerifyContentExists(hash types.Hash) (bool, error)`
- `GetContentSize(hash types.Hash) (int64, error)`
- `ListPinnedContent() ([]types.Hash, error)`

#### Node Information
- `GetNodeInfo() (map[string]interface{}, error)`

### DAO Integration Methods

#### Proposal Operations
- `CreateProposalWithMetadata(...) (types.Hash, types.Hash, error)`
- `GetProposalMetadata(proposalID types.Hash) (*ProposalMetadata, error)`
- `UpdateProposalMetadata(proposalID types.Hash, updates *ProposalMetadata) (types.Hash, error)`
- `VerifyProposalMetadata(proposalID types.Hash) (bool, error)`

#### Document Operations
- `UploadProposalDocument(name string, data []byte, mimeType string) (*DocumentReference, error)`
- `RetrieveProposalDocument(docRef *DocumentReference) ([]byte, error)`

#### Management Operations
- `ListPinnedContent() ([]types.Hash, error)`
- `CleanupUnusedMetadata() error`
- `GetIPFSNodeInfo() (map[string]interface{}, error)`

## Testing

### Unit Tests
```bash
# Run IPFS-specific tests
go test ./dao -v -run TestIPFS

# Run all DAO tests
go test ./dao -v
```

### Integration Tests
The integration tests automatically detect if an IPFS node is available:
- **With IPFS node**: Full integration testing
- **Without IPFS node**: Mock testing with expected errors

### Example Test Output
```
=== RUN   TestDAO_IPFSIntegration
    ipfs_integration_test.go:23: IPFS node not available, skipping integration test
--- SKIP: TestDAO_IPFSIntegration (0.01s)

=== RUN   TestIPFSClient_HashConversion
--- PASS: TestIPFSClient_HashConversion (0.00s)
```

## Security Considerations

### Data Integrity
- **Checksums**: All metadata includes SHA-256 checksums
- **Verification**: Content existence and integrity checks
- **Versioning**: Automatic version tracking for updates

### Access Control
- **Pin Management**: Only authorized operations can pin/unpin content
- **Content Validation**: Size and format validation for uploads
- **Error Handling**: Graceful handling of network failures

### Privacy
- **Public Network**: IPFS is a public network - sensitive data should be encrypted
- **Content Addressing**: Content is addressed by hash, providing some privacy
- **Pinning Strategy**: Important content is pinned to prevent loss

## Performance Considerations

### Optimization Strategies
- **Lazy Loading**: Metadata is loaded on-demand
- **Caching**: Local caching of frequently accessed content
- **Batch Operations**: Support for batch uploads and retrievals
- **Cleanup**: Automatic cleanup of unused content

### Scalability
- **Distributed Storage**: IPFS provides natural scalability
- **Content Deduplication**: Identical content is stored once
- **Pin Management**: Strategic pinning prevents storage bloat

## Troubleshooting

### Common Issues

1. **IPFS Node Not Running**
   ```
   Error: dial tcp [::1]:5001: connectex: No connection could be made
   ```
   **Solution**: Start IPFS daemon with `ipfs daemon`

2. **Content Not Found**
   ```
   Error: failed to retrieve from IPFS: not found
   ```
   **Solution**: Verify content hash and ensure it's pinned

3. **Network Connectivity**
   ```
   Error: context deadline exceeded
   ```
   **Solution**: Check network connectivity and increase timeout

### Debug Mode
Enable debug logging for detailed IPFS operations:
```go
client := NewIPFSClient("localhost:5001")
// Debug logging would be configured here in production
```

## Future Enhancements

### Planned Features
- **Encryption**: Client-side encryption for sensitive data
- **Compression**: Automatic compression for large documents
- **CDN Integration**: Content delivery network for faster access
- **Backup Strategies**: Multi-node pinning for redundancy

### Performance Improvements
- **Connection Pooling**: Reuse IPFS connections
- **Async Operations**: Non-blocking uploads and downloads
- **Streaming**: Support for large file streaming
- **Caching Layer**: Redis/Memcached integration

## Dependencies

- `github.com/ipfs/go-ipfs-api`: IPFS HTTP API client
- `github.com/BOCK-CHAIN/BockChain/types`: BockDAO type system
- `github.com/BOCK-CHAIN/BockChain/crypto`: Cryptographic utilities

## License

This IPFS integration is part of the ProjectX DAO system and follows the same license terms as the main project.