# Hardware Configuration Module
# Handles graphics, CPU, Bluetooth, and power management

{ config, pkgs, lib, ... }:

{
  hardware = {
    graphics = {
      enable = true;
      enable32Bit = true;
      
      extraPackages = with pkgs; [
        intel-media-driver
        vaapiIntel
        vaapiVdpau
        libvdpau-va-gl
      ] ++ lib.optionals (config.services.xserver.videoDrivers or [] == [ "nvidia" ]) [
        nvidia-vaapi-driver
      ];
    };
    
    cpu.intel.updateMicrocode = lib.mkDefault config.hardware.enableRedistributableFirmware;
    cpu.amd.updateMicrocode = lib.mkDefault config.hardware.enableRedistributableFirmware;
    
    bluetooth = {
      enable = true;
      powerOnBoot = true;
      settings = {
        General = {
          Enable = "Source,Sink,Media,Socket";
          Experimental = true;
        };
      };
    };
  };

  # Performance optimizations
  zramSwap = {
    enable = true;
    algorithm = "zstd";
    memoryPercent = 25;
  };
  
  powerManagement = {
    enable = true;
    cpuFreqGovernor = lib.mkDefault "ondemand";
  };
}