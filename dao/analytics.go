package dao

import (
	"math"
	"sort"
	"time"
)

// AnalyticsSystem provides comprehensive analytics and reporting for DAO operations
type AnalyticsSystem struct {
	governanceState *GovernanceState
	tokenState      *GovernanceToken
}

// NewAnalyticsSystem creates a new analytics system instance
func NewAnalyticsSystem(governanceState *GovernanceState, tokenState *GovernanceToken) *AnalyticsSystem {
	return &AnalyticsSystem{
		governanceState: governanceState,
		tokenState:      tokenState,
	}
}

// GovernanceParticipationMetrics tracks participation in governance activities
type GovernanceParticipationMetrics struct {
	TotalProposals      uint64                   `json:"total_proposals"`
	ActiveProposals     uint64                   `json:"active_proposals"`
	TotalVotes          uint64                   `json:"total_votes"`
	UniqueVoters        uint64                   `json:"unique_voters"`
	AverageVotesPerUser float64                  `json:"average_votes_per_user"`
	ParticipationRate   float64                  `json:"participation_rate"`
	TopParticipants     []ParticipantStats       `json:"top_participants"`
	VotingPatterns      map[VoteChoice]uint64    `json:"voting_patterns"`
	ProposalsByType     map[ProposalType]uint64  `json:"proposals_by_type"`
	VotingByType        map[VotingType]uint64    `json:"voting_by_type"`
	TimeSeriesData      []ParticipationTimePoint `json:"time_series_data"`
	DelegationMetrics   DelegationAnalytics      `json:"delegation_metrics"`
}

// ParticipantStats represents statistics for individual participants
type ParticipantStats struct {
	Address           string  `json:"address"`
	ProposalsCreated  uint64  `json:"proposals_created"`
	VotesCast         uint64  `json:"votes_cast"`
	ParticipationRate float64 `json:"participation_rate"`
	AverageVoteWeight uint64  `json:"average_vote_weight"`
	LastActivity      int64   `json:"last_activity"`
	Reputation        uint64  `json:"reputation"`
}

// ParticipationTimePoint represents participation data at a specific time
type ParticipationTimePoint struct {
	Timestamp          int64   `json:"timestamp"`
	ProposalsCreated   uint64  `json:"proposals_created"`
	VotesCast          uint64  `json:"votes_cast"`
	UniqueParticipants uint64  `json:"unique_participants"`
	ParticipationRate  float64 `json:"participation_rate"`
}

// DelegationAnalytics provides insights into delegation patterns
type DelegationAnalytics struct {
	TotalDelegations       uint64            `json:"total_delegations"`
	ActiveDelegations      uint64            `json:"active_delegations"`
	DelegationRate         float64           `json:"delegation_rate"`
	TopDelegates           []DelegateStats   `json:"top_delegates"`
	AverageDelegationTime  float64           `json:"average_delegation_time"`
	DelegationDistribution map[string]uint64 `json:"delegation_distribution"`
}

// DelegateStats represents statistics for delegates
type DelegateStats struct {
	Address          string  `json:"address"`
	DelegatorsCount  uint64  `json:"delegators_count"`
	TotalVotingPower uint64  `json:"total_voting_power"`
	VotesCast        uint64  `json:"votes_cast"`
	EfficiencyRate   float64 `json:"efficiency_rate"`
}

// TreasuryPerformanceMetrics tracks treasury operations and performance
type TreasuryPerformanceMetrics struct {
	CurrentBalance         uint64              `json:"current_balance"`
	TotalInflows           uint64              `json:"total_inflows"`
	TotalOutflows          uint64              `json:"total_outflows"`
	NetFlow                int64               `json:"net_flow"`
	TransactionCount       uint64              `json:"transaction_count"`
	AverageTransactionSize uint64              `json:"average_transaction_size"`
	LargestTransaction     uint64              `json:"largest_transaction"`
	SmallestTransaction    uint64              `json:"smallest_transaction"`
	TransactionsByPurpose  map[string]uint64   `json:"transactions_by_purpose"`
	MonthlyFlows           []TreasuryFlowPoint `json:"monthly_flows"`
	SigningEfficiency      float64             `json:"signing_efficiency"`
	AverageSigningTime     float64             `json:"average_signing_time"`
	PendingTransactions    uint64              `json:"pending_transactions"`
	ExecutedTransactions   uint64              `json:"executed_transactions"`
	ExpiredTransactions    uint64              `json:"expired_transactions"`
}

