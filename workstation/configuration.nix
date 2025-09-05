# NixOS Configuration - Professional Development Machine
# Modular, Clean, Zero Technical Debt Architecture
# System-level configuration with all packages and programs

{ config, pkgs, lib, ... }:

let
  # Version and metadata management
  systemVersion = "2025.09.04-no-home-manager";
  developmentEnvironment = "dev-unstable";
in

{
  imports = [
    # Hardware configuration (auto-generated)
    ./hardware-configuration.nix
    
    # Modular configuration imports
    ./modules/boot
    ./modules/networking  
    ./modules/users
    ./modules/services
    ./modules/security
    ./modules/virtualisation
    ./modules/hardware
    ./modules/desktop
    ./modules/fonts
    ./modules/packages
    ./modules/programs
  ];

  # System metadata
  system = {
    stateVersion = "25.05";
    configurationRevision = systemVersion;
    nixos.label = "dev-workstation-${systemVersion}";
    
    autoUpgrade = {
      enable = false;
      channel = "https://nixos.org/channels/nixos-unstable";
    };
  };


  # Nix configuration
  nixpkgs.config = {
    allowUnfree = true;
    allowBroken = false;
  };

  nix = {
    package = pkgs.nixVersions.stable;
    
    settings = {
      experimental-features = [ "nix-command" "flakes" "auto-allocate-uids" "cgroups" ];
      auto-optimise-store = true;
      max-jobs = "auto";
      cores = 0;
      keep-outputs = true;
      keep-derivations = true;
      builders-use-substitutes = true;
      fallback = false;

      # Development-friendly settings
      accept-flake-config = true;
      warn-dirty = false;
      allow-import-from-derivation = true;
      
      substituters = [
        "https://cache.nixos.org/"
        "https://nix-community.cachix.org"
        "https://devenv.cachix.org"
      ];
      
      trusted-public-keys = [
        "cache.nixos.org-1:6NCHdD59X431o0gWypbMrAURkbJ16ZPMQFGspcDShjY="
        "nix-community.cachix.org-1:mB9FSh9qf2dCimDSUo8Zy7bkq5CX+/rkCWyvRCYg3Fs="
        "devenv.cachix.org-1:w1cLUi8dv3hnoSPGAuibQv+f9TZLr6cv/Hm9XgU50cw="
      ];
      
      require-sigs = true;
      trusted-users = [ "root" "tojkuv" ];
      allowed-users = [ "@wheel" ];
      sandbox = true;
      restrict-eval = false; # Needed for home-manager
    };
    
    gc = {
      automatic = true;
      dates = "daily";
      options = "--delete-older-than 7d --max-freed $((64 * 1024**3))";
    };
    
    optimise = {
      automatic = true;
      dates = [ "03:45" ];
    };
    
    nrBuildUsers = 32;
    
    extraOptions = ''
      max-silent-time = 3600
      timeout = 7200
      log-lines = 100
      connect-timeout = 10
      stalled-download-timeout = 30
      accept-flake-config = true
      warn-dirty = false
    '';
  };
}