# Boot Configuration Module
# Handles boot loader, kernel settings, and system startup

{ config, pkgs, lib, ... }:

{
  boot = {
    loader = {
      systemd-boot.enable = true;
      efi.canTouchEfiVariables = true;
      timeout = 3;
    };
    
    # Performance optimizations for development workloads
    kernel = {
      sysctl = {
        # Memory management for development
        "vm.swappiness" = 10;
        "vm.dirty_ratio" = 15;
        "vm.dirty_background_ratio" = 5;
        "vm.max_map_count" = 2147483642;
        
        # Network optimizations for API development and Kubernetes
        "net.core.default_qdisc" = "fq";
        "net.ipv4.tcp_congestion_control" = "bbr";
        "net.core.rmem_max" = 134217728;
        "net.core.wmem_max" = 134217728;
        "net.ipv4.tcp_rmem" = "4096 87380 134217728";
        "net.ipv4.tcp_wmem" = "4096 65536 134217728";
        
        # IPv6 configuration for development
        "net.ipv6.conf.all.disable_ipv6" = 0;
        "net.ipv6.conf.default.disable_ipv6" = 0;
        "net.ipv6.conf.lo.disable_ipv6" = 0;
        "net.ipv6.conf.all.forwarding" = 1;
        "net.ipv4.ip_forward" = 1;
        
        # File system optimizations
        "fs.file-max" = 2097152;
        "fs.nr_open" = 1048576;
        "fs.inotify.max_user_watches" = 524288;
        "fs.inotify.max_user_instances" = 8192;
        
        # Basic kernel settings
        "kernel.keys.maxkeys" = 2000;
        "kernel.keys.maxbytes" = 2000000;
      };
    };
    
    supportedFilesystems = [ "btrfs" "ntfs" ];
    tmp.useTmpfs = true;
    tmp.tmpfsSize = "50%";
    
    # Basic kernel modules
    kernelModules = [
      "kvm-intel" "kvm-amd"
      "vfio-pci"
    ];
    
    # Basic module loading
    initrd.kernelModules = [ "overlay" "br_netfilter" ];
  };
}