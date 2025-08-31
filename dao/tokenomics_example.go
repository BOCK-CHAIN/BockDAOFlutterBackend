package dao

import (
	"fmt"
	"log"
	"time"

	"github.com/BOCK-CHAIN/BockChain/crypto"
)

// TokenomicsExample demonstrates the complete tokenomics system functionality
func TokenomicsExample() {
	fmt.Println("=== ProjectX DAO Tokenomics System Example ===")

	// Create a new DAO
	dao := NewDAO("PXGOV", "ProjectX Governance Token", 18)
	tm := NewTokenomicsManager(dao.GovernanceState, dao.TokenState)

	fmt.Printf("Created DAO with token: %s (%s)\n", dao.TokenState.Name, dao.TokenState.Symbol)

	// Initialize tokenomics
	err := tm.InitializeTokenDistribution()
	if err != nil {
		log.Fatalf("Failed to initialize tokenomics: %v", err)
	}

	fmt.Printf("Total Supply: %d tokens\n", tm.GetTokenomicsConfig().TotalSupply)
	fmt.Println("\n--- Token Distribution Allocations ---")

	// Display allocation breakdown
	config := tm.GetTokenomicsConfig()
	fmt.Printf("Founder Allocation: %d%% (%d tokens)\n",
		config.FounderAllocation/100, (config.TotalSupply*config.FounderAllocation)/10000)
	fmt.Printf("Team Allocation: %d%% (%d tokens)\n",
		config.TeamAllocation/100, (config.TotalSupply*config.TeamAllocation)/10000)
	fmt.Printf("Community Allocation: %d%% (%d tokens)\n",
		config.CommunityAllocation/100, (config.TotalSupply*config.CommunityAllocation)/10000)
	fmt.Printf("Treasury Allocation: %d%% (%d tokens)\n",
		config.TreasuryAllocation/100, (config.TotalSupply*config.TreasuryAllocation)/10000)
	fmt.Printf("Ecosystem Allocation: %d%% (%d tokens)\n",
		config.EcosystemAllocation/100, (config.TotalSupply*config.EcosystemAllocation)/10000)

	// Create test participants
	founder1 := crypto.GeneratePrivateKey().PublicKey()
	founder2 := crypto.GeneratePrivateKey().PublicKey()
	teamLead := crypto.GeneratePrivateKey().PublicKey()
	developer := crypto.GeneratePrivateKey().PublicKey()
	communityMember := crypto.GeneratePrivateKey().PublicKey()

	fmt.Println("\n--- Founder Token Distribution with Vesting ---")

	// Distribute founder tokens with vesting
	err = tm.AddDistributionRecipient(DistributionFounders, founder1, 100000000) // 100M tokens
	if err != nil {
		log.Fatalf("Failed to add founder1: %v", err)
	}

	err = tm.AddDistributionRecipient(DistributionFounders, founder2, 100000000) // 100M tokens
	if err != nil {
		log.Fatalf("Failed to add founder2: %v", err)
	}

	fmt.Printf("Founder1 allocated: 100,000,000 tokens (with vesting)\n")
	fmt.Printf("Founder2 allocated: 100,000,000 tokens (with vesting)\n")

	// Check vesting schedules
	founder1Schedules := tm.GetVestingSchedulesByBeneficiary(founder1)
	if len(founder1Schedules) > 0 {
		schedule := founder1Schedules[0]
		fmt.Printf("Founder1 vesting: %d tokens over %d seconds with %d second cliff\n",
			schedule.TotalAmount, schedule.Duration, schedule.CliffTime-schedule.StartTime)
	}

	fmt.Println("\n--- Team Token Distribution with Shorter Vesting ---")

	// Distribute team tokens with shorter vesting
	err = tm.AddDistributionRecipient(DistributionTeam, teamLead, 75000000) // 75M tokens
	if err != nil {
		log.Fatalf("Failed to add team lead: %v", err)
	}

	err = tm.AddDistributionRecipient(DistributionTeam, developer, 75000000) // 75M tokens
	if err != nil {
		log.Fatalf("Failed to add developer: %v", err)
	}

	fmt.Printf("Team Lead allocated: 75,000,000 tokens (with shorter vesting)\n")
	fmt.Printf("Developer allocated: 75,000,000 tokens (with shorter vesting)\n")

	fmt.Println("\n--- Community Token Distribution (Immediate) ---")

	// Distribute community tokens immediately
	err = tm.AddDistributionRecipient(DistributionCommunity, communityMember, 150000000) // 150M tokens
	if err != nil {
		log.Fatalf("Failed to add community member: %v", err)
	}

	fmt.Printf("Community Member allocated: 150,000,000 tokens (immediate)\n")
	fmt.Printf("Community Member balance: %d tokens\n", dao.GetTokenBalance(communityMember))

	fmt.Println("\n--- Staking Pool Creation ---")

	// Create staking pools
	mainPoolID := "main-staking"
	highRewardPoolID := "high-reward-staking"
	lockedPoolID := "locked-staking"

	err = tm.CreateStakingPool(mainPoolID, "Main Staking Pool", 100, 1000, 0)
	if err != nil {
		log.Fatalf("Failed to create main staking pool: %v", err)
	}

	err = tm.CreateStakingPool(highRewardPoolID, "High Reward Pool", 500, 10000, 0)
	if err != nil {
		log.Fatalf("Failed to create high reward pool: %v", err)
	}

	err = tm.CreateStakingPool(lockedPoolID, "Locked Staking Pool", 300, 5000, 86400) // 1 day lockup
	if err != nil {
		log.Fatalf("Failed to create locked pool: %v", err)
	}

	fmt.Printf("Created 3 staking pools:\n")
	fmt.Printf("- Main Pool: 100 reward rate, 1000 min stake\n")
	fmt.Printf("- High Reward Pool: 500 reward rate, 10000 min stake\n")
	fmt.Printf("- Locked Pool: 300 reward rate, 5000 min stake, 1 day lockup\n")

	fmt.Println("\n--- Community Member Staking ---")

	// Community member stakes tokens
	stakeAmount := uint64(50000000) // 50M tokens
	err = tm.StakeTokens(mainPoolID, communityMember, stakeAmount, 0)
	if err != nil {
		log.Fatalf("Failed to stake tokens: %v", err)
	}

	fmt.Printf("Community member staked: %d tokens in main pool\n", stakeAmount)
	fmt.Printf("Remaining balance: %d tokens\n", dao.GetTokenBalance(communityMember))

	// Check staking info
	stakerInfo, exists := tm.GetStakerInfo(mainPoolID, communityMember)
	if exists {
		fmt.Printf("Staked amount: %d tokens\n", stakerInfo.StakedAmount)
		fmt.Printf("Stake time: %s\n", time.Unix(stakerInfo.StakeTime, 0).Format("2006-01-02 15:04:05"))
	}

	fmt.Println("\n--- Simulating Time Passage for Vesting ---")

	// Simulate time passing for founder vesting
	founder1Schedule := founder1Schedules[0]
	originalCliffTime := founder1Schedule.CliffTime
	originalStartTime := founder1Schedule.StartTime

	// Simulate 6 months passing (cliff period)
	founder1Schedule.CliffTime = time.Now().Unix() - 1000
	founder1Schedule.StartTime = time.Now().Unix() - 2000

	fmt.Printf("Simulated time passage: cliff period completed\n")

	// Founder1 claims vested tokens
	claimedAmount, err := tm.ClaimVestedTokens(founder1Schedule.ID, founder1)
	if err != nil {
		log.Fatalf("Failed to claim vested tokens: %v", err)
	}

	fmt.Printf("Founder1 claimed: %d vested tokens\n", claimedAmount)
	fmt.Printf("Founder1 balance: %d tokens\n", dao.GetTokenBalance(founder1))

	// Restore original times for demonstration
	founder1Schedule.CliffTime = originalCliffTime
	founder1Schedule.StartTime = originalStartTime

	fmt.Println("\n--- Founder Staking Vested Tokens ---")

	// Founder1 stakes some of their vested tokens (if enough)
	founderStakeAmount := claimedAmount / 2 // Stake half
	if founderStakeAmount >= 10000 {        // Check if meets high reward pool minimum
		err = tm.StakeTokens(highRewardPoolID, founder1, founderStakeAmount, 0)
		if err != nil {
			log.Fatalf("Failed to stake founder tokens: %v", err)
		}
		fmt.Printf("Founder1 staked: %d tokens in high reward pool\n", founderStakeAmount)
	} else if founderStakeAmount >= 1000 { // Try main pool instead
		err = tm.StakeTokens(mainPoolID, founder1, founderStakeAmount, 0)
		if err != nil {
			log.Fatalf("Failed to stake founder tokens: %v", err)
		}
		fmt.Printf("Founder1 staked: %d tokens in main pool\n", founderStakeAmount)
	} else {
		fmt.Printf("Founder1 claimed amount (%d) too small to stake, keeping as liquid balance\n", claimedAmount)
	}
	fmt.Printf("Founder1 remaining balance: %d tokens\n", dao.GetTokenBalance(founder1))

	fmt.Println("\n--- Simulating Staking Rewards ---")

	// Simulate time passing for staking rewards
	mainPool, _ := tm.GetStakingPool(mainPoolID)
	highRewardPool, _ := tm.GetStakingPool(highRewardPoolID)

	// Set last update time to simulate rewards accumulation
	mainPool.LastUpdateTime = time.Now().Unix() - 3600       // 1 hour ago
	highRewardPool.LastUpdateTime = time.Now().Unix() - 3600 // 1 hour ago

	fmt.Printf("Simulated 1 hour of staking rewards accumulation\n")

	// Community member claims rewards from main pool
	communityRewards, err := tm.ClaimStakingRewards(mainPoolID, communityMember)
	if err != nil {
		log.Fatalf("Failed to claim community rewards: %v", err)
	}

	fmt.Printf("Community member claimed rewards: %d tokens\n", communityRewards)
	fmt.Printf("Community member new balance: %d tokens\n", dao.GetTokenBalance(communityMember))

	// Founder1 claims rewards from whichever pool they're in
	founderTotalStaked := tm.GetTotalStakedByUser(founder1)
	if founderTotalStaked > 0 {
		// Try to claim from main pool first
		founderRewards, err := tm.ClaimStakingRewards(mainPoolID, founder1)
		if err == nil && founderRewards > 0 {
			fmt.Printf("Founder1 claimed rewards from main pool: %d tokens\n", founderRewards)
		} else {
			// Try high reward pool
			founderRewards, err = tm.ClaimStakingRewards(highRewardPoolID, founder1)
			if err == nil && founderRewards > 0 {
				fmt.Printf("Founder1 claimed rewards from high reward pool: %d tokens\n", founderRewards)
			} else {
				fmt.Printf("Founder1 has no rewards to claim at this time\n")
			}
		}
	} else {
		fmt.Printf("Founder1 has no staked tokens to generate rewards\n")
	}
	fmt.Printf("Founder1 new balance: %d tokens\n", dao.GetTokenBalance(founder1))

	fmt.Println("\n--- Partial Unstaking ---")

	// Community member unstakes partial amount
	unstakeAmount := uint64(20000000) // 20M tokens
	err = tm.UnstakeTokens(mainPoolID, communityMember, unstakeAmount)
	if err != nil {
		log.Fatalf("Failed to unstake tokens: %v", err)
	}

	fmt.Printf("Community member unstaked: %d tokens\n", unstakeAmount)
	fmt.Printf("Community member balance after unstaking: %d tokens\n", dao.GetTokenBalance(communityMember))

	// Check remaining staked amount
	updatedStakerInfo, _ := tm.GetStakerInfo(mainPoolID, communityMember)
	fmt.Printf("Community member remaining staked: %d tokens\n", updatedStakerInfo.StakedAmount)

	fmt.Println("\n--- Multi-Pool Staking Summary ---")

	// Show total staked across all pools for each user
	communityTotalStaked := tm.GetTotalStakedByUser(communityMember)
	founderTotalStaked2 := tm.GetTotalStakedByUser(founder1)

	fmt.Printf("Community member total staked across all pools: %d tokens\n", communityTotalStaked)
	fmt.Printf("Founder1 total staked across all pools: %d tokens\n", founderTotalStaked2)

	// Show total pending rewards
	communityTotalRewards := tm.GetTotalRewardsByUser(communityMember)
	founderTotalRewards := tm.GetTotalRewardsByUser(founder1)

	fmt.Printf("Community member total pending rewards: %d tokens\n", communityTotalRewards)
	fmt.Printf("Founder1 total pending rewards: %d tokens\n", founderTotalRewards)

	fmt.Println("\n--- Staking Pool Statistics ---")

	// Display pool statistics
	pools := []string{mainPoolID, highRewardPoolID, lockedPoolID}
	poolNames := []string{"Main Pool", "High Reward Pool", "Locked Pool"}

	for i, poolID := range pools {
		pool, exists := tm.GetStakingPool(poolID)
		if exists {
			fmt.Printf("%s:\n", poolNames[i])
			fmt.Printf("  Total Staked: %d tokens\n", pool.TotalStaked)
			fmt.Printf("  Active Stakers: %d\n", len(pool.Stakers))
			fmt.Printf("  Reward Rate: %d tokens/second/token\n", pool.RewardRate)
			fmt.Printf("  Min Stake: %d tokens\n", pool.MinStakeAmount)
			if pool.LockupPeriod > 0 {
				fmt.Printf("  Lockup Period: %d seconds\n", pool.LockupPeriod)
			}
		}
	}

	fmt.Println("\n--- Distribution Summary ---")

	// Show distribution status
	distributions := tm.ListAllDistributions()
	for categoryName, dist := range distributions {
		fmt.Printf("%s Distribution:\n", categoryName)
		fmt.Printf("  Allocated: %d tokens\n", dist.Allocation)
		fmt.Printf("  Distributed: %d tokens\n", dist.Distributed)
		fmt.Printf("  Remaining: %d tokens\n", dist.Allocation-dist.Distributed)
		fmt.Printf("  Recipients: %d\n", len(dist.Recipients))
		fmt.Printf("  Vesting Type: %d\n", dist.VestingType)
	}

	fmt.Println("\n--- Vesting Schedule Summary ---")

	// Show all vesting schedules
	allSchedules := tm.ListAllVestingSchedules()
	fmt.Printf("Total Vesting Schedules: %d\n", len(allSchedules))

	for vestingID, schedule := range allSchedules {
		fmt.Printf("Schedule %s:\n", vestingID[:8]+"...")
		fmt.Printf("  Beneficiary: %s...\n", schedule.Beneficiary.String()[:8])
		fmt.Printf("  Total Amount: %d tokens\n", schedule.TotalAmount)
		fmt.Printf("  Released: %d tokens\n", schedule.Released)
		fmt.Printf("  Remaining: %d tokens\n", schedule.TotalAmount-schedule.Released)
		fmt.Printf("  Vesting Type: %d\n", schedule.VestingType)
	}

	fmt.Println("\n--- Final Token Holder Summary ---")

	// Display all token holders
	holders := []crypto.PublicKey{founder1, founder2, teamLead, developer, communityMember}
	names := []string{"Founder1", "Founder2", "Team Lead", "Developer", "Community Member"}

	for i, holder := range holders {
		balance := dao.GetTokenBalance(holder)
		totalStaked := tm.GetTotalStakedByUser(holder)
		totalRewards := tm.GetTotalRewardsByUser(holder)

		if tokenHolder, exists := dao.GetTokenHolder(holder); exists {
			fmt.Printf("%s:\n", names[i])
			fmt.Printf("  Liquid Balance: %d tokens\n", balance)
			fmt.Printf("  Total Staked: %d tokens\n", totalStaked)
			fmt.Printf("  Pending Rewards: %d tokens\n", totalRewards)
			fmt.Printf("  Total Value: %d tokens\n", balance+totalStaked+totalRewards)
			fmt.Printf("  Reputation: %d\n", tokenHolder.Reputation)
		} else if balance > 0 || totalStaked > 0 {
			fmt.Printf("%s:\n", names[i])
			fmt.Printf("  Liquid Balance: %d tokens\n", balance)
			fmt.Printf("  Total Staked: %d tokens\n", totalStaked)
			fmt.Printf("  Pending Rewards: %d tokens\n", totalRewards)
			fmt.Printf("  Total Value: %d tokens\n", balance+totalStaked+totalRewards)
		}
	}

	fmt.Printf("\nTotal Supply: %d tokens\n", dao.GetTotalSupply())
	fmt.Println("\n=== Tokenomics System Example Complete ===")
}

