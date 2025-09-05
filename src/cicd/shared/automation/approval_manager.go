package automation

import (
	"context"
	"fmt"
	"time"
)

// NewApprovalManager creates a new approval manager
func NewApprovalManager() *ApprovalManager {
	am := &ApprovalManager{
		approvalHandlers: make(map[string]ApprovalHandler),
	}
	
	// Register default approval handlers
	am.RegisterHandler("manual", &ManualApprovalHandler{})
	am.RegisterHandler("automated", &AutomatedApprovalHandler{})
	
	return am
}

// RegisterHandler registers an approval handler
func (am *ApprovalManager) RegisterHandler(name string, handler ApprovalHandler) {
	am.approvalHandlers[name] = handler
}

// RequestApproval requests approval for a deployment plan
func (am *ApprovalManager) RequestApproval(ctx context.Context, plan *DeploymentPlan) (*ApprovalResult, error) {
	// Determine which approval handler to use based on environment
	handlerName := am.getApprovalHandlerForEnvironment(plan.Environment)
	
	handler, exists := am.approvalHandlers[handlerName]
	if !exists {
		return nil, fmt.Errorf("approval handler %s not found", handlerName)
	}
	
	return handler.RequestApproval(ctx, plan)
}

// getApprovalHandlerForEnvironment returns the appropriate approval handler
func (am *ApprovalManager) getApprovalHandlerForEnvironment(environment string) string {
	switch environment {
	case "production":
		return "manual" // Production always requires manual approval
	case "staging":
		return "automated" // Staging can use automated approval
	default:
		return "automated"
	}
}

// ManualApprovalHandler handles manual approval processes
type ManualApprovalHandler struct {
	pendingApprovals map[string]*PendingApproval
}

// PendingApproval represents a pending approval request
type PendingApproval struct {
	ID          string
	Plan        *DeploymentPlan
	RequestedAt time.Time
	Approvers   []string
	Status      ApprovalStatus
}

func (mah *ManualApprovalHandler) RequestApproval(ctx context.Context, plan *DeploymentPlan) (*ApprovalResult, error) {
	if mah.pendingApprovals == nil {
		mah.pendingApprovals = make(map[string]*PendingApproval)
	}
	
	approvalID := fmt.Sprintf("approval-%s-%d", plan.Environment, time.Now().Unix())
	
	pending := &PendingApproval{
		ID:          approvalID,
		Plan:        plan,
		RequestedAt: time.Now(),
		Approvers:   mah.getApproversForEnvironment(plan.Environment),
		Status:      ApprovalStatusPending,
	}
	
	mah.pendingApprovals[approvalID] = pending
	
	// Send approval notification (this would integrate with notification systems)
	fmt.Printf("Manual approval requested for deployment to %s (ID: %s)\n", plan.Environment, approvalID)
	fmt.Printf("Required approvers: %v\n", pending.Approvers)
	
	// For demo purposes, return a mock approval result
	// In a real system, this would create a pending approval and return immediately
	return &ApprovalResult{
		ID:         approvalID,
		Status:     ApprovalStatusApproved, // Mock approval
		Approver:   "system-admin",
		ApprovedAt: time.Now(),
		Comments:   "Automatically approved for demonstration",
	}, nil
}

func (mah *ManualApprovalHandler) CheckApprovalStatus(ctx context.Context, approvalID string) (*ApprovalResult, error) {
	pending, exists := mah.pendingApprovals[approvalID]
	if !exists {
		return nil, fmt.Errorf("approval request %s not found", approvalID)
	}
	
	return &ApprovalResult{
		ID:         approvalID,
		Status:     pending.Status,
		ApprovedAt: time.Now(),
	}, nil
}

// getApproversForEnvironment returns required approvers for an environment
func (mah *ManualApprovalHandler) getApproversForEnvironment(environment string) []string {
	switch environment {
	case "production":
		return []string{"infrastructure-lead", "security-team", "cto"}
	case "staging":
		return []string{"infrastructure-lead"}
	default:
		return []string{"developer"}
	}
}

// AutomatedApprovalHandler handles automated approval processes
type AutomatedApprovalHandler struct {
	rules []AutomatedApprovalRule
}

// AutomatedApprovalRule defines rules for automated approval
type AutomatedApprovalRule struct {
	Environment string
	Conditions  []ApprovalCondition
	AutoApprove bool
}

// ApprovalCondition defines conditions for approval
type ApprovalCondition struct {
	Name  string
	Check func(plan *DeploymentPlan) bool
}

func (aah *AutomatedApprovalHandler) RequestApproval(ctx context.Context, plan *DeploymentPlan) (*ApprovalResult, error) {
	// Check automated approval rules
	if aah.canAutoApprove(plan) {
		return &ApprovalResult{
			ID:         fmt.Sprintf("auto-approval-%s-%d", plan.Environment, time.Now().Unix()),
			Status:     ApprovalStatusApproved,
			Approver:   "automated-system",
			ApprovedAt: time.Now(),
			Comments:   "Automatically approved based on deployment rules",
		}, nil
	}
	
	// Fall back to manual approval if automated approval is not possible
	return &ApprovalResult{
		ID:       fmt.Sprintf("manual-required-%s-%d", plan.Environment, time.Now().Unix()),
		Status:   ApprovalStatusPending,
		Comments: "Automated approval criteria not met, manual approval required",
	}, nil
}

