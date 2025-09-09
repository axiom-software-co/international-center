#!/usr/bin/env bash

set -e

echo "📦 Installing generated TypeScript clients to frontend applications..."

# Check if generated clients exist
if [ ! -d "./generated" ]; then
    echo "❌ Generated clients not found. Run 'pnpm run generate' first."
    exit 1
fi

# Install to admin portal
if [ -d "../../frontend/admin-portal" ]; then
    echo "🔧 Installing admin client to admin portal..."
    
    # Create node_modules directories for local packages
    mkdir -p ../../frontend/admin-portal/node_modules/@international-center
    
    # Copy admin client
    cp -r ./generated/admin ../../frontend/admin-portal/node_modules/@international-center/admin-api-client
    
    # Copy public client (admin portal might need public API too)
    cp -r ./generated/public ../../frontend/admin-portal/node_modules/@international-center/public-api-client
    
    echo "✅ Admin portal clients installed"
fi

# Install to public website
if [ -d "../../frontend/public-website" ]; then
    echo "🔧 Installing public client to public website..."
    
    # Create node_modules directories for local packages
    mkdir -p ../../frontend/public-website/node_modules/@international-center
    
    # Copy public client
    cp -r ./generated/public ../../frontend/public-website/node_modules/@international-center/public-api-client
    
    echo "✅ Public website client installed"
fi

echo "🎉 TypeScript clients installed successfully!"
echo "📁 Clients available as:"
echo "   - @international-center/admin-api-client (admin portal)"
echo "   - @international-center/public-api-client (both frontends)"