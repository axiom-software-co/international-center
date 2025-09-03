package authz.admin_gateway.audit

import rego.v1

# Default audit requirements
default audit_required := false

# Always audit admin gateway access attempts
audit_required := true if {
	input.request.gateway == "admin-gateway"
}

# Audit level based on operation sensitivity
audit_level := "HIGH" if {
	sensitive_operations[input.request.resource]
	input.request.gateway == "admin-gateway"
}

audit_level := "MEDIUM" if {
	standard_operations[input.request.resource]
	input.request.gateway == "admin-gateway"
	not sensitive_operations[input.request.resource]
}

audit_level := "LOW" if {
	input.request.gateway == "admin-gateway"
	not sensitive_operations[input.request.resource]
	not standard_operations[input.request.resource]
}

# Sensitive operations requiring high-level audit
sensitive_operations := {
	"/admin/api/v1/users",
	"/admin/api/v1/roles",
	"/admin/api/v1/permissions", 
	"/admin/api/v1/audit",
	"/admin/api/v1/system-config"
}

# Standard operations requiring medium-level audit
standard_operations := {
	"/admin/api/v1/patients",
	"/admin/api/v1/appointments",
	"/admin/api/v1/medical-records",
	"/admin/api/v1/prescriptions"
}

# Required audit fields
audit_fields := {
	"user_id": input.user.user_id,
	"client_ip": input.request.client_ip,
	"resource": input.request.resource,
	"action": input.request.action,
	"timestamp": time.now_ns(),
	"gateway": input.request.gateway,
	"session_id": input.request.headers["x-session-id"],
	"correlation_id": input.request.headers["x-correlation-id"]
}

# Audit retention requirements
retention_days := 2555 if {  # 7 years for healthcare compliance
	sensitive_operations[input.request.resource]
}

retention_days := 1825 if {  # 5 years for standard operations
	standard_operations[input.request.resource]
	not sensitive_operations[input.request.resource]
}

retention_days := 365 if {  # 1 year for other operations
	not sensitive_operations[input.request.resource]
	not standard_operations[input.request.resource]
}

# Audit destination
audit_destination := "loki" if {
	input.request.gateway == "admin-gateway"
}

# Audit metadata
audit_metadata := {
	"policy_id": "admin_gateway_audit",
	"compliance_framework": "healthcare", 
	"data_classification": "sensitive"
}