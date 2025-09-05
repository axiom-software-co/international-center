# Fonts Configuration Module
# Handles system fonts and font rendering

{ config, pkgs, lib, ... }:

{
  fonts = {
    enableDefaultPackages = true;
    
    packages = with pkgs; [
      # Professional coding fonts
      nerd-fonts.fira-code
      nerd-fonts.jetbrains-mono
      nerd-fonts.hack
      nerd-fonts.sauce-code-pro
      nerd-fonts.victor-mono
      nerd-fonts.ubuntu-mono
      
      # System fonts
      dejavu_fonts
      liberation_ttf
      noto-fonts
      noto-fonts-cjk-sans
      noto-fonts-emoji
      
      # Professional presentation fonts
      source-sans-pro
      source-serif-pro
      inter
    ];
    
    fontconfig = {
      enable = true;
      antialias = true;
      cache32Bit = true;
      
      hinting = {
        enable = true;
        style = "slight";
        autohint = false;
      };
      
      subpixel = {
        rgba = "rgb";
        lcdfilter = "default";
      };
      
      defaultFonts = {
        monospace = [ "JetBrains Mono" "FiraCode Nerd Font" ];
        sansSerif = [ "Inter" "Source Sans Pro" ];
        serif = [ "Source Serif Pro" ];
        emoji = [ "Noto Color Emoji" ];
      };
    };
  };
}