// TreasuryFlowPoint represents treasury flow data at a specific time
type TreasuryFlowPoint struct {
	Timestamp int64  `json:"timestamp"`
	Inflows   uint64 `json:"inflows"`
	Outflows  uint64 `json:"outflows"`
	Balance   uint64 `json:"balance"`
}

// ProposalAnalytics provides insights into proposal patterns and success rates
type ProposalAnalytics struct {
	TotalProposals        uint64                         `json:"total_proposals"`
	PassedProposals       uint64                         `json:"passed_proposals"`
	RejectedProposals     uint64                         `json:"rejected_proposals"`
	PendingProposals      uint64                         `json:"pending_proposals"`
	SuccessRate           float64                        `json:"success_rate"`
	AverageVotingPeriod   float64                        `json:"average_voting_period"`
	QuorumAchievementRate float64                        `json:"quorum_achievement_rate"`
	ProposalsByCreator    map[string]uint64              `json:"proposals_by_creator"`
	SuccessRateByType     map[ProposalType]float64       `json:"success_rate_by_type"`
	VotingPatternsByType  map[ProposalType]VotingPattern `json:"voting_patterns_by_type"`
	TimeToResolution      []ResolutionTimePoint          `json:"time_to_resolution"`
	PopularityMetrics     ProposalPopularityMetrics      `json:"popularity_metrics"`
}

// VotingPattern represents voting behavior for a specific category
type VotingPattern struct {
	YesVotes     uint64  `json:"yes_votes"`
	NoVotes      uint64  `json:"no_votes"`
	AbstainVotes uint64  `json:"abstain_votes"`
	YesRate      float64 `json:"yes_rate"`
	NoRate       float64 `json:"no_rate"`
	AbstainRate  float64 `json:"abstain_rate"`
}

// ResolutionTimePoint represents time to resolution data
type ResolutionTimePoint struct {
	ProposalType      ProposalType `json:"proposal_type"`
	ResolutionTime    float64      `json:"resolution_time_hours"`
	VotingPeriod      float64      `json:"voting_period_hours"`
	ParticipationRate float64      `json:"participation_rate"`
}

// ProposalPopularityMetrics tracks proposal engagement
type ProposalPopularityMetrics struct {
	MostVotedProposal    string  `json:"most_voted_proposal"`
	HighestParticipation float64 `json:"highest_participation"`
	AverageParticipation float64 `json:"average_participation"`
	EngagementTrend      string  `json:"engagement_trend"`
}

// DAOHealthMetrics provides overall health indicators for the DAO
type DAOHealthMetrics struct {
	OverallScore        float64         `json:"overall_score"`
	ParticipationHealth float64         `json:"participation_health"`
	TreasuryHealth      float64         `json:"treasury_health"`
	GovernanceHealth    float64         `json:"governance_health"`
	SecurityHealth      float64         `json:"security_health"`
	GrowthMetrics       GrowthMetrics   `json:"growth_metrics"`
	RiskIndicators      []RiskIndicator `json:"risk_indicators"`
	Recommendations     []string        `json:"recommendations"`
	HealthTrend         string          `json:"health_trend"`
	LastUpdated         int64           `json:"last_updated"`
}

// GrowthMetrics tracks DAO growth over time
type GrowthMetrics struct {
	MemberGrowthRate      float64 `json:"member_growth_rate"`
	TokenHolderGrowthRate float64 `json:"token_holder_growth_rate"`
	ProposalGrowthRate    float64 `json:"proposal_growth_rate"`
	TreasuryGrowthRate    float64 `json:"treasury_growth_rate"`
	ActivityGrowthRate    float64 `json:"activity_growth_rate"`
}

// RiskIndicator represents potential risks to DAO health
type RiskIndicator struct {
	Type        string  `json:"type"`
	Severity    string  `json:"severity"`
	Description string  `json:"description"`
	Score       float64 `json:"score"`
	Mitigation  string  `json:"mitigation"`
}

