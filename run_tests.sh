#!/bin/bash
set -e

# Setup Verification Environment for x402-a2a MultiversX Integration
echo "🚀 Setting up verification environment..."

# 1. Create venv
if [ ! -d "venv_test" ]; then
    python3 -m venv venv_test
fi
source venv_test/bin/activate

# Upgrade pip
pip install --upgrade pip

# 2. Install dependencies (REAL)
echo "📦 Installing dependencies (Editable Mde)..."
# Install core package first (Modified for Py3.9)
pip install -e x402_repo/python/x402

# Install a2a-sdk (local AP2, modified for Py3.9)
pip install -e google-agentic-commerce/AP2

# Install a2a package with extras
cd google-agentic-commerce/a2a-x402/python/x402_a2a
pip install -e ".[multiversx]"
cd -

# 3. Run Integration Tests
echo "🧪 Running Integration Tests (No Mocks)..."
export PYTHONPATH=$PYTHONPATH:$(pwd)/google-agentic-commerce/a2a-x402/python/x402_a2a/src:$(pwd)/x402_repo/python/x402/src

python3 -m unittest google-agentic-commerce/a2a-x402/python/x402_a2a/tests/test_integration_real.py

echo "✅ Integration Verification Complete!"
