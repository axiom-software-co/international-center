# Desktop Environment Configuration Module
# Handles X11, display managers, and desktop environment

{ config, pkgs, lib, ... }:

{
  # Desktop environment
  services.xserver = {
    enable = true;
    
    xkb = {
      layout = "us";
      variant = "";
    };
  };

  # Display and desktop managers
  services.displayManager.gdm.enable = true;
  services.desktopManager.gnome.enable = true;
  
  services.gnome = {
    gnome-keyring.enable = true;
    sushi.enable = true;
  };
}