// GetGovernanceParticipationMetrics calculates comprehensive participation metrics
func (as *AnalyticsSystem) GetGovernanceParticipationMetrics() *GovernanceParticipationMetrics {
	metrics := &GovernanceParticipationMetrics{
		VotingPatterns:  make(map[VoteChoice]uint64),
		ProposalsByType: make(map[ProposalType]uint64),
		VotingByType:    make(map[VotingType]uint64),
		TimeSeriesData:  make([]ParticipationTimePoint, 0),
	}

	// Count proposals and analyze by type
	for _, proposal := range as.governanceState.Proposals {
		metrics.TotalProposals++
		metrics.ProposalsByType[proposal.ProposalType]++
		metrics.VotingByType[proposal.VotingType]++

		if proposal.Status == ProposalStatusActive {
			metrics.ActiveProposals++
		}
	}

	// Analyze votes and voting patterns
	uniqueVoters := make(map[string]bool)
	participantStats := make(map[string]*ParticipantStats)

	for proposalID, votes := range as.governanceState.Votes {
		for voterStr, vote := range votes {
			metrics.TotalVotes++
			uniqueVoters[voterStr] = true
			metrics.VotingPatterns[vote.Choice]++

			// Track participant statistics
			if _, exists := participantStats[voterStr]; !exists {
				participantStats[voterStr] = &ParticipantStats{
					Address: voterStr,
				}
			}
			participantStats[voterStr].VotesCast++
			participantStats[voterStr].LastActivity = vote.Timestamp

			// Check if this voter created the proposal
			if proposal, exists := as.governanceState.Proposals[proposalID]; exists {
				if proposal.Creator.String() == voterStr {
					participantStats[voterStr].ProposalsCreated++
				}
			}
		}
	}

	metrics.UniqueVoters = uint64(len(uniqueVoters))

	// Calculate participation rate
	totalTokenHolders := uint64(len(as.governanceState.TokenHolders))
	if totalTokenHolders > 0 {
		metrics.ParticipationRate = float64(metrics.UniqueVoters) / float64(totalTokenHolders) * 100
	}

	// Calculate average votes per user
	if metrics.UniqueVoters > 0 {
		metrics.AverageVotesPerUser = float64(metrics.TotalVotes) / float64(metrics.UniqueVoters)
	}

	// Build top participants list
	participants := make([]ParticipantStats, 0, len(participantStats))
	for _, stats := range participantStats {
		// Calculate participation rate for individual
		if metrics.TotalProposals > 0 {
			stats.ParticipationRate = float64(stats.VotesCast) / float64(metrics.TotalProposals) * 100
		}

		// Get reputation from token holder data
		if holder, exists := as.governanceState.TokenHolders[stats.Address]; exists {
			stats.Reputation = holder.Reputation
		}

		participants = append(participants, *stats)
	}

	// Sort by votes cast and take top 10
	sort.Slice(participants, func(i, j int) bool {
		return participants[i].VotesCast > participants[j].VotesCast
	})

	if len(participants) > 10 {
		participants = participants[:10]
	}
	metrics.TopParticipants = participants

	// Get delegation metrics
	metrics.DelegationMetrics = as.getDelegationAnalytics()

	return metrics
}

// getDelegationAnalytics calculates delegation-specific metrics
func (as *AnalyticsSystem) getDelegationAnalytics() DelegationAnalytics {
	analytics := DelegationAnalytics{
		TopDelegates:           make([]DelegateStats, 0),
		DelegationDistribution: make(map[string]uint64),
	}

	now := time.Now().Unix()
	delegateStats := make(map[string]*DelegateStats)

	// Analyze delegations
	for _, delegation := range as.governanceState.Delegations {
		analytics.TotalDelegations++

		if delegation.Active && now >= delegation.StartTime && now <= delegation.EndTime {
			analytics.ActiveDelegations++

			delegateStr := delegation.Delegate.String()
			if _, exists := delegateStats[delegateStr]; !exists {
				delegateStats[delegateStr] = &DelegateStats{
					Address: delegateStr,
				}
			}
			delegateStats[delegateStr].DelegatorsCount++

			// Track delegation distribution
			analytics.DelegationDistribution[delegateStr]++
		}
	}

	// Calculate delegation rate
	totalTokenHolders := uint64(len(as.governanceState.TokenHolders))
	if totalTokenHolders > 0 {
		analytics.DelegationRate = float64(analytics.ActiveDelegations) / float64(totalTokenHolders) * 100
	}

	// Build top delegates list
	delegates := make([]DelegateStats, 0, len(delegateStats))
	for _, stats := range delegateStats {
		delegates = append(delegates, *stats)
	}

	// Sort by delegators count
	sort.Slice(delegates, func(i, j int) bool {
		return delegates[i].DelegatorsCount > delegates[j].DelegatorsCount
	})

	if len(delegates) > 10 {
		delegates = delegates[:10]
	}
	analytics.TopDelegates = delegates

	return analytics
}

