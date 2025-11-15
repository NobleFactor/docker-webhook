# Contributor Guide

Welcome to the docker-webhook contributor guide! This page outlines how to set up your development environment, follow the development workflow, and participate in the pull request process.

## Environment Setup

### Prerequisites

Before contributing, ensure you have the following installed:

- **Git**: Version control system
- **Docker**: For building and testing containers
- **Docker Compose**: For multi-container setups
- **Make**: Build automation tool
- **Bash**: Shell scripting (macOS/Linux default)
- **OpenSSL**: For JWT operations
- **GitHub CLI (gh)**: For PR management
- **Azure CLI (az)**: For Key Vault integration (optional, for full testing)

### Installation Steps

1. **Clone the repository**:

   ```bash
   git clone https://github.com/NobleFactor/docker-webhook.git
   cd docker-webhook
   ```

2. **Set up Git hooks** (optional, for pre-commit checks):

   ```bash
   # If hooks exist in .githooks
   git config core.hooksPath .githooks
   ```

3. **Verify Docker installation**:

   ```bash
   docker --version
   docker compose version
   ```

4. **Test basic functionality**:

   ```bash
   make test
   ```

### Development Environment

The project uses a Makefile-based workflow. Key directories:

- `bin/`: Scripts like `New-JwtToken` and `Declare-BashScript`
- `webhook.config/`: Configuration files for deployments
- `docs/`: Documentation and man pages
- `test/`: Test scripts
- `.github/`: CI workflows and Copilot instructions

For local development, you can build images and run tests without affecting production systems.

## Development Workflow

### 1. Choose an Issue

- Check the [Issues](https://github.com/NobleFactor/docker-webhook/issues) page for open tasks
- Comment on an issue to indicate you're working on it
- For new features, create an issue first to discuss the approach

### 2. Create a Feature Branch

Always work on a feature branch, never directly on `develop` or `master`.

```bash
# Switch to develop and pull latest
git checkout develop
git pull origin develop

# Create feature branch
git checkout -b feature/your-feature-name
```

Branch naming conventions:

- `feature/description`: New features
- `fix/description`: Bug fixes
- `docs/description`: Documentation updates
- `refactor/description`: Code refactoring

### 3. Make Changes

- Follow the [Copilot Instructions](.github/copilot-instructions.md) for coding standards
- Use the Makefile for building and testing:

  ```bash
  # Build image
  make New-WebhookImage
  
  # Run tests
  make test
  
  # Format code
  make format
  ```

- For webhook configuration changes:
  - Edit `webhook.config/$(LOCATION)/hooks.json`
  - Test with `make New-WebhookLocation`

- For script changes:
  - Test with `make test`
  - Ensure shellcheck passes

### 4. Testing

- Run the full test suite: `make test`
- Test Docker builds: `make New-WebhookImage`
- For JWT features: Test `bin/New-JwtToken` with various options
- Manual testing: Use `make Start-Webhook` for integration tests

### 5. Commit Changes

- Write clear, descriptive commit messages
- Follow conventional commit format: `type: description`
- Keep commits focused on single changes

```bash
git add <files>
git commit -m "feat: add JWT token validation

- Implement HMAC signature verification
- Add error handling for invalid tokens
- Update tests"
```

### 6. Push and Create PR

- Push your branch: `git push -u origin feature/your-feature-name`
- Create a PR using GitHub CLI or web interface

## Pull Request Process

### Creating a PR

1. **Use GitHub CLI** (recommended):

   ```bash
   gh pr create --title "Feature: Your Feature Description" --body "Detailed description of changes..."
   ```

2. **Or use the web interface**: Click "New Pull Request" on GitHub

### PR Requirements

- **Title**: Clear and descriptive (e.g., "feat: Add JWT authentication support")
- **Description**:
  - What changes were made
  - Why they were necessary
  - How to test the changes
  - Screenshots/videos if UI changes
- **Base branch**: `develop` (not `master`)
- **Labels**: Add appropriate labels (enhancement, bug, documentation, etc.)

### PR Checklist

Before submitting:

- [ ] Code follows project conventions
- [ ] Tests pass: `make test`
- [ ] Documentation updated if needed
- [ ] Commit messages are clear
- [ ] No sensitive data committed
- [ ] Branch is up to date with `develop`

### Review Process

1. **Automated Checks**: CI will run tests and linting
2. **Code Review**: At least one maintainer review required
3. **Feedback**: Address review comments with additional commits
4. **Approval**: Maintainers will approve when ready
5. **Merge**: Squash merge preferred, or rebase if requested

### After Merge

- Delete your feature branch
- Pull changes to local develop branch
- Close related issues if applicable

## Additional Resources

- [README.md](README.md): Project overview and deployment guide
- [Makefile](makefile): Build and deployment commands
- [Issues](https://github.com/NobleFactor/docker-webhook/issues): Bug reports and feature requests
- [Discussions](https://github.com/NobleFactor/docker-webhook/discussions): General questions and ideas

## Getting Help

- Open an issue for bugs or feature requests
- Use Discussions for questions
- Check existing issues/PRs before creating new ones

Thank you for contributing to docker-webhook! Your help makes the project better for everyone.
