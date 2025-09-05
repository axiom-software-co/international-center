# Virtualisation Configuration Module
# Handles Podman, libvirtd, and container settings

{ config, pkgs, lib, ... }:

{
  virtualisation = {
    podman = {
      enable = true;
      
      # Docker compatibility for seamless transition
      dockerCompat = true;
      dockerSocket.enable = true;
      
      # Default network settings
      defaultNetwork.settings.dns_enabled = true;
      
      # Auto-update and cleanup
      autoPrune = {
        enable = true;
        dates = "weekly";
        flags = [ "--all" ];
      };
      
      # Extra packages for full compatibility
      extraPackages = with pkgs; [ 
        buildah 
        skopeo 
        runc 
        crun
        fuse-overlayfs
        slirp4netns
      ];
    };
    
    libvirtd = {
      enable = true;
      qemu = {
        package = pkgs.qemu_kvm;
        runAsRoot = false;
        swtpm.enable = true;
        ovmf = {
          enable = true;
          packages = [ pkgs.OVMFFull.fd ];
        };
      };
    };
    
    containers.enable = true;
  };
  
  # Systemd optimizations for container orchestration
  systemd = {
    settings.Manager = {
      DefaultTimeoutStopSec = 30;
      DefaultTimeoutStartSec = 30; 
      DefaultLimitNOFILE = 1048576;
      DefaultLimitNPROC = 1048576;
      DefaultLimitMEMLOCK = "infinity";
    };
    
    # Enable services needed for container orchestration
    services.systemd-resolved.enable = true;
  };
}