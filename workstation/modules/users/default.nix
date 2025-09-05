# Users Configuration Module  
# Handles user accounts and SSH configuration

{ config, pkgs, lib, ... }:

{
  users = {
    mutableUsers = false;
    
    users.tojkuv = {
      isNormalUser = true;
      description = "Tojkuv";
      extraGroups = [ 
        "wheel" "networkmanager" "podman" "kvm" "libvirtd"
        "audio" "video" "input" "storage" "adbusers"
      ];
      shell = pkgs.bash;
      
      openssh.authorizedKeys.keys = [
        # Add your SSH keys here for automation
      ];
      
      hashedPassword = "$6$h/mWSrAzrdzXRgjV$w.EpkrSlFL3bQ61dqyZWFHaXJMYrnbPB6W1QO8ClkXhe4j2O2MCrBWU7KZl9PD.BYlA2VivyzCLkoFM6XYQ6c.";
    };
  };
}