#!/bin/bash

set -e

echo "🚀 Generating TypeScript API clients from OpenAPI specifications..."

# Check if required tools are installed
if ! command -v npx &> /dev/null; then
    echo "❌ npx is required but not installed. Please install Node.js."
    exit 1
fi

# Create directories for generated code
mkdir -p generated/public
mkdir -p generated/admin

echo "📋 Validating OpenAPI specifications..."
# Validate specifications
npx swagger-codegen-cli validate -i ../../openapi/public-api.yaml
npx swagger-codegen-cli validate -i ../../openapi/admin-api.yaml

echo "🔧 Generating Public API client..."
# Generate Public API client
npx openapi-generator-cli generate \
    -i ../../openapi/public-api.yaml \
    -g typescript-fetch \
    -o ./generated/public \
    --additional-properties=typescriptThreePlus=true,supportsES6=true,npmName=@international-center/public-api-client,npmVersion=1.0.0,withInterfaces=true

echo "🔧 Generating Admin API client..."
# Generate Admin API client
npx openapi-generator-cli generate \
    -i ../../openapi/admin-api.yaml \
    -g typescript-fetch \
    -o ./generated/admin \
    --additional-properties=typescriptThreePlus=true,supportsES6=true,npmName=@international-center/admin-api-client,npmVersion=1.0.0,withInterfaces=true

echo "🏗️  Building TypeScript project..."
# Install dependencies and build
npm install
npm run build

echo "✅ TypeScript API clients generated successfully!"
echo "📁 Generated files:"
echo "   - generated/public/ - Public API client"
echo "   - generated/admin/ - Admin API client"
echo "   - dist/ - Built TypeScript files"

echo "🎉 Generation complete! You can now use the API clients in your frontend applications."