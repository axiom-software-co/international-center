# Networking Configuration Module
# Handles network settings, firewall, and DNS configuration

{ config, pkgs, lib, ... }:

let
  developmentEnvironment = "dev-unstable";
in

{
  networking = {
    hostName = "dev-workstation-${developmentEnvironment}";
    networkmanager.enable = true;
    # nameservers = [ "1.1.1.1" "8.8.8.8" "8.8.4.4" ]; # Disabled - let resolved handle DNS
    
    # Enable IPv6 support explicitly
    enableIPv6 = true;
    
    # Simple firewall configuration
    firewall = {
      enable = true;
      allowedTCPPorts = [
        22    # SSH
        80 443 # HTTP/HTTPS
        3000  # Development server
        22220 # SSH forwarded port for Windows Server
        33389 # RDP forwarded port for Windows Server
        139 445 # SMB/CIFS for Windows Server VM file sharing
      ];
      allowedUDPPorts = [
        137 138 # NetBIOS for SMB discovery
      ];
    };
    
    # Simple network configuration
    interfaces = {
      lo = {
      };
    };
  };
}