func (aah *AutomatedApprovalHandler) CheckApprovalStatus(ctx context.Context, approvalID string) (*ApprovalResult, error) {
	// This would check the status of an automated approval
	return &ApprovalResult{
		ID:     approvalID,
		Status: ApprovalStatusApproved,
	}, nil
}

// canAutoApprove checks if a deployment can be automatically approved
func (aah *AutomatedApprovalHandler) canAutoApprove(plan *DeploymentPlan) bool {
	// Define automated approval criteria
	conditions := []struct {
		name  string
		check func() bool
	}{
		{
			name: "Non-production environment",
			check: func() bool {
				return plan.Environment != "production"
			},
		},
		{
			name: "Low-risk deployment",
			check: func() bool {
				// This would analyze the deployment for risk factors
				return true
			},
		},
		{
			name: "All validations passed",
			check: func() bool {
				// This would check that all pre-deployment validations passed
				return true
			},
		},
		{
			name: "Off-peak hours",
			check: func() bool {
				// Check if deployment is during off-peak hours
				hour := time.Now().Hour()
				return hour < 8 || hour > 18 // Outside business hours
			},
		},
	}
	
	// All conditions must be met for auto-approval
	for _, condition := range conditions {
		if !condition.check() {
			fmt.Printf("Auto-approval condition not met: %s\n", condition.name)
			return false
		}
	}
	
	return true
}

// ApprovalPolicy defines approval policies
type ApprovalPolicy struct {
	Environment        string
	RequiredApprovers  int
	ApproverGroups     []string
	TimeoutDuration    time.Duration
	EscalationRules    []EscalationRule
}

// EscalationRule defines escalation rules for approvals
type EscalationRule struct {
	TriggerAfter time.Duration
	Escalatees   []string
	Action       EscalationAction
}

// EscalationAction defines escalation actions
type EscalationAction string

const (
	EscalationActionNotify      EscalationAction = "notify"
	EscalationActionAutoApprove EscalationAction = "auto_approve"
	EscalationActionDeny        EscalationAction = "deny"
)

// PolicyManager manages approval policies
type PolicyManager struct {
	policies map[string]*ApprovalPolicy
}

// NewPolicyManager creates a new policy manager
func NewPolicyManager() *PolicyManager {
	pm := &PolicyManager{
		policies: make(map[string]*ApprovalPolicy),
	}
	
	// Set default policies
	pm.setDefaultPolicies()
	
	return pm
}

// setDefaultPolicies sets default approval policies
func (pm *PolicyManager) setDefaultPolicies() {
	// Production policy
	pm.policies["production"] = &ApprovalPolicy{
		Environment:       "production",
		RequiredApprovers: 2,
		ApproverGroups:    []string{"infrastructure-leads", "security-team"},
		TimeoutDuration:   24 * time.Hour,
		EscalationRules: []EscalationRule{
			{
				TriggerAfter: 4 * time.Hour,
				Escalatees:   []string{"cto", "engineering-director"},
				Action:       EscalationActionNotify,
			},
		},
	}
	
	// Staging policy
	pm.policies["staging"] = &ApprovalPolicy{
		Environment:       "staging",
		RequiredApprovers: 1,
		ApproverGroups:    []string{"infrastructure-leads"},
		TimeoutDuration:   2 * time.Hour,
		EscalationRules: []EscalationRule{
			{
				TriggerAfter: 1 * time.Hour,
				Escalatees:   []string{"infrastructure-leads"},
				Action:       EscalationActionAutoApprove,
			},
		},
	}
}

// GetPolicy returns the approval policy for an environment
func (pm *PolicyManager) GetPolicy(environment string) *ApprovalPolicy {
	return pm.policies[environment]
}

// UpdatePolicy updates the approval policy for an environment
func (pm *PolicyManager) UpdatePolicy(environment string, policy *ApprovalPolicy) {
	pm.policies[environment] = policy
}

// ApprovalAuditLog tracks approval history
type ApprovalAuditLog struct {
	entries []ApprovalAuditEntry
}

// ApprovalAuditEntry represents an audit log entry
type ApprovalAuditEntry struct {
	Timestamp   time.Time
	ApprovalID  string
	Environment string
	Action      string
	Actor       string
	Details     map[string]interface{}
}

// NewApprovalAuditLog creates a new approval audit log
func NewApprovalAuditLog() *ApprovalAuditLog {
	return &ApprovalAuditLog{
		entries: []ApprovalAuditEntry{},
	}
}

// LogApprovalAction logs an approval action
func (aal *ApprovalAuditLog) LogApprovalAction(approvalID, environment, action, actor string, details map[string]interface{}) {
	entry := ApprovalAuditEntry{
		Timestamp:   time.Now(),
		ApprovalID:  approvalID,
		Environment: environment,
		Action:      action,
		Actor:       actor,
		Details:     details,
	}
	
	aal.entries = append(aal.entries, entry)
	
	// Log to console for visibility
	fmt.Printf("Approval audit: %s performed %s on approval %s for environment %s\n", 
		actor, action, approvalID, environment)
}

// GetApprovalHistory returns approval history for an environment
func (aal *ApprovalAuditLog) GetApprovalHistory(environment string) []ApprovalAuditEntry {
	var history []ApprovalAuditEntry
	for _, entry := range aal.entries {
		if entry.Environment == environment {
			history = append(history, entry)
		}
	}
	return history
}