# Programs Configuration Module
# Handles shell, git, starship, and development environment configuration

{ config, pkgs, lib, ... }:

let
  # Reference to the pythonEnv from packages module
  pythonEnv = pkgs.python312.withPackages (ps: with ps; [
    # Web frameworks and API development
    fastapi flask django uvicorn gunicorn
    # Data science and analysis
    pandas numpy matplotlib seaborn jupyter
    # Database connectivity
    sqlalchemy alembic psycopg2 redis asyncpg
    # HTTP and API clients
    httpx aiohttp requests
    # Development and testing
    pytest pytest-asyncio black isort mypy flake8
    # Configuration and utilities
    python-dotenv pydantic pydantic-settings typer rich loguru
    # Optional AI/ML libraries
    openai anthropic langchain langchain-community
  ]);
in

{
  # ===== BASH CONFIGURATION =====
  programs.bash = {
    enableCompletion = true;
    
    # ===== DEVELOPMENT ALIASES =====
    shellAliases = {
      # ===== MODERN CLI REPLACEMENTS =====
      ls = "eza --icons --group-directories-first";
      ll = "eza -la --icons --group-directories-first --git";
      la = "eza -la --icons --group-directories-first";
      cat = "bat --style=auto";
      grep = "rg";
      find = "fd";
      cd = "z";  # zoxide enhanced cd
      
      # ===== GIT SHORTCUTS =====
      g = "git";
      gs = "git status";
      ga = "git add";
      gc = "git commit";
      gp = "git push";
      gl = "git log --oneline --graph";
      gd = "git diff";
      gb = "git branch";
      gco = "git checkout";
      
      # ===== NODE.JS SHORTCUTS =====
      ni = "npm install";
      nr = "npm run";
      ns = "npm start";
      nt = "npm test";
      nb = "npm run build";
      nd = "npm run dev";
      
      # Yarn shortcuts
      yi = "yarn install";
      yr = "yarn run";
      ys = "yarn start";
      yt = "yarn test";
      yb = "yarn build";
      yd = "yarn dev";
      
      # PNPM shortcuts
      pi = "pnpm install";
      pr = "pnpm run";
      ps = "pnpm start";
      pt = "pnpm test";
      pb = "pnpm build";
      pd = "pnpm dev";
      
      # ===== DOCKER & KUBERNETES =====
      dk = "docker";
      dkc = "docker-compose";
      dcp = "docker-compose";
      k = "kubectl";
      
      # ===== PYTHON SHORTCUTS =====
      py = "python";
      python = "python3";
      pip = "pip3";
      
      # ===== CLOUD TOOLS =====
      tf = "terraform";
      tg = "terragrunt";
      aws = "aws";
      gcloud = "gcloud";
      
      # ===== SYSTEM MAINTENANCE =====
      rebuild = "sudo nixos-rebuild switch";
      cleanup = "sudo nix-collect-garbage -d && nix-collect-garbage -d";
    };
    
    # ===== BASH INITIALIZATION =====
    interactiveShellInit = ''
      # Initialize modern tools
      command -v zoxide >/dev/null && eval "$(zoxide init bash)"
      command -v starship >/dev/null && eval "$(starship init bash)"
      command -v direnv >/dev/null && eval "$(direnv hook bash)"
      
      # Set development environment indicator
      export DEV_ENV="dev-unstable"
    '';
  };

  # ===== GIT CONFIGURATION =====
  programs.git = {
    enable = true;
    
    config = {
      user = {
        name = "tojkuv";
        email = "dev@example.com";
      };
      
      init.defaultBranch = "main";
      core.editor = "nvim";
      pull.rebase = true;
      push.autoSetupRemote = true;
    };
  };

  # ===== SHELL PROMPT =====
  programs.starship = {
    enable = true;
    settings = {
      format = "[dev-unstable](bold blue) $directory$git_branch$git_status$nodejs$python$golang$rust$java$dotnet$character";
      character = {
        success_symbol = "[→](bold green)";
        error_symbol = "[→](bold red)";
      };
      directory = {
        style = "cyan";
        truncation_length = 3;
        truncate_to_repo = true;
      };
      git_branch = {
        format = "[$symbol$branch]($style) ";
        symbol = " ";
        style = "purple";
      };
      git_status = {
        format = "[$all_status$ahead_behind]($style) ";
        style = "red";
      };
      # Language indicators (clean, minimal)
      nodejs = {
        format = "[node:$version]($style) ";
        style = "green";
        detect_files = ["package.json" ".nvmrc"];
      };
      python = {
        format = "[py:$version]($style) ";
        style = "yellow";
      };
      golang = {
        format = "[go:$version]($style) ";
        style = "cyan";
      };
      rust = {
        format = "[rust:$version]($style) ";
        style = "red";
      };
      java = {
        format = "[java:$version]($style) ";
        style = "orange";
      };
      dotnet = {
        format = "[.net:$version]($style) ";
        style = "blue";
      };
    };
  };

  # ===== DIRECTORY NAVIGATION =====
  programs.zoxide.enable = true;

  # ===== ENVIRONMENT MANAGEMENT =====
  programs.direnv = {
    enable = true;
    nix-direnv.enable = true;
  };

  # ===== ENVIRONMENT VARIABLES =====
  environment.variables = {
    # Core development environment
    EDITOR = "nvim";
    VISUAL = "nvim";
    BROWSER = "firefox";
    PAGER = "bat";
    
    # Programming language configurations
    NODE_OPTIONS = "--max-old-space-size=8192";
    NODE_ENV = "development";
    
    # Python environment
    PYTHONPATH = "${pythonEnv}/${pythonEnv.sitePackages}";
    PYTHONUNBUFFERED = "1";
    
    # Go development
    GOPATH = "$HOME/go";
    GO111MODULE = "on";
    
    # Rust development
    CARGO_HOME = "$HOME/.cargo";
    
    # Container environment
    DOCKER_BUILDKIT = "1";
    COMPOSE_DOCKER_CLI_BUILD = "1";
    BUILDKIT_PROGRESS = "plain";
    
    # Development paths
    JAVA_HOME = "${pkgs.jdk21}";
    DOTNET_ROOT = "${pkgs.dotnet-sdk_9}";
    
    # Cloud SDK configurations
    AWS_PAGER = "";
    AZURE_CORE_OUTPUT = "table";
    
    # Set development environment indicator
    DEV_ENV = "dev-unstable";
  };
  
  # Additional PATH entries
  environment.extraInit = ''
    export PATH="$HOME/.local/bin:$HOME/.cargo/bin:$HOME/.npm-global/bin:$HOME/go/bin:$HOME/.dotnet/tools:$PATH"
  '';

  # ===== DEVELOPMENT DIRECTORIES =====
  # Create git message template
  environment.etc."gitmessage".text = ''
    # Professional Development Commit
    # Type: feat|fix|docs|style|refactor|test|chore
  '';
  
  # Set global git message template
  programs.git.config.commit.template = "/etc/gitmessage";

  # Global git ignore patterns
  programs.git.config.core.excludesFile = pkgs.writeText "gitignore_global" ''
    # Dependencies
    node_modules/
    vendor/
    __pycache__/
    *.pyc
    target/
    bin/
    obj/
    
    # Environment and secrets
    .env
    .env.local
    .env.production
    *.key
    *.pem
    
    # Build outputs
    dist/
    build/
    out/
    coverage/
    
    # IDE and editor files
    .vscode/
    .idea/
    *.swp
    *.swo
    *~
    
    # OS files
    .DS_Store
    Thumbs.db
    
    # Logs
    *.log
    logs/
    
    # Development tools
    .direnv/
    result
    .pytest_cache/
    .coverage
    .nyc_output
  '';
}