# Contributing to Kairo

First off, thank you for considering contributing to Kairo! It's people like you that make Kairo such a great tool.

## Code of Conduct

This project and everyone participating in it is governed by the [Kairo Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

This section guides you through submitting a bug report for Kairo. Following these guidelines helps maintainers and the community understand your report, reproduce the behavior, and find related reports.

- **Check if the bug has already been reported.**
- **Use a clear and descriptive title.**
- **Describe the exact steps which reproduce the problem.**
- **Explain which behavior you expected to see and why.**
- **Include screenshots or animated GIFs if possible.**

### Suggesting Enhancements

This section guides you through submitting an enhancement suggestion for Kairo, including completely new features and minor improvements to existing functionality.

- **Check if there's already a suggestion for the enhancement.**
- **Use a clear and descriptive title.**
- **Provide a step-by-step description of the suggested enhancement in as many details as possible.**
- **Explain why this enhancement would be useful to most Kairo users.**

### Pull Requests

- **Fill in the pull request template.**
- **Include screenshots and animated GIFs in your pull request description if they help.**
- **Ensure that your code follows the existing style.**
- **Add tests for new features and bug fixes.**
- **Ensure that all tests pass before submitting.**

## Development Setup

1. Fork the repository and clone it locally.
2. Install Go (1.26+).
3. Install dependencies: `go mod download`.
4. Run tests: `go test ./...`.
5. Build the project: `go build ./cmd/kairo`.

## Style Guide

We follow the standard Go coding style. Please run `go fmt ./...` before committing.
