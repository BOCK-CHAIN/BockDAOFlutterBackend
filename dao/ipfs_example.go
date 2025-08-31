package dao

import (
	"fmt"
	"log"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
	"github.com/BOCK-CHAIN/BockChain/types"
)

// ExampleIPFSIntegration demonstrates how to use IPFS integration with the DAO
func ExampleIPFSIntegration() {
	// Create a new DAO instance with IPFS integration
	dao := NewDAO("GOVTOKEN", "Governance Token", 18)

	// Create a user with some tokens
	userPrivKey := crypto.GeneratePrivateKey()
	userPubKey := userPrivKey.PublicKey()

	// Mint some tokens for the user
	err := dao.MintTokens(userPubKey, 10000)
	if err != nil {
		log.Printf("Failed to mint tokens: %v", err)
		return
	}

	fmt.Printf("User %s has %d tokens\n", userPubKey.String()[:8], dao.GetTokenBalance(userPubKey))

	// Example 1: Create a proposal with rich metadata
	fmt.Println("\n=== Creating Proposal with IPFS Metadata ===")

	// Prepare proposal documents
	documents := []DocumentReference{
		{
			Name:        "proposal-specification.pdf",
			Description: "Technical specification document",
			Hash:        "QmExampleHash1", // In practice, this would be uploaded first
			Size:        2048,
			MimeType:    "application/pdf",
		},
		{
			Name:        "budget-breakdown.xlsx",
			Description: "Detailed budget breakdown",
			Hash:        "QmExampleHash2",
			Size:        1024,
			MimeType:    "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		},
	}

	// Prepare reference links
	links := []LinkReference{
		{
			Title:       "Community Discussion",
			URL:         "https://forum.example.com/proposal-123",
			Description: "Community discussion thread",
		},
		{
			Title:       "Technical Documentation",
			URL:         "https://docs.example.com/technical-spec",
			Description: "Related technical documentation",
		},
	}

	// Proposal tags for categorization
	tags := []string{"protocol-upgrade", "treasury", "technical", "high-priority"}

	// Create proposal with metadata
	proposalHash, metadataHash, err := dao.CreateProposalWithMetadata(
		userPubKey,
		"Protocol Upgrade Proposal v2.1",
		"Proposal to upgrade the protocol to version 2.1 with enhanced security features",
		`This proposal outlines a comprehensive upgrade to our protocol that includes:
		
1. Enhanced cryptographic security
2. Improved transaction throughput
3. Better governance mechanisms
4. Reduced gas costs

The upgrade has been thoroughly tested on testnet and reviewed by security auditors.
Implementation timeline is 4 weeks from approval.`,
		documents,
		links,
		tags,
		ProposalTypeTechnical,
		VotingTypeQuadratic, // Use quadratic voting for technical proposals
		time.Now().Unix(),
		time.Now().Unix()+7*24*3600, // 7 days voting period
		6000,                        // 60% approval threshold
	)

	if err != nil {
		log.Printf("Failed to create proposal (expected without IPFS node): %v", err)
	} else {
		fmt.Printf("Created proposal: %x\n", proposalHash)
		fmt.Printf("Metadata stored at: %x\n", metadataHash)
	}

	// Example 2: Upload and manage documents
	fmt.Println("\n=== Document Management ===")

	// Upload a document
	documentContent := []byte(`
# Proposal Technical Specification

## Overview
This document outlines the technical details of the proposed protocol upgrade.

## Changes
1. Cryptographic improvements
2. Performance optimizations
3. Security enhancements

## Implementation Plan
- Phase 1: Core protocol changes
- Phase 2: Client updates
- Phase 3: Network migration
`)

	docRef, err := dao.UploadProposalDocument(
		"technical-spec.md",
		documentContent,
		"text/markdown",
	)

	if err != nil {
		log.Printf("Failed to upload document (expected without IPFS node): %v", err)
	} else {
		fmt.Printf("Uploaded document: %s (Hash: %s, Size: %d bytes)\n",
			docRef.Name, docRef.Hash, docRef.Size)

		// Retrieve the document
		retrievedContent, err := dao.RetrieveProposalDocument(docRef)
		if err != nil {
			log.Printf("Failed to retrieve document: %v", err)
		} else {
			fmt.Printf("Retrieved document content (%d bytes)\n", len(retrievedContent))
		}
	}

	// Example 3: Metadata operations
	fmt.Println("\n=== Metadata Operations ===")

	// Check IPFS node status
	nodeInfo, err := dao.GetIPFSNodeInfo()
	if err != nil {
		log.Printf("IPFS node not available: %v", err)
	} else {
		fmt.Printf("Connected to IPFS node: %s\n", nodeInfo["id"])
	}

	// List pinned content
	pinnedContent, err := dao.ListPinnedContent()
	if err != nil {
		log.Printf("Failed to list pinned content: %v", err)
	} else {
		fmt.Printf("Pinned content items: %d\n", len(pinnedContent))
	}

	// Example 4: Metadata updates
	fmt.Println("\n=== Updating Proposal Metadata ===")

	if proposalHash != (types.Hash{}) {
		// Update proposal metadata
		updates := &ProposalMetadata{
			Details: `Updated details with additional information:
			
## Security Audit Results
The proposal has been audited by three independent security firms:
- Audit Firm A: No critical issues found
- Audit Firm B: Minor recommendations implemented
- Audit Firm C: Full approval granted

## Community Feedback
Based on community feedback, we have made the following adjustments:
- Reduced implementation timeline to 3 weeks
- Added additional testing phases
- Improved documentation`,
			Tags: []string{"protocol-upgrade", "treasury", "technical", "high-priority", "audited", "community-approved"},
		}

		newMetadataHash, err := dao.UpdateProposalMetadata(proposalHash, updates)
		if err != nil {
			log.Printf("Failed to update metadata: %v", err)
		} else {
			fmt.Printf("Updated metadata hash: %x\n", newMetadataHash)
		}
	}

	// Example 5: Content verification
	fmt.Println("\n=== Content Verification ===")

	if proposalHash != (types.Hash{}) {
		exists, err := dao.VerifyProposalMetadata(proposalHash)
		if err != nil {
			log.Printf("Failed to verify metadata: %v", err)
		} else {
			fmt.Printf("Metadata exists and is accessible: %t\n", exists)
		}

		// Retrieve full metadata
		metadata, err := dao.GetProposalMetadata(proposalHash)
		if err != nil {
			log.Printf("Failed to retrieve metadata: %v", err)
		} else {
			fmt.Printf("Retrieved metadata:\n")
			fmt.Printf("  Title: %s\n", metadata.Title)
			fmt.Printf("  Version: %s\n", metadata.Version)
			fmt.Printf("  Documents: %d\n", len(metadata.Documents))
			fmt.Printf("  Links: %d\n", len(metadata.Links))
			fmt.Printf("  Tags: %v\n", metadata.Tags)
		}
	}

	// Example 6: Cleanup operations
	fmt.Println("\n=== Cleanup Operations ===")

	err = dao.CleanupUnusedMetadata()
	if err != nil {
		log.Printf("Failed to cleanup unused metadata: %v", err)
	} else {
		fmt.Println("Cleanup completed successfully")
	}

	fmt.Println("\n=== IPFS Integration Example Complete ===")
}

// ExampleIPFSClientUsage demonstrates direct IPFS client usage
func ExampleIPFSClientUsage() {
	// Create IPFS client directly
	ipfsClient := NewIPFSClient("localhost:5001")

	// Create sample metadata
	metadata := &ProposalMetadata{
		Title:       "Direct IPFS Usage Example",
		Description: "Demonstrating direct IPFS client usage",
		Details:     "This example shows how to use the IPFS client directly",
		Documents: []DocumentReference{
			{
				Name:     "example.txt",
				Hash:     "QmExampleDirectHash",
				Size:     100,
				MimeType: "text/plain",
			},
		},
		Links: []LinkReference{
			{
				Title: "IPFS Documentation",
				URL:   "https://docs.ipfs.io",
			},
		},
		Tags:    []string{"example", "ipfs", "direct"},
		Version: "1.0",
	}

	// Upload metadata
	hash, err := ipfsClient.UploadProposalMetadata(metadata)
	if err != nil {
		log.Printf("Failed to upload metadata: %v", err)
		return
	}

	fmt.Printf("Uploaded metadata with hash: %x\n", hash)

	// Pin the content
	err = ipfsClient.PinContent(hash)
	if err != nil {
		log.Printf("Failed to pin content: %v", err)
	} else {
		fmt.Println("Content pinned successfully")
	}

	// Retrieve metadata
	retrievedMetadata, err := ipfsClient.RetrieveProposalMetadata(hash)
	if err != nil {
		log.Printf("Failed to retrieve metadata: %v", err)
		return
	}

	fmt.Printf("Retrieved metadata: %s (Version: %s)\n",
		retrievedMetadata.Title, retrievedMetadata.Version)

	// Verify content exists
	exists, err := ipfsClient.VerifyContentExists(hash)
	if err != nil {
		log.Printf("Failed to verify content: %v", err)
	} else {
		fmt.Printf("Content exists: %t\n", exists)
	}

	// Get content size
	size, err := ipfsClient.GetContentSize(hash)
	if err != nil {
		log.Printf("Failed to get content size: %v", err)
	} else {
		fmt.Printf("Content size: %d bytes\n", size)
	}
}

// ExampleBatchOperations demonstrates batch operations with IPFS
func ExampleBatchOperations() {
	dao := NewDAO("BATCH", "Batch Token", 18)
	creator := crypto.GeneratePrivateKey().PublicKey()
	dao.MintTokens(creator, 50000)

	fmt.Println("=== Batch IPFS Operations Example ===")

	// Create multiple proposals with metadata
	proposals := []struct {
		title        string
		description  string
		proposalType ProposalType
		votingType   VotingType
	}{
		{
			"Treasury Allocation Q1",
			"Quarterly treasury allocation for development",
			ProposalTypeTreasury,
			VotingTypeSimple,
		},
		{
			"Protocol Security Update",
			"Critical security update for the protocol",
			ProposalTypeTechnical,
			VotingTypeReputation,
		},
		{
			"Community Governance Rules",
			"Updated governance rules based on community feedback",
			ProposalTypeGeneral,
			VotingTypeQuadratic,
		},
	}

	var createdProposals []types.Hash

	for i, p := range proposals {
		tags := []string{fmt.Sprintf("batch-%d", i), "automated"}

		proposalHash, metadataHash, err := dao.CreateProposalWithMetadata(
			creator,
			p.title,
			p.description,
			fmt.Sprintf("Detailed description for proposal %d", i+1),
			[]DocumentReference{},
			[]LinkReference{},
			tags,
			p.proposalType,
			p.votingType,
			time.Now().Unix(),
			time.Now().Unix()+24*3600, // 24 hours
			5000,                      // 50% threshold
		)

		if err != nil {
			log.Printf("Failed to create proposal %d: %v", i+1, err)
			continue
		}

		createdProposals = append(createdProposals, proposalHash)
		fmt.Printf("Created proposal %d: %x (Metadata: %x)\n",
			i+1, proposalHash, metadataHash)
	}

	fmt.Printf("Successfully created %d proposals\n", len(createdProposals))

	// Verify all proposals
	fmt.Println("\n=== Verifying Proposals ===")
	for i, proposalHash := range createdProposals {
		exists, err := dao.VerifyProposalMetadata(proposalHash)
		if err != nil {
			log.Printf("Failed to verify proposal %d: %v", i+1, err)
		} else {
			fmt.Printf("Proposal %d metadata exists: %t\n", i+1, exists)
		}
	}
}
