#!/bin/bash

set -e

# Create test directories
TMP_DIR=$(mktemp -d)
SRC_DIR="${TMP_DIR}/src"
DST_DIR="${TMP_DIR}/dst"

mkdir -p "${SRC_DIR}"
mkdir -p "${DST_DIR}"

# Cleanup on exit
trap 'rm -rf ${TMP_DIR}' EXIT

echo "Creating test provider files..."

# Create test provider files
echo "test binary content" > "${SRC_DIR}/terraform-provider-test_v1.0.0_linux_amd64"
chmod +x "${SRC_DIR}/terraform-provider-test_v1.0.0_linux_amd64"

echo "test zip content" > "${SRC_DIR}/terraform-provider-example_v2.0.0_darwin_arm64.zip"

mkdir -p "${SRC_DIR}/nested/dir"
echo "nested provider content" > "${SRC_DIR}/nested/dir/terraform-provider-nested_v3.0.0_windows_386"
chmod +x "${SRC_DIR}/nested/dir/terraform-provider-nested_v3.0.0_windows_386"

# Run the builder
echo "Running terraform-registry-builder..."
go run main.go "${SRC_DIR}" "${DST_DIR}"

# Verify output structure
echo -e "\nVerifying output structure..."

# Function to check if a file exists
check_file() {
    if [ -f "$1" ]; then
        echo "✓ $1"
    else
        echo "✗ $1 (not found)"
        exit 1
    fi
}

# Check for expected files
echo "Checking test provider files..."
check_file "${DST_DIR}/test/versions/index.json"
check_file "${DST_DIR}/test/1.0.0/download/linux/amd64/index.json"
check_file "${DST_DIR}/test/1.0.0/download/linux/amd64/terraform-provider-test_v1.0.0_linux_amd64.zip"
check_file "${DST_DIR}/test/1.0.0/download/linux/amd64/terraform-provider-test_v1.0.0_linux_amd64.zip_SHA256SUMS"
check_file "${DST_DIR}/test/1.0.0/download/linux/amd64/terraform-provider-test_v1.0.0_linux_amd64.zip_SHA256SUMS.sig"

echo "Checking example provider files..."
check_file "${DST_DIR}/example/versions/index.json"
check_file "${DST_DIR}/example/2.0.0/download/darwin/arm64/index.json"
check_file "${DST_DIR}/example/2.0.0/download/darwin/arm64/terraform-provider-example_v2.0.0_darwin_arm64.zip"
check_file "${DST_DIR}/example/2.0.0/download/darwin/arm64/terraform-provider-example_v2.0.0_darwin_arm64.zip_SHA256SUMS"
check_file "${DST_DIR}/example/2.0.0/download/darwin/arm64/terraform-provider-example_v2.0.0_darwin_arm64.zip_SHA256SUMS.sig"

echo "Checking nested provider files..."
check_file "${DST_DIR}/nested/versions/index.json"
check_file "${DST_DIR}/nested/3.0.0/download/windows/386/index.json"
check_file "${DST_DIR}/nested/3.0.0/download/windows/386/terraform-provider-nested_v3.0.0_windows_386.zip"
check_file "${DST_DIR}/nested/3.0.0/download/windows/386/terraform-provider-nested_v3.0.0_windows_386.zip_SHA256SUMS"
check_file "${DST_DIR}/nested/3.0.0/download/windows/386/terraform-provider-nested_v3.0.0_windows_386.zip_SHA256SUMS.sig"

echo -e "\nAll tests passed! ✓"
