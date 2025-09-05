# System Packages Configuration Module
# Comprehensive development environment packages

{ config, pkgs, lib, ... }:

let
  # ===== PYTHON DEVELOPMENT ENVIRONMENT =====
  pythonEnv = pkgs.python312.withPackages (ps: with ps; [
    # Web frameworks and API development
    fastapi
    flask
    django
    uvicorn
    gunicorn
    
    # Data science and analysis
    pandas
    numpy
    matplotlib
    seaborn
    jupyter
    
    # Database connectivity
    sqlalchemy
    alembic
    psycopg2
    redis
    asyncpg
    
    # HTTP and API clients
    httpx
    aiohttp
    requests
    
    # Development and testing
    pytest
    pytest-asyncio
    black
    isort
    mypy
    flake8
    
    # Configuration and utilities
    python-dotenv
    pydantic
    pydantic-settings
    typer
    rich
    loguru
    
    # Optional AI/ML libraries (for users who need them)
    openai
    anthropic
    langchain
    langchain-community
  ]);
in

{
  environment.systemPackages = with pkgs; [
    # ===== ESSENTIAL SYSTEM TOOLS =====
    vim neovim
    wget curl
    git
    htop
    tree file
    unzip zip
    
    # Networking tools
    inetutils
    nmap
    nettools
    iproute2
    iptables
    iputils
    dig
    
    # Security and monitoring
    gnupg
    openssh
    openssl
    
    # System administration
    systemctl-tui
    nixos-rebuild
    
    # Container runtime (system-level) - Podman for better NixOS integration
    podman
    podman-compose
    podman-tui
    kubectl
    
    # System dependencies for .NET Core and containers
    nix-ld
    stdenv.cc.cc
    glibc
    
    # SSL/TLS Certificate management
    cacert
    
    # Windows Server Management Tools
    powershell                             # PowerShell Core for Windows remote management
    sshpass                               # SSH password authentication automation
    remmina                              # Remote desktop client for Windows RDP
    freerdp                              # FreeRDP for Windows remote desktop
    expect                               # Automated interaction for command-line tools
    
    # File sharing tools
    samba                                # SMB/CIFS server for Windows file sharing
    cifs-utils                          # CIFS utilities for mounting Windows shares
    
    # System Process and Resource Management
    procps                                 # Process monitoring and management utilities
    util-linux                            # Essential Linux system utilities
    
    # ===== PROGRAMMING LANGUAGE RUNTIMES =====
    # Node.js ecosystem
    nodejs_20                   # Primary Node.js LTS
    yarn                        # Package manager
    pnpm                        # Fast package manager
    bun                         # Fast JavaScript runtime
    
    # Python environment (defined above)
    pythonEnv                   # Complete Python development stack
    pipx                        # Install Python applications in isolated environments
    
    # JVM languages
    jdk21                       # Java Development Kit
    maven                       # Java build tool
    gradle                      # Java/Kotlin build tool
    kotlin                      # Kotlin programming language
    scala                       # Scala programming language
    sbt                         # Scala build tool
    
    # Go development
    go                         # Go programming language
    gopls                      # Go language server
    golangci-lint             # Go linting
    delve                     # Go debugger
    
    # Rust development
    rustc                     # Rust compiler
    cargo                     # Rust package manager
    rustfmt                   # Rust formatter
    clippy                    # Rust linting
    rust-analyzer            # Rust language server
    
    # Ruby development
    ruby_3_3                  # Ruby programming language
    
    # PHP development
    php83                     # PHP 8.3
    php83Packages.composer    # PHP dependency manager
    
    # C/C++ development
    gcc                       # GNU Compiler Collection
    cmake                     # Build system
    ninja                     # Build system
    gdb                       # GNU Debugger
    lldb                      # LLVM Debugger
    
    # ===== .NET 9 DEVELOPMENT ENVIRONMENT =====
    dotnet-sdk_9                           # Complete .NET 9 SDK with build tools, MSBuild, and NuGet
    dotnet-aspnetcore_9                    # ASP.NET Core 9 runtime for web applications
    dotnetCorePackages.runtime_9_0-bin     # .NET 9 runtime binary for application execution
    dotnetCorePackages.aspnetcore_9_0      # ASP.NET Core 9 runtime package
    dotnetCorePackages.sdk_9_0             # Additional .NET 9 SDK components and tools
    
    # ===== CLOUD DEVELOPMENT TOOLS =====
    # AWS ecosystem
    awscli2                   # AWS CLI
    aws-sam-cli              # AWS SAM CLI
    eksctl                    # Amazon EKS CLI
    aws-iam-authenticator    # AWS IAM Authenticator
    
    # Google Cloud Platform
    google-cloud-sdk         # Google Cloud SDK
    
    # Microsoft Azure
    azure-cli                # Azure CLI
    azurite                  # Azure Storage Emulator
    
    # Other cloud providers
    doctl                    # DigitalOcean CLI
    linode-cli               # Linode CLI
    
    # Cloud platform tools
    dapr-cli
    pulumi
    go-migrate
    wrangler
    grafana
    hcp
    
    # ===== INFRASTRUCTURE AS CODE =====
    terraform                # Infrastructure provisioning
    terragrunt              # Terraform wrapper
    terraform-docs          # Terraform documentation generator
    tflint                  # Terraform linter
    skaffold
    
    # Configuration management
    ansible                 # Configuration management
    ansible-lint           # Ansible linter
    
    # Alternative IaC tools
    pulumi-bin             # Modern infrastructure as code
    
    # ===== CONTAINER AND ORCHESTRATION =====
    # Kubernetes
    kubernetes-helm-wrapped # Kubernetes package manager
    kustomize              # Kubernetes configuration management
    k9s                    # Kubernetes TUI
    
    # Docker ecosystem
    docker-compose         # Multi-container orchestration
    
    # ===== DATABASE TOOLS (Container-First Development) =====
    # PostgreSQL tools
    pgcli                 # PostgreSQL interactive terminal
    minio
    
    # MySQL/MariaDB tools
    mysql-workbench      # MySQL GUI client
    
    # Database GUI clients
    dbeaver-bin          # Universal database tool
    
    # ===== VERSION CONTROL =====
    gh                   # GitHub CLI
    gitlab-runner        # GitLab CI runner
    git-lfs             # Git Large File Storage
    
    firebase-tools
    
    # ===== CODE EDITORS AND IDEs =====
    # Terminal editors
    emacs               # Emacs editor
    
    # GUI editors
    vscode              # Visual Studio Code
    zed-editor          # Modern collaborative editor
    
    # ===== WEB BROWSERS =====
    firefox             # Privacy-focused browser
    microsoft-edge      # Professional browser
    
    # ===== USER APPLICATIONS =====
    zoom-us
    
    # ===== API DEVELOPMENT =====
    httpie             # HTTP client
    postman           # API development platform
    insomnia          # REST/GraphQL client
    
    # ===== DATA FORMATS =====
    jq                 # JSON processor
    yq                 # YAML processor
    xmlstarlet         # XML processor
    
    # ===== MODERN CLI TOOLS =====
    # File operations
    bat                # Enhanced cat
    eza                # Enhanced ls
    fd                 # Enhanced find
    ripgrep            # Enhanced grep
    fzf                # Fuzzy finder
    zoxide             # Enhanced cd
    lsof
    
    # System monitoring
    btop               # Modern process monitor
    ncdu               # Disk usage analyzer
    
    # ===== DEVELOPMENT UTILITIES =====
    # Environment management
    direnv             # Environment variable manager
    
    # Shell and terminal
    tmux               # Terminal multiplexer
    starship           # Shell prompt
    
    # ===== BUILD TOOLS =====
    gnumake            # Universal build tool
    just               # Modern command runner
    
    # ===== SECURITY & QUALITY TOOLS =====
    trivy              # Container/code security scanner
    semgrep            # Static code analysis
    
    # ===== GIT WORKFLOW TOOLS =====
    lazygit            # Superior git TUI
    
    # ===== PERFORMANCE TOOLS =====
    hyperfine          # Command-line benchmarking
    wrk                # HTTP load testing
    
    # ===== CONTAINER MONITORING =====
    ctop               # Container monitoring TUI
    
    # Archive tools
    p7zip              # 7-Zip archiver
    
    # File utilities
    rsync              # File synchronization
    
    # ===== OPTIONAL AI/ML TOOLS =====
    # Include Claude Code as requested but not prominently featured
    claude-code        # Claude Code CLI (available but not emphasized)
    
    # ===== DOCUMENTATION AND WRITING =====
    pandoc             # Document converter
    hugo               # Static site generator
    
    # ===== VIRTUALIZATION TOOLS =====
    quickemu           # Quick VM management
    
    # Node.js global packages
    nodePackages.typescript
    nodePackages.eslint
    nodePackages.nodemon
    nodePackages.pm2
    nodePackages.serve
    
    # Tailwind/CSS tools  
    nodePackages.tailwindcss      # Tailwind CSS
    nodePackages.postcss         # CSS processor
    nodePackages.autoprefixer    # CSS vendor prefixes
  ];

  # Dynamic linking support for Claude Code and other native binaries
  programs.nix-ld = {
    enable = true;
    libraries = with pkgs; [
      stdenv.cc.cc
      openssl
      curl
      glib
      zlib
      libz
      xz
      bzip2
      ncurses
      
      # Additional development libraries
      libxml2
      libxslt
      readline
      sqlite
      libnl
      libseccomp
      systemd
      util-linux
    ];
  };
}