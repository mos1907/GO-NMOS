# Contributing to go-NMOS

Thank you for your interest in contributing to go-NMOS! This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for all contributors.

## How to Contribute

### Reporting Bugs

If you find a bug, please create an issue with:
- **Clear title**: Brief description of the issue
- **Steps to reproduce**: Detailed steps to reproduce the bug
- **Expected behavior**: What you expected to happen
- **Actual behavior**: What actually happened
- **Environment**: OS, Go version, Node.js version, Docker version
- **Screenshots**: If applicable

### Suggesting Features

Feature suggestions are welcome! Please create an issue with:
- **Clear title**: Brief description of the feature
- **Use case**: Why this feature would be useful
- **Proposed solution**: How you envision it working
- **Alternatives**: Other solutions you've considered

### Pull Requests

1. **Fork the repository**
   ```bash
   git clone https://github.com/mos1907/GO-NMOS.git
   cd GO-NMOS
   ```

2. **Create a branch**
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

3. **Make your changes**
   - Follow the coding standards (see below)
   - Add tests if applicable
   - Update documentation as needed

4. **Test your changes**
   ```bash
   # Backend tests
   cd backend
   go test ./...
   
   # Frontend build
   cd ../frontend
   npm run build
   ```

5. **Commit your changes**
   ```bash
   git commit -m "feat: add new feature"  # or "fix:", "docs:", "refactor:", etc.
   ```
   Use [Conventional Commits](https://www.conventionalcommits.org/) format:
   - `feat:` - New feature
   - `fix:` - Bug fix
   - `docs:` - Documentation changes
   - `refactor:` - Code refactoring
   - `test:` - Test additions/changes
   - `chore:` - Build process or auxiliary tool changes

6. **Push to your fork**
   ```bash
   git push origin feature/your-feature-name
   ```

7. **Create a Pull Request**
   - Fill out the PR template
   - Link any related issues
   - Request review from maintainers

## Coding Standards

### Go (Backend)

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Run `go fmt` before committing
- Run `go vet ./...` to check for issues
- Use meaningful variable and function names
- Add comments for exported functions/types
- Keep functions focused and small

### JavaScript/Svelte (Frontend)

- Follow existing code style
- Use meaningful variable names
- Keep components focused and reusable
- Add comments for complex logic
- Run `npm run build` to ensure no build errors

### Git Commit Messages

- Use present tense ("add feature" not "added feature")
- Use imperative mood ("move cursor to..." not "moves cursor to...")
- First line should be concise (50 chars or less)
- Reference issues/PRs: `Closes #123` or `Fixes #456`

## Development Setup

See [README.md](README.md#local-development) for detailed setup instructions.

### Quick Start

```bash
# Clone repository
git clone https://github.com/mos1907/GO-NMOS.git
cd GO-NMOS

# Backend setup
cd backend
go mod tidy
go run ./cmd/api

# Frontend setup (in another terminal)
cd frontend
npm install
npm run dev
```

## Project Structure

- `backend/` - Go backend service
- `frontend/` - Svelte frontend application
- `deploy/` - Deployment configurations
- `.github/` - GitHub workflows and templates

## Testing

### Backend Tests

```bash
cd backend
go test ./...
go test -v ./...  # Verbose output
go test -race ./...  # Race condition detection
```

### Frontend Tests

Currently, frontend tests are not set up. Contributions for adding tests are welcome!

## Documentation

- Update README.md for user-facing changes
- Add code comments for complex logic
- Update API documentation if endpoints change
- Keep CHANGELOG.md updated (if maintained)

## Questions?

- Open an issue for questions or discussions
- Check existing issues/PRs for similar questions
- Be patient - maintainers are volunteers

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