// TokenomicsConfigExample demonstrates tokenomics configuration management
func TokenomicsConfigExample() {
	fmt.Println("=== Tokenomics Configuration Management Example ===")

	// Create DAO
	dao := NewDAO("PXGOV", "ProjectX Governance Token", 18)
	tm := NewTokenomicsManager(dao.GovernanceState, dao.TokenState)

	fmt.Println("\n--- Default Configuration ---")
	defaultConfig := tm.GetTokenomicsConfig()
	displayTokenomicsConfig("Default", defaultConfig)

	fmt.Println("\n--- Custom Configuration ---")
	// Create custom configuration
	customConfig := &TokenomicsConfig{
		TotalSupply:            2000000000,          // 2B tokens
		FounderAllocation:      1500,                // 15%
		TeamAllocation:         2500,                // 25%
		CommunityAllocation:    3500,                // 35%
		TreasuryAllocation:     1500,                // 15%
		EcosystemAllocation:    1000,                // 10%
		DefaultVestingCliff:    180 * 24 * 3600,     // 6 months
		DefaultVestingDuration: 3 * 365 * 24 * 3600, // 3 years
		StakingRewardRate:      300,                 // Higher rewards
		MinStakeAmount:         5000,                // Higher minimum
	}

	err := tm.UpdateTokenomicsConfig(customConfig)
	if err != nil {
		log.Fatalf("Failed to update config: %v", err)
	}

	displayTokenomicsConfig("Custom", customConfig)

	fmt.Println("\n--- Initialize with Custom Configuration ---")
	err = tm.InitializeTokenDistribution()
	if err != nil {
		log.Fatalf("Failed to initialize with custom config: %v", err)
	}

	// Show updated allocations
	distributions := tm.ListAllDistributions()
	for categoryName, dist := range distributions {
		percentage := float64(dist.Allocation) / float64(customConfig.TotalSupply) * 100
		fmt.Printf("%s: %d tokens (%.1f%%)\n", categoryName, dist.Allocation, percentage)
	}

	fmt.Println("\n=== Configuration Management Example Complete ===")
}

func displayTokenomicsConfig(name string, config *TokenomicsConfig) {
	fmt.Printf("%s Configuration:\n", name)
	fmt.Printf("  Total Supply: %d tokens\n", config.TotalSupply)
	fmt.Printf("  Founder Allocation: %.1f%%\n", float64(config.FounderAllocation)/100)
	fmt.Printf("  Team Allocation: %.1f%%\n", float64(config.TeamAllocation)/100)
	fmt.Printf("  Community Allocation: %.1f%%\n", float64(config.CommunityAllocation)/100)
	fmt.Printf("  Treasury Allocation: %.1f%%\n", float64(config.TreasuryAllocation)/100)
	fmt.Printf("  Ecosystem Allocation: %.1f%%\n", float64(config.EcosystemAllocation)/100)
	fmt.Printf("  Default Vesting Cliff: %d days\n", config.DefaultVestingCliff/(24*3600))
	fmt.Printf("  Default Vesting Duration: %d days\n", config.DefaultVestingDuration/(24*3600))
	fmt.Printf("  Staking Reward Rate: %d tokens/second/token\n", config.StakingRewardRate)
	fmt.Printf("  Min Stake Amount: %d tokens\n", config.MinStakeAmount)
}
