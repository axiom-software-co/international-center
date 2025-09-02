package authz.public_gateway.anonymous

import rego.v1

# Default deny - anonymous requests denied unless explicitly allowed
default allow := false

# Allow anonymous access to public endpoints
allow if {
	public_endpoints[input.request.resource]
	input.request.gateway == "public-gateway"
}

# Allow anonymous access to health check endpoints
allow if {
	startswith(input.request.resource, "/api/v1/public/")
	input.request.gateway == "public-gateway"
}

# Allow anonymous access to information endpoints
allow if {
	information_endpoints[input.request.resource]
	input.request.gateway == "public-gateway"
}

# Public endpoints that allow anonymous access
public_endpoints := {
	"/api/v1/public/health",
	"/api/v1/public/info",
	"/api/v1/public/status",
	"/api/v1/public/version"
}

# Information endpoints for public consumption
information_endpoints := {
	"/api/v1/info/services",
	"/api/v1/info/locations", 
	"/api/v1/info/contact",
	"/api/v1/info/hours"
}

# Protected endpoints requiring authentication
protected_endpoints := {
	"/api/v1/patients",
	"/api/v1/appointments",
	"/api/v1/medical-records",
	"/api/v1/prescriptions",
	"/api/v1/users"
}

# Reason for access decision
reason := "public endpoint allows anonymous access" if {
	allow
	public_endpoints[input.request.resource]
}

reason := "information endpoint allows anonymous access" if {
	allow
	information_endpoints[input.request.resource]
}

reason := "public API allows anonymous access" if {
	allow
	startswith(input.request.resource, "/api/v1/public/")
}

reason := "authentication required" if {
	not allow
	protected_endpoints[input.request.resource]
}

reason := "endpoint not found or forbidden" if {
	not allow
	not protected_endpoints[input.request.resource]
	not public_endpoints[input.request.resource]
	not information_endpoints[input.request.resource]
	not startswith(input.request.resource, "/api/v1/public/")
}

# Policy metadata
metadata := {
	"policy_id": "public_gateway_anonymous",
	"version": "1.0.0",
	"access_type": "anonymous"
}