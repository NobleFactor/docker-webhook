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

3. **Bootstrap the repository (recommended)**

  Run the single setup command to install dependencies and register pre-commit hooks:

  ```bash
  make setup
  ```

  This will install required tools (shfmt, shellcheck, pre-commit, etc.) and configure pre-commit to run `make Test-ShellFormatting` and `make Test-ShellScripts` on commits.

4. **Verify Docker installation**:

   ```bash
   docker --version
   docker compose version
   ```

5. **Test basic functionality**:

   ```bash
   make test
   ```

### Development Environment

The project uses a Makefile-based workflow. Key directories:

- `bin/`: Scripts like `New-JsonWebToken` and `Declare-BashScript`
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
  - Test with `make Prepare-WebhookDeployment`
  Example: `Prepare-WebhookDeployment --env-file ./examples/us-wa.env`

- For script changes:
  - Test with `make test`
  - Ensure shellcheck passes (see `docs/Shell-script-style.md`)

### 4. Testing

- Run the full test suite: `make test`
- Test Docker builds: `make New-WebhookImage`
- For JWT features: Test `bin/New-JsonWebToken` with various options
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

## Contributing as an External Contributor

If you don't have write access to this repository, follow these steps to contribute:

### 1. Fork the Repository

- Click the "Fork" button on the repository's GitHub page
- This creates a copy of the repository under your GitHub account

### 2. Clone Your Fork

```bash
git clone https://github.com/YOUR_USERNAME/docker-webhook.git
cd docker-webhook
```

### 3. Set Up the Upstream Remote

```bash
git remote add upstream https://github.com/NobleFactor/docker-webhook.git
git fetch upstream
```

### 4. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
```

### 5. Make Your Changes

- Follow the development workflow in this guide
- Test your changes thoroughly
- Ensure code follows project conventions

### 6. Keep Your Branch Updated

```bash
git fetch upstream
git rebase upstream/develop
```

### 7. Push to Your Fork

```bash
git push origin feature/your-feature-name
```

### 8. Create a Pull Request

- Go to the original repository on GitHub
- Click "New Pull Request"
- Select your fork and branch as the source
- Fill in the PR template with:
  - Clear title and description
  - What changes were made and why
  - How to test the changes
  - Any relevant screenshots or context

### 9. Address Review Feedback

- Respond to reviewer comments
- Make additional commits to your branch if needed
- The PR will be merged by maintainers once approved

### 10. Clean Up

- After merge, delete your feature branch
- Keep your fork updated with the main repository

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
