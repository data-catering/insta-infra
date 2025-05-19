#!/bin/bash
set -e # Exit immediately if a command exits with a non-zero status.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Install Python dependencies from requirements.txt
echo "Installing Python dependencies from requirements.txt..."
pip install -r "${SCRIPT_DIR}/requirements.txt" --quiet

# Run Text Classification loader
echo "Running Text Classification data loader..."
python "${SCRIPT_DIR}/load_text_classification_data.py"

# Run Token Classification loader
echo "Running Token Classification data loader..."
python "${SCRIPT_DIR}/load_token_classification_data.py"

# Run Image Classification loader
echo "Running Image Classification data loader..."
python "${SCRIPT_DIR}/load_image_classification_data.py"

# Run Image Preference loader (currently does not generate suggestions)
echo "Running Image Preference data loader..."
python "${SCRIPT_DIR}/load_image_preference_data.py"

echo "All Argilla data loading scripts completed." 