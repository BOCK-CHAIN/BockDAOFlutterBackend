package dao

import "fmt"

// ErrorCode represents different types of DAO errors
type ErrorCode int

const (
	ErrInsufficientTokens   ErrorCode = 4001
	ErrProposalNotFound     ErrorCode = 4002
	ErrVotingClosed         ErrorCode = 4003
	ErrUnauthorized         ErrorCode = 4004
	ErrInvalidSignature     ErrorCode = 4005
	ErrQuorumNotMet         ErrorCode = 4006
	ErrTreasuryInsufficient ErrorCode = 4007
	ErrInvalidProposal      ErrorCode = 4008
	ErrDuplicateVote        ErrorCode = 4009
	ErrInvalidDelegation    ErrorCode = 4010
	ErrInvalidTimeframe     ErrorCode = 4011
	ErrInvalidThreshold     ErrorCode = 4012
	ErrTokenTransferFailed  ErrorCode = 4013
	ErrInvalidVoteChoice    ErrorCode = 4014
	ErrProposalExpired      ErrorCode = 4015
	ErrSecurityViolation    ErrorCode = 4016
	ErrEmergencyActive      ErrorCode = 4017
	ErrFunctionPaused       ErrorCode = 4018
	ErrRoleExpired          ErrorCode = 4019
	ErrAuditAccessDenied    ErrorCode = 4020
)

// DAOError represents a DAO-specific error
type DAOError struct {
	Code    ErrorCode
	Message string
	Details map[string]interface{}
}

// Error implements the error interface
func (e *DAOError) Error() string {
	return fmt.Sprintf("DAO Error %d: %s", e.Code, e.Message)
}

// NewDAOError creates a new DAO error
func NewDAOError(code ErrorCode, message string, details map[string]interface{}) *DAOError {
	return &DAOError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Common DAO errors
var (
	ErrInsufficientTokensForProposal = NewDAOError(
		ErrInsufficientTokens,
		"insufficient tokens to create proposal",
		nil,
	)

	ErrInsufficientTokensForVote = NewDAOError(
		ErrInsufficientTokens,
		"insufficient tokens to cast vote",
		nil,
	)

	ErrProposalNotFoundError = NewDAOError(
		ErrProposalNotFound,
		"proposal not found",
		nil,
	)

	ErrVotingPeriodClosed = NewDAOError(
		ErrVotingClosed,
		"voting period has ended",
		nil,
	)

	ErrVotingNotStarted = NewDAOError(
		ErrVotingClosed,
		"voting period has not started",
		nil,
	)

	ErrUnauthorizedAccess = NewDAOError(
		ErrUnauthorized,
		"unauthorized access to DAO function",
		nil,
	)

	ErrInvalidSignatureError = NewDAOError(
		ErrInvalidSignature,
		"invalid transaction signature",
		nil,
	)

	ErrQuorumNotMetError = NewDAOError(
		ErrQuorumNotMet,
		"quorum threshold not met",
		nil,
	)

	ErrTreasuryInsufficientFunds = NewDAOError(
		ErrTreasuryInsufficient,
		"insufficient treasury funds",
		nil,
	)

	ErrInvalidProposalFormat = NewDAOError(
		ErrInvalidProposal,
		"invalid proposal format or content",
		nil,
	)

	ErrDuplicateVoteError = NewDAOError(
		ErrDuplicateVote,
		"user has already voted on this proposal",
		nil,
	)

	ErrInvalidDelegationError = NewDAOError(
		ErrInvalidDelegation,
		"invalid delegation parameters",
		nil,
	)

	ErrInvalidTimeframeError = NewDAOError(
		ErrInvalidTimeframe,
		"invalid proposal timeframe",
		nil,
	)

	ErrInvalidThresholdError = NewDAOError(
		ErrInvalidThreshold,
		"invalid voting threshold",
		nil,
	)

	ErrTokenTransferFailedError = NewDAOError(
		ErrTokenTransferFailed,
		"token transfer operation failed",
		nil,
	)

	ErrInvalidVoteChoiceError = NewDAOError(
		ErrInvalidVoteChoice,
		"invalid vote choice",
		nil,
	)

	ErrProposalExpiredError = NewDAOError(
		ErrProposalExpired,
		"proposal has expired",
		nil,
	)

	ErrSecurityViolationError = NewDAOError(
		ErrSecurityViolation,
		"security violation detected",
		nil,
	)

	ErrEmergencyActiveError = NewDAOError(
		ErrEmergencyActive,
		"emergency mode is active",
		nil,
	)

	ErrFunctionPausedError = NewDAOError(
		ErrFunctionPaused,
		"function is currently paused",
		nil,
	)

	ErrRoleExpiredError = NewDAOError(
		ErrRoleExpired,
		"user role has expired",
		nil,
	)

	ErrAuditAccessDeniedError = NewDAOError(
		ErrAuditAccessDenied,
		"access to audit log denied",
		nil,
	)
)