// GetTreasuryPerformanceMetrics calculates treasury performance metrics
func (as *AnalyticsSystem) GetTreasuryPerformanceMetrics() *TreasuryPerformanceMetrics {
	metrics := &TreasuryPerformanceMetrics{
		CurrentBalance:        as.governanceState.Treasury.Balance,
		TransactionsByPurpose: make(map[string]uint64),
		MonthlyFlows:          make([]TreasuryFlowPoint, 0),
	}

	var totalTransactionSize uint64
	var signingTimes []float64

	// Analyze treasury transactions
	for _, tx := range as.governanceState.Treasury.Transactions {
		metrics.TransactionCount++

		if tx.Executed {
			metrics.ExecutedTransactions++
			metrics.TotalOutflows += tx.Amount
			totalTransactionSize += tx.Amount

			// Track by purpose
			if tx.Purpose != "" {
				metrics.TransactionsByPurpose[tx.Purpose]++
			} else {
				metrics.TransactionsByPurpose["Unspecified"]++
			}

			// Track transaction sizes
			if metrics.LargestTransaction == 0 || tx.Amount > metrics.LargestTransaction {
				metrics.LargestTransaction = tx.Amount
			}
			if metrics.SmallestTransaction == 0 || tx.Amount < metrics.SmallestTransaction {
				metrics.SmallestTransaction = tx.Amount
			}

			// Calculate signing time if we have creation and execution timestamps
			if tx.CreatedAt > 0 {
				signingTime := float64(time.Now().Unix()-tx.CreatedAt) / 3600 // Convert to hours
				signingTimes = append(signingTimes, signingTime)
			}
		} else if time.Now().Unix() > tx.ExpiresAt {
			metrics.ExpiredTransactions++
		} else {
			metrics.PendingTransactions++
		}
	}

	// Calculate averages
	if metrics.ExecutedTransactions > 0 {
		metrics.AverageTransactionSize = totalTransactionSize / metrics.ExecutedTransactions
	}

	if len(signingTimes) > 0 {
		var totalSigningTime float64
		for _, time := range signingTimes {
			totalSigningTime += time
		}
		metrics.AverageSigningTime = totalSigningTime / float64(len(signingTimes))
	}

	// Calculate signing efficiency
	if metrics.TransactionCount > 0 {
		metrics.SigningEfficiency = float64(metrics.ExecutedTransactions) / float64(metrics.TransactionCount) * 100
	}

	// Calculate net flow
	metrics.NetFlow = int64(metrics.TotalInflows) - int64(metrics.TotalOutflows)

	return metrics
}

