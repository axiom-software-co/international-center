# Services Configuration Module
# Handles system services like SSH, DNS, monitoring, etc.

{ config, pkgs, lib, ... }:

{
  services = {
    # SSH configuration
    openssh = {
      enable = true;
      settings = {
        PasswordAuthentication = false;
        PermitRootLogin = "no";
        KbdInteractiveAuthentication = false;
        X11Forwarding = false;
        UseDns = false;
        MaxAuthTries = 3;
        ClientAliveInterval = 300;
        ClientAliveCountMax = 2;
        Compression = true;
        
        # Modern cryptographic algorithms
        Ciphers = [ 
          "chacha20-poly1305@openssh.com" 
          "aes256-gcm@openssh.com" 
          "aes256-ctr" 
        ];
        KexAlgorithms = [ 
          "curve25519-sha256@libssh.org" 
          "diffie-hellman-group16-sha512" 
        ];
        Macs = [ 
          "hmac-sha2-256-etm@openssh.com" 
          "hmac-sha2-512-etm@openssh.com" 
        ];
      };
    };
    
    chrony.enable = true;
    
    # Printing
    printing.enable = true;
    
    # Audio system
    pipewire = {
      enable = true;
      alsa.enable = true;
      pulse.enable = true;
      jack.enable = false;
    };
    pulseaudio.enable = false;
    
    # Bluetooth
    blueman.enable = true;
    
    # System monitoring
    prometheus = {
      enable = true;
      port = 9090;
      
      scrapeConfigs = [
        {
          job_name = "node";
          static_configs = [{
            targets = [ "localhost:9100" ];
          }];
        }
      ];
    };
    
    # Log management
    journald = {
      extraConfig = ''
        SystemMaxUse=2G
        MaxFileSec=1month
        MaxRetentionSec=3month
        Compress=yes
      '';
    };
    
    # Enable resolved for better DNS handling
    resolved = {
      enable = true;
      dnssec = "false";  # Disable DNSSEC for development
      domains = [ "~." ];
      fallbackDns = [ "1.1.1.1" "8.8.8.8" "8.8.4.4" ];
      extraConfig = ''
        DNS=1.1.1.1 8.8.8.8 8.8.4.4
        DNSStubListener=yes
        DNSStubListenerExtra=172.17.0.1:53
        DNSOverTLS=false
        MulticastDNS=true
        LLMNR=true
        Cache=yes
        CacheFromLocalhost=no
        ReadEtcHosts=yes
        ResolveUnicastSingleLabel=no
        StaleRetentionSec=0
      '';
    };
  };
}