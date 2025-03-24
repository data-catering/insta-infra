# Homebrew Tap for insta-infra

This is a Homebrew tap repository for [insta-infra](https://github.com/data-catering/insta-infra).

## Setup Instructions

1. Create a new repository named `homebrew-insta-infra` on GitHub
2. Clone the repository:
   ```
   git clone https://github.com/data-catering/homebrew-insta-infra.git
   cd homebrew-insta-infra
   ```
3. Create a Formula directory:
   ```
   mkdir -p Formula
   ```
4. Copy the `insta-infra.rb` formula from the main repository:
   ```
   cp /path/to/insta-infra.rb Formula/
   ```
5. Update the SHA256 hash in the formula with the correct value from the latest release.
6. Commit and push:
   ```
   git add Formula/insta-infra.rb
   git commit -m "Add insta-infra formula"
   git push
   ```
7. Set up a GitHub Actions workflow in this repository to automatically update the formula when a new release is published in the main repository.

## Usage

Once the tap is set up, users can install insta-infra with:

```
brew tap data-catering/insta-infra
brew install insta-infra
``` 