// GetProposalAnalytics calculates proposal success rates and patterns
func (as *AnalyticsSystem) GetProposalAnalytics() *ProposalAnalytics {
	analytics := &ProposalAnalytics{
		ProposalsByCreator:   make(map[string]uint64),
		SuccessRateByType:    make(map[ProposalType]float64),
		VotingPatternsByType: make(map[ProposalType]VotingPattern),
		TimeToResolution:     make([]ResolutionTimePoint, 0),
	}

	typeStats := make(map[ProposalType]struct {
		total  uint64
		passed uint64
		votes  VotingPattern
	})

	var totalVotingPeriod float64
	var quorumAchieved uint64
	var totalParticipation float64
	var participationCount uint64

	// Analyze proposals
	for _, proposal := range as.governanceState.Proposals {
		analytics.TotalProposals++
		analytics.ProposalsByCreator[proposal.Creator.String()]++

		// Initialize type stats if needed
		if _, exists := typeStats[proposal.ProposalType]; !exists {
			typeStats[proposal.ProposalType] = struct {
				total  uint64
				passed uint64
				votes  VotingPattern
			}{}
		}

		stats := typeStats[proposal.ProposalType]
		stats.total++

		switch proposal.Status {
		case ProposalStatusPassed:
			analytics.PassedProposals++
			stats.passed++
		case ProposalStatusRejected:
			analytics.RejectedProposals++
		case ProposalStatusPending, ProposalStatusActive:
			analytics.PendingProposals++
		}

		// Calculate voting period
		if proposal.EndTime > proposal.StartTime {
			votingPeriod := float64(proposal.EndTime-proposal.StartTime) / 3600 // Convert to hours
			totalVotingPeriod += votingPeriod
		}

		// Analyze voting patterns for this proposal
		if votes, exists := as.governanceState.Votes[proposal.ID]; exists {
			var yesVotes, noVotes, abstainVotes uint64
			totalVotes := uint64(len(votes))

			for _, vote := range votes {
				switch vote.Choice {
				case VoteChoiceYes:
					yesVotes++
					stats.votes.YesVotes++
				case VoteChoiceNo:
					noVotes++
					stats.votes.NoVotes++
				case VoteChoiceAbstain:
					abstainVotes++
					stats.votes.AbstainVotes++
				}
			}

			// Check if quorum was achieved
			if proposal.Results != nil && proposal.Results.Quorum > 0 {
				if totalVotes >= proposal.Results.Quorum {
					quorumAchieved++
				}
			}

			// Calculate participation rate for this proposal
			totalTokenHolders := uint64(len(as.governanceState.TokenHolders))
			if totalTokenHolders > 0 {
				participationRate := float64(totalVotes) / float64(totalTokenHolders) * 100
				totalParticipation += participationRate
				participationCount++

				// Add to time to resolution data
				if proposal.Status == ProposalStatusPassed || proposal.Status == ProposalStatusRejected {
					resolutionTime := float64(proposal.EndTime-proposal.StartTime) / 3600
					analytics.TimeToResolution = append(analytics.TimeToResolution, ResolutionTimePoint{
						ProposalType:      proposal.ProposalType,
						ResolutionTime:    resolutionTime,
						VotingPeriod:      resolutionTime,
						ParticipationRate: participationRate,
					})
				}
			}
		}

		typeStats[proposal.ProposalType] = stats
	}

	// Calculate success rate
	if analytics.TotalProposals > 0 {
		analytics.SuccessRate = float64(analytics.PassedProposals) / float64(analytics.TotalProposals) * 100
	}

	// Calculate average voting period
	if analytics.TotalProposals > 0 {
		analytics.AverageVotingPeriod = totalVotingPeriod / float64(analytics.TotalProposals)
	}

	// Calculate quorum achievement rate
	if analytics.TotalProposals > 0 {
		analytics.QuorumAchievementRate = float64(quorumAchieved) / float64(analytics.TotalProposals) * 100
	}

	// Calculate success rates and voting patterns by type
	for proposalType, stats := range typeStats {
		if stats.total > 0 {
			analytics.SuccessRateByType[proposalType] = float64(stats.passed) / float64(stats.total) * 100

			// Calculate voting pattern percentages
			totalVotes := stats.votes.YesVotes + stats.votes.NoVotes + stats.votes.AbstainVotes
			if totalVotes > 0 {
				pattern := VotingPattern{
					YesVotes:     stats.votes.YesVotes,
					NoVotes:      stats.votes.NoVotes,
					AbstainVotes: stats.votes.AbstainVotes,
					YesRate:      float64(stats.votes.YesVotes) / float64(totalVotes) * 100,
					NoRate:       float64(stats.votes.NoVotes) / float64(totalVotes) * 100,
					AbstainRate:  float64(stats.votes.AbstainVotes) / float64(totalVotes) * 100,
				}
				analytics.VotingPatternsByType[proposalType] = pattern
			}
		}
	}

	// Calculate popularity metrics
	if participationCount > 0 {
		analytics.PopularityMetrics.AverageParticipation = totalParticipation / float64(participationCount)
	}

	return analytics
}

// GetDAOHealthMetrics calculates overall DAO health indicators
func (as *AnalyticsSystem) GetDAOHealthMetrics() *DAOHealthMetrics {
	participationMetrics := as.GetGovernanceParticipationMetrics()
	treasuryMetrics := as.GetTreasuryPerformanceMetrics()
	proposalMetrics := as.GetProposalAnalytics()

	health := &DAOHealthMetrics{
		RiskIndicators:  make([]RiskIndicator, 0),
		Recommendations: make([]string, 0),
		LastUpdated:     time.Now().Unix(),
	}

	// Calculate participation health (0-100)
	health.ParticipationHealth = math.Min(participationMetrics.ParticipationRate*2, 100) // Scale participation rate

	// Calculate treasury health (0-100)
	treasuryHealth := 50.0 // Base score
	if treasuryMetrics.SigningEfficiency > 80 {
		treasuryHealth += 20
	} else if treasuryMetrics.SigningEfficiency > 60 {
		treasuryHealth += 10
	}
	if treasuryMetrics.CurrentBalance > 0 {
		treasuryHealth += 20
	}
	if treasuryMetrics.PendingTransactions < treasuryMetrics.ExecutedTransactions/10 {
		treasuryHealth += 10
	}
	health.TreasuryHealth = math.Min(treasuryHealth, 100)

	// Calculate governance health (0-100)
	governanceHealth := 50.0 // Base score
	if proposalMetrics.SuccessRate > 40 && proposalMetrics.SuccessRate < 80 {
		governanceHealth += 25 // Healthy success rate
	}
	if proposalMetrics.QuorumAchievementRate > 70 {
		governanceHealth += 25
	}
	health.GovernanceHealth = math.Min(governanceHealth, 100)

	// Calculate security health (simplified - would need more security metrics)
	health.SecurityHealth = 75.0 // Placeholder

	// Calculate overall score
	health.OverallScore = (health.ParticipationHealth + health.TreasuryHealth +
		health.GovernanceHealth + health.SecurityHealth) / 4

	// Generate risk indicators
	if participationMetrics.ParticipationRate < 20 {
		health.RiskIndicators = append(health.RiskIndicators, RiskIndicator{
			Type:        "Low Participation",
			Severity:    "High",
			Description: "Participation rate is below 20%",
			Score:       100 - participationMetrics.ParticipationRate,
			Mitigation:  "Increase engagement through incentives and communication",
		})
	}

	if proposalMetrics.SuccessRate > 90 {
		health.RiskIndicators = append(health.RiskIndicators, RiskIndicator{
			Type:        "Rubber Stamping",
			Severity:    "Medium",
			Description: "Success rate is too high, indicating lack of critical evaluation",
			Score:       proposalMetrics.SuccessRate - 80,
			Mitigation:  "Encourage more diverse viewpoints and critical discussion",
		})
	}

	if treasuryMetrics.PendingTransactions > treasuryMetrics.ExecutedTransactions/2 {
		health.RiskIndicators = append(health.RiskIndicators, RiskIndicator{
			Type:        "Treasury Bottleneck",
			Severity:    "Medium",
			Description: "Too many pending treasury transactions",
			Score:       float64(treasuryMetrics.PendingTransactions) / float64(treasuryMetrics.ExecutedTransactions) * 100,
			Mitigation:  "Review signing process and consider reducing required signatures",
		})
	}

	// Generate recommendations
	if health.ParticipationHealth < 50 {
		health.Recommendations = append(health.Recommendations, "Implement participation incentives to increase engagement")
	}
	if health.TreasuryHealth < 70 {
		health.Recommendations = append(health.Recommendations, "Optimize treasury management processes")
	}
	if health.GovernanceHealth < 60 {
		health.Recommendations = append(health.Recommendations, "Review proposal processes and voting mechanisms")
	}

	// Determine health trend (simplified)
	if health.OverallScore > 75 {
		health.HealthTrend = "Improving"
	} else if health.OverallScore > 50 {
		health.HealthTrend = "Stable"
	} else {
		health.HealthTrend = "Declining"
	}

	return health
}

// GetAnalyticsSummary provides a comprehensive analytics summary
func (as *AnalyticsSystem) GetAnalyticsSummary() map[string]interface{} {
	return map[string]interface{}{
		"participation_metrics": as.GetGovernanceParticipationMetrics(),
		"treasury_metrics":      as.GetTreasuryPerformanceMetrics(),
		"proposal_analytics":    as.GetProposalAnalytics(),
		"health_metrics":        as.GetDAOHealthMetrics(),
		"generated_at":          time.Now().Unix(),
	}
}
