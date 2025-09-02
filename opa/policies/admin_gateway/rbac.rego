package authz.admin_gateway.rbac

import rego.v1

# Default deny - all requests are denied unless explicitly allowed
default allow := false

# Allow admin users access to all admin endpoints
allow if {
	input.user.roles[_] == "admin"
	startswith(input.request.resource, "/admin/")
	input.request.gateway == "admin-gateway"
}

# Allow healthcare staff access to patient-related endpoints
allow if {
	input.user.roles[_] == "healthcare_staff"
	healthcare_staff_allowed_resources[input.request.resource]
	input.request.gateway == "admin-gateway"
}

# Allow user managers access to user management endpoints  
allow if {
	input.user.roles[_] == "user_manager"
	user_manager_allowed_resources[input.request.resource]
	input.request.gateway == "admin-gateway"
}

# Healthcare staff allowed resources
healthcare_staff_allowed_resources := {
	"/admin/api/v1/patients",
	"/admin/api/v1/appointments", 
	"/admin/api/v1/medical-records",
	"/admin/api/v1/prescriptions"
}

# User manager allowed resources
user_manager_allowed_resources := {
	"/admin/api/v1/users",
	"/admin/api/v1/roles",
	"/admin/api/v1/permissions"
}

# Reason for access decision
reason := "admin role permits all admin operations" if {
	input.user.roles[_] == "admin"
	startswith(input.request.resource, "/admin/")
}

reason := "healthcare staff role permits patient operations" if {
	input.user.roles[_] == "healthcare_staff"
	healthcare_staff_allowed_resources[input.request.resource]
}

reason := "user manager role permits user management operations" if {
	input.user.roles[_] == "user_manager" 
	user_manager_allowed_resources[input.request.resource]
}

reason := "insufficient permissions" if {
	not allow
	input.user.user_id != ""
}

reason := "authentication required" if {
	not allow
	input.user.user_id == ""
}

# Policy metadata
metadata := {
	"policy_id": "admin_gateway_rbac",
	"version": "1.0.0",
	"last_updated": "2025-09-02"
}