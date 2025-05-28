#!/bin/bash

# Script to sync Postman collection with OpenAPI documentation
# This script uses a custom Node.js converter to convert OpenAPI specs to Postman collections

# Function to install Node.js
install_nodejs() {
    echo "Node.js is not installed. Attempting to install..."
    
    # Check the operating system
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        if command -v brew &> /dev/null; then
            echo "Installing Node.js via Homebrew..."
            brew install node
        else
            echo "Homebrew not found. Installing Homebrew first..."
            /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
            echo "Installing Node.js via Homebrew..."
            brew install node
        fi
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        if command -v apt-get &> /dev/null; then
            echo "Installing Node.js via apt..."
            curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
            sudo apt-get install -y nodejs
        elif command -v yum &> /dev/null; then
            echo "Installing Node.js via yum..."
            curl -fsSL https://rpm.nodesource.com/setup_18.x | sudo bash -
            sudo yum install -y nodejs
        else
            echo "Could not determine package manager. Please install Node.js manually."
            echo "Visit https://nodejs.org/ to download and install Node.js"
            exit 1
        fi
    else
        echo "Unsupported operating system. Please install Node.js manually."
        echo "Visit https://nodejs.org/ to download and install Node.js"
        exit 1
    fi
    
    # Verify installation
    if ! command -v node &> /dev/null; then
        echo "Failed to install Node.js. Please install it manually."
        echo "Visit https://nodejs.org/ to download and install Node.js"
        exit 1
    fi
    
    echo "Node.js installed successfully."
}

# Function to install jq
install_jq() {
    echo "jq is not installed. Attempting to install..."
    
    # Check the operating system
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        if command -v brew &> /dev/null; then
            echo "Installing jq via Homebrew..."
            brew install jq
        else
            echo "Homebrew not found. Installing Homebrew first..."
            /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
            echo "Installing jq via Homebrew..."
            brew install jq
        fi
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        if command -v apt-get &> /dev/null; then
            echo "Installing jq via apt..."
            sudo apt-get update
            sudo apt-get install -y jq
        elif command -v yum &> /dev/null; then
            echo "Installing jq via yum..."
            sudo yum install -y jq
        else
            echo "Could not determine package manager. Please install jq manually."
            echo "For installation instructions, visit: https://stedolan.github.io/jq/download/"
            exit 1
        fi
    else
        echo "Unsupported operating system. Please install jq manually."
        echo "For installation instructions, visit: https://stedolan.github.io/jq/download/"
        exit 1
    fi
    
    # Verify installation
    if ! command -v jq &> /dev/null; then
        echo "Failed to install jq. Please install it manually."
        echo "For installation instructions, visit: https://stedolan.github.io/jq/download/"
        exit 1
    fi
    
    echo "jq installed successfully."
}

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    install_nodejs
fi

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    install_jq
fi

# Define paths
PLUGIN_ROOT=$(pwd)
POSTMAN_DIR="${PLUGIN_ROOT}/postman"
TEMP_DIR="${PLUGIN_ROOT}/postman/temp"
PLUGIN_API="${PLUGIN_ROOT}/api"
POSTMAN_COLLECTION="${POSTMAN_DIR}/plugin.postman_collection.json"
BACKUP_DIR="${POSTMAN_DIR}/backups"
CONVERTER_SCRIPT="${PLUGIN_ROOT}/scripts/convert-openapi.js"

# Create necessary directories
mkdir -p "${TEMP_DIR}"
mkdir -p "${BACKUP_DIR}"

# Backup existing Postman collection
TIMESTAMP=$(date +"%Y%m%d%H%M%S")
BACKUP_FILE="${BACKUP_DIR}/plugin.postman_collection.${TIMESTAMP}.json"
if [ -f "${POSTMAN_COLLECTION}" ]; then
    echo "Backing up existing Postman collection to ${BACKUP_FILE}..."
    cp "${POSTMAN_COLLECTION}" "${BACKUP_FILE}"
fi

# Convert OpenAPI specs to Postman collections
echo "Converting OpenAPI specs to Postman collections..."

# Install js-yaml dependency if not already installed
SCRIPTS_DIR="${PLUGIN_ROOT}/scripts"
if [ ! -d "${SCRIPTS_DIR}/node_modules/js-yaml" ]; then
    echo "Installing js-yaml dependency in scripts directory..."
    (cd "${SCRIPTS_DIR}" && npm install js-yaml)
    if [ $? -ne 0 ]; then
        echo "Failed to install js-yaml dependency. Please install it manually with 'cd ${SCRIPTS_DIR} && npm install js-yaml'."
        exit 1
    fi
fi

# Process PLUGIN component - prefer OpenAPI YAML over Swagger JSON
if [ -f "${PLUGIN_API}/openapi.yaml" ]; then
    echo "Processing plugin using OpenAPI YAML..."
    node "${CONVERTER_SCRIPT}" "${PLUGIN_API}/openapi.yaml" "${TEMP_DIR}/plugin.postman_collection.json"
    if [ $? -ne 0 ]; then
        echo "Failed to convert plugin API spec to Postman collection."
    else
        # Move the generated collection to the final destination
        echo "Moving generated collection to ${POSTMAN_COLLECTION}..."
        cp "${TEMP_DIR}/plugin.postman_collection.json" "${POSTMAN_COLLECTION}"
        if [ $? -ne 0 ]; then
            echo "Failed to move generated collection to final destination."
        fi
    fi
elif [ -f "${PLUGIN_API}/swagger.json" ]; then
    echo "OpenAPI YAML not found. Processing plugin using Swagger JSON..."
    node "${CONVERTER_SCRIPT}" "${PLUGIN_API}/swagger.json" "${TEMP_DIR}/plugin.postman_collection.json"
    if [ $? -ne 0 ]; then
        echo "Failed to convert plugin API spec to Postman collection."
    else
        # Move the generated collection to the final destination
        echo "Moving generated collection to ${POSTMAN_COLLECTION}..."
        cp "${TEMP_DIR}/plugin.postman_collection.json" "${POSTMAN_COLLECTION}"
        if [ $? -ne 0 ]; then
            echo "Failed to move generated collection to final destination."
        fi
    fi
else
    echo "Neither OpenAPI YAML nor Swagger JSON found. Skipping..."
fi

# Clean up temporary files
echo "Cleaning up temporary files..."
rm -rf "${TEMP_DIR}"

echo "[ok] Postman collection synced successfully with OpenAPI documentation ✔️"
echo "Note: The synced collection is available at ${POSTMAN_COLLECTION}"
echo "A backup of the previous collection is available at ${BACKUP_FILE}"
