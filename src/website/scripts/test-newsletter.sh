#!/bin/bash

# Newsletter Testing Script
# Ensures newsletter service is running before executing tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NEWSLETTER_API_URL="http://localhost:8086"
MAX_WAIT_ATTEMPTS=30
WAIT_INTERVAL=2

echo -e "${BLUE}🧪 Newsletter Testing Script${NC}"
echo -e "${BLUE}=============================${NC}\n"

# Function to check if newsletter service is running
check_newsletter_service() {
    echo -e "${YELLOW}📡 Checking newsletter service availability...${NC}"
    
    local attempts=0
    while [ $attempts -lt $MAX_WAIT_ATTEMPTS ]; do
        if curl -s -o /dev/null -w "%{http_code}" "${NEWSLETTER_API_URL}/api/newsletter/confirm/test" | grep -q "400"; then
            echo -e "${GREEN}✅ Newsletter service is running and responding${NC}"
            return 0
        fi
        
        attempts=$((attempts + 1))
        echo -e "${YELLOW}⏳ Waiting for newsletter service... (attempt $attempts/$MAX_WAIT_ATTEMPTS)${NC}"
        sleep $WAIT_INTERVAL
    done
    
    echo -e "${RED}❌ Newsletter service is not available after ${MAX_WAIT_ATTEMPTS} attempts${NC}"
    echo -e "${RED}   Please ensure the newsletter service is running:${NC}"
    echo -e "${RED}   podman-compose up newsletter-domain${NC}"
    return 1
}

# Function to run integration tests
run_integration_tests() {
    echo -e "\n${BLUE}🧪 Running Newsletter Integration Tests...${NC}"
    echo -e "${BLUE}=========================================${NC}"
    
    if pnpm run test:newsletter; then
        echo -e "\n${GREEN}✅ Integration tests passed${NC}"
        return 0
    else
        echo -e "\n${RED}❌ Integration tests failed${NC}"
        return 1
    fi
}

# Function to run E2E tests
run_e2e_tests() {
    echo -e "\n${BLUE}🎭 Running Newsletter E2E Tests...${NC}"
    echo -e "${BLUE}==================================${NC}"
    
    local headed=""
    if [[ "$1" == "--headed" ]]; then
        headed="--headed"
        echo -e "${YELLOW}🖥️  Running tests in headed mode${NC}"
    fi
    
    if pnpm run test:newsletter:e2e $headed; then
        echo -e "\n${GREEN}✅ E2E tests passed${NC}"
        return 0
    else
        echo -e "\n${RED}❌ E2E tests failed${NC}"
        return 1
    fi
}

# Function to check if all required services are running
check_all_services() {
    echo -e "${YELLOW}🔍 Checking all required services...${NC}"
    
    # Check if website is running (for E2E tests)
    if curl -s -o /dev/null "http://localhost:4321"; then
        echo -e "${GREEN}✅ Website is running${NC}"
    else
        echo -e "${YELLOW}⚠️  Website not detected (required for E2E tests)${NC}"
        echo -e "${YELLOW}   Start with: pnpm run dev${NC}"
    fi
    
    # Check newsletter service
    check_newsletter_service || return 1
    
    echo -e "${GREEN}✅ All services are ready${NC}\n"
}

# Function to display help
show_help() {
    echo -e "${BLUE}Newsletter Testing Script Usage:${NC}"
    echo -e ""
    echo -e "  $0                      # Run both integration and E2E tests"
    echo -e "  $0 --integration-only   # Run only integration tests"
    echo -e "  $0 --e2e-only          # Run only E2E tests"
    echo -e "  $0 --e2e-headed        # Run E2E tests in headed mode"
    echo -e "  $0 --help              # Show this help message"
    echo -e ""
    echo -e "${YELLOW}Prerequisites:${NC}"
    echo -e "  - Newsletter service running on localhost:8086"
    echo -e "  - Website running on localhost:4321 (for E2E tests)"
    echo -e "  - All dependencies installed (pnpm install)"
    echo -e ""
    echo -e "${YELLOW}Start services with:${NC}"
    echo -e "  podman-compose up newsletter-domain"
    echo -e "  pnpm run dev"
}

# Main script logic
main() {
    case "${1:-}" in
        --help|-h)
            show_help
            exit 0
            ;;
        --integration-only)
            check_newsletter_service || exit 1
            run_integration_tests || exit 1
            ;;
        --e2e-only)
            check_all_services || exit 1
            run_e2e_tests || exit 1
            ;;
        --e2e-headed)
            check_all_services || exit 1
            run_e2e_tests --headed || exit 1
            ;;
        "")
            # Run both integration and E2E tests
            echo -e "${BLUE}🚀 Running complete newsletter test suite${NC}\n"
            
            # Check services first
            check_all_services || exit 1
            
            # Run integration tests
            if run_integration_tests; then
                echo -e "\n${GREEN}✅ Integration tests completed successfully${NC}"
            else
                echo -e "\n${RED}❌ Integration tests failed, skipping E2E tests${NC}"
                exit 1
            fi
            
            # Run E2E tests
            if run_e2e_tests; then
                echo -e "\n${GREEN}🎉 All newsletter tests passed!${NC}"
            else
                echo -e "\n${RED}❌ E2E tests failed${NC}"
                exit 1
            fi
            ;;
        *)
            echo -e "${RED}❌ Unknown option: $1${NC}"
            show_help
            exit 1
            ;;
    esac
}

# Trap to handle script interruption
trap 'echo -e "\n${YELLOW}⚠️  Test execution interrupted${NC}"; exit 130' INT

# Run main function with all arguments
main "$@"

echo -e "\n${BLUE}🏁 Newsletter testing completed${NC}"