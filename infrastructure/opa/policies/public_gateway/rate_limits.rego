package authz.public_gateway.rate_limits

import rego.v1

# Default rate limits for public gateway (IP-based)
default requests_per_window := 1000
default time_window_seconds := 60

# Standard IP-based rate limits for public gateway
requests_per_window := 1000 if {
	input.request.gateway == "public-gateway"
	input.request.client_ip != ""
}

# Lower limits for suspicious patterns
requests_per_window := 100 if {
	input.request.gateway == "public-gateway"
	suspicious_ip_patterns[input.request.client_ip]
}

# Higher limits for whitelisted IPs (internal networks)
requests_per_window := 5000 if {
	input.request.gateway == "public-gateway"  
	whitelisted_ip_ranges[_]
	net.cidr_contains(whitelisted_ip_ranges[_], input.request.client_ip)
}

# Endpoint-specific rate limits
requests_per_window := 100 if {
	input.request.gateway == "public-gateway"
	resource_specific_limits[input.request.resource]
	not whitelisted_ip_ranges[_]
}

# Time window is always 1 minute for public gateway
time_window_seconds := 60 if {
	input.request.gateway == "public-gateway"
}

# Rate limit key - always use IP address for public gateway
rate_limit_key := input.request.client_ip if {
	input.request.gateway == "public-gateway"
	input.request.client_ip != ""
}

# Whitelisted IP ranges (internal networks)
whitelisted_ip_ranges := [
	"10.0.0.0/8",
	"172.16.0.0/12", 
	"192.168.0.0/16",
	"127.0.0.0/8"
]

# Suspicious IP patterns (example - would be dynamic in production)
suspicious_ip_patterns := {
	"192.0.2.1",    # Example suspicious IP
	"203.0.113.1"   # Example suspicious IP
}

# Resources with specific rate limits
resource_specific_limits := {
	"/api/v1/info/services": 100,
	"/api/v1/info/locations": 100, 
	"/api/v1/info/contact": 50,
	"/api/v1/public/health": 500
}

# Rate limit metadata
rate_limit_metadata := {
	"policy_id": "public_gateway_rate_limits",
	"limit_type": "ip_based", 
	"enforcement_level": "standard"
}