package dao

import (
	"fmt"
	"log"

	"github.com/BOCK-CHAIN/BockChain/crypto"
)

// TokenExample demonstrates the governance token system functionality
func TokenExample() {
	fmt.Println("=== ProjectX DAO Governance Token System Example ===")

	// Create a new DAO
	dao := NewDAO("PXGOV", "ProjectX Governance Token", 18)
	fmt.Printf("Created DAO with token: %s (%s)\n", dao.TokenState.Name, dao.TokenState.Symbol)

	// Create some test addresses
	founder := crypto.GeneratePrivateKey().PublicKey()
	developer := crypto.GeneratePrivateKey().PublicKey()
	community := crypto.GeneratePrivateKey().PublicKey()
	treasury := crypto.GeneratePrivateKey().PublicKey()

	fmt.Println("\n--- Initial Token Distribution ---")

	// Initial token distribution
	distributions := map[string]uint64{
		founder.String():   50000, // 50,000 tokens for founder
		developer.String(): 30000, // 30,000 tokens for developer
		community.String(): 20000, // 20,000 tokens for community
	}

	err := dao.InitialTokenDistribution(distributions)
	if err != nil {
		log.Fatalf("Failed to distribute initial tokens: %v", err)
	}

	fmt.Printf("Total Supply: %d tokens\n", dao.GetTotalSupply())
	fmt.Printf("Founder Balance: %d tokens\n", dao.GetTokenBalance(founder))
	fmt.Printf("Developer Balance: %d tokens\n", dao.GetTokenBalance(developer))
	fmt.Printf("Community Balance: %d tokens\n", dao.GetTokenBalance(community))

	fmt.Println("\n--- Token Transfer Example ---")

	// Transfer tokens from founder to treasury
	transferTx := &TokenTransferTx{
		Fee:       100,
		Recipient: treasury,
		Amount:    10000,
	}

	err = dao.Processor.ProcessTokenTransferTx(transferTx, founder)
	if err != nil {
		log.Fatalf("Failed to transfer tokens: %v", err)
	}

	fmt.Printf("After transfer - Founder Balance: %d tokens\n", dao.GetTokenBalance(founder))
	fmt.Printf("After transfer - Treasury Balance: %d tokens\n", dao.GetTokenBalance(treasury))

	fmt.Println("\n--- Token Approval and TransferFrom Example ---")

	// Developer approves community to spend 5000 tokens
	approveTx := &TokenApproveTx{
		Fee:     50,
		Spender: community,
		Amount:  5000,
	}

	err = dao.Processor.ProcessTokenApproveTx(approveTx, developer)
	if err != nil {
		log.Fatalf("Failed to approve tokens: %v", err)
	}

	fmt.Printf("Allowance from Developer to Community: %d tokens\n",
		dao.GetTokenAllowance(developer, community))

	// Community uses allowance to transfer tokens to treasury
	transferFromTx := &TokenTransferFromTx{
		Fee:       75,
		From:      developer,
		Recipient: treasury,
		Amount:    3000,
	}

	err = dao.Processor.ProcessTokenTransferFromTx(transferFromTx, community)
	if err != nil {
		log.Fatalf("Failed to transferFrom tokens: %v", err)
	}

	fmt.Printf("After transferFrom - Developer Balance: %d tokens\n", dao.GetTokenBalance(developer))
	fmt.Printf("After transferFrom - Treasury Balance: %d tokens\n", dao.GetTokenBalance(treasury))
	fmt.Printf("After transferFrom - Community Balance: %d tokens\n", dao.GetTokenBalance(community))
	fmt.Printf("Remaining Allowance: %d tokens\n", dao.GetTokenAllowance(developer, community))

	fmt.Println("\n--- Token Minting Example ---")

	// Mint additional tokens for ecosystem rewards
	mintTx := &TokenMintTx{
		Fee:       200,
		Recipient: treasury,
		Amount:    15000,
		Reason:    "Ecosystem development rewards",
	}

	err = dao.Processor.ProcessTokenMintTx(mintTx, founder)
	if err != nil {
		log.Fatalf("Failed to mint tokens: %v", err)
	}

	fmt.Printf("After minting - Total Supply: %d tokens\n", dao.GetTotalSupply())
	fmt.Printf("After minting - Treasury Balance: %d tokens\n", dao.GetTokenBalance(treasury))

	fmt.Println("\n--- Token Burning Example ---")

	// Burn tokens for deflationary mechanism
	burnTx := &TokenBurnTx{
		Fee:    100,
		Amount: 5000,
		Reason: "Deflationary burn mechanism",
	}

	err = dao.Processor.ProcessTokenBurnTx(burnTx, treasury)
	if err != nil {
		log.Fatalf("Failed to burn tokens: %v", err)
	}

	fmt.Printf("After burning - Total Supply: %d tokens\n", dao.GetTotalSupply())
	fmt.Printf("After burning - Treasury Balance: %d tokens\n", dao.GetTokenBalance(treasury))

	fmt.Println("\n--- Final Token Holder Summary ---")

	// Display all token holders
	holders := []crypto.PublicKey{founder, developer, community, treasury}
	names := []string{"Founder", "Developer", "Community", "Treasury"}

	for i, holder := range holders {
		balance := dao.GetTokenBalance(holder)
		if tokenHolder, exists := dao.GetTokenHolder(holder); exists {
			fmt.Printf("%s: %d tokens (Reputation: %d)\n",
				names[i], balance, tokenHolder.Reputation)
		} else {
			fmt.Printf("%s: %d tokens\n", names[i], balance)
		}
	}

	fmt.Printf("\nFinal Total Supply: %d tokens\n", dao.GetTotalSupply())
	fmt.Println("\n=== Token System Example Complete ===")
}
