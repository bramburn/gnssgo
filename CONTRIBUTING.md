# Contributing to GNSSGO

Thank you for your interest in contributing to GNSSGO! This document provides guidelines and instructions for contributing to this project.

## Code of Conduct

Please be respectful and considerate of others when contributing to this project. We aim to foster an inclusive and welcoming community.

## How to Contribute

### Reporting Bugs

If you find a bug in the code:

1. Check if the bug has already been reported in the [Issues](https://github.com/bramburn/gnssgo/issues) section.
2. If not, create a new issue with a clear description of the bug, including:
   - Steps to reproduce the issue
   - Expected behavior
   - Actual behavior
   - Any relevant logs or error messages
   - Your environment (Go version, OS, etc.)

### Suggesting Enhancements

If you have an idea for an enhancement:

1. Check if the enhancement has already been suggested in the [Issues](https://github.com/bramburn/gnssgo/issues) section.
2. If not, create a new issue with a clear description of your enhancement idea, including:
   - The problem it solves
   - How it should work
   - Any relevant examples or references

### Pull Requests

1. Fork the repository
2. Create a new branch for your changes
3. Make your changes
4. Run tests to ensure your changes don't break existing functionality
5. Submit a pull request with a clear description of your changes

#### Pull Request Guidelines

- Follow the existing code style and conventions
- Include tests for new functionality
- Update documentation as needed
- Keep pull requests focused on a single change
- Reference any related issues in your pull request description

## Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/bramburn/gnssgo.git
   cd gnssgo
   ```

2. Build the project:
   ```bash
   # Build the main library
   cd src
   go build .
   cd ..

   # Build an example application
   cd app/convbin
   go build .
   cd ../..
   ```

3. Run tests:
   ```bash
   # Run tests for the main library
   cd src
   go test .
   cd ..

   # Run unit tests (requires uncommenting unittest in go.work)
   # Edit go.work and uncomment the ./unittest line
   cd unittest
   go test .
   cd ..
   ```

## Coding Standards

- Follow standard Go coding conventions
- Use meaningful variable and function names
- Write clear comments for complex code
- Document public functions and types

## Testing

- Write tests for new functionality
- Ensure all tests pass before submitting a pull request
- Consider edge cases in your tests

## Documentation

- Update documentation for any changes to public APIs
- Add examples for new functionality
- Keep documentation clear and concise

## License

By contributing to GNSSGO, you agree that your contributions will be licensed under the project's license.
