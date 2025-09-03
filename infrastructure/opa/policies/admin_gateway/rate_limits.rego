package authz.admin_gateway.rate_limits

import rego.v1

# Default rate limits for admin gateway
default requests_per_window := 100
default time_window_seconds := 60

# Admin users get higher rate limits
requests_per_window := 200 if {
	input.user.roles[_] == "admin"
	input.request.gateway == "admin-gateway"
}

# Healthcare staff get standard rate limits
requests_per_window := 100 if {
	input.user.roles[_] == "healthcare_staff" 
	input.request.gateway == "admin-gateway"
}

# Regular users get lower rate limits
requests_per_window := 50 if {
	input.user.roles[_] == "user"
	not input.user.roles[_] == "admin"
	not input.user.roles[_] == "healthcare_staff"
	input.request.gateway == "admin-gateway"
}

# Time window is always 1 minute for admin gateway
time_window_seconds := 60 if {
	input.request.gateway == "admin-gateway"
}

# Rate limit key - use user ID for user-based limiting
rate_limit_key := input.user.user_id if {
	input.user.user_id != ""
	input.request.gateway == "admin-gateway"
}

# Fallback to IP-based limiting if no user ID
rate_limit_key := input.request.client_ip if {
	input.user.user_id == ""
	input.request.gateway == "admin-gateway"
}

# Rate limit metadata
rate_limit_metadata := {
	"policy_id": "admin_gateway_rate_limits",
	"limit_type": "user_based",
	"enforcement_level": "strict"
}