package models

import "time"

// ApprovalStatus represents the status of an approval request
type ApprovalStatus string

const (
	// ApprovalStatusPending indicates the approval is waiting for a decision
	ApprovalStatusPending ApprovalStatus = "pending"
	// ApprovalStatusApproved indicates the approval was granted
	ApprovalStatusApproved ApprovalStatus = "approved"
	// ApprovalStatusRejected indicates the approval was denied
	ApprovalStatusRejected ApprovalStatus = "rejected"
	// ApprovalStatusExpired indicates the approval request has expired
	ApprovalStatusExpired ApprovalStatus = "expired"
)

// ApprovalRequest represents a request for approval
type ApprovalRequest struct {
	ID            string
	ExecutionID   string
	StageName     string
	Approvers     []string
	Status        ApprovalStatus
	RequestedAt   time.Time
	RespondedAt   time.Time
	ResponderID   string
	ResponderName string
	Comment       string
	ExpiresAt     time.Time
}

// ApprovalResponse represents a response to an approval request
type ApprovalResponse struct {
	RequestID     string
	Approved      bool
	ResponderID   string
	ResponderName string
	Comment       string
	RespondedAt   time.Time
} 