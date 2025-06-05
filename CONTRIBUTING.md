# Contributing to Blockchain Logistics Traceability

We warmly welcome all contributions from the community! This project aims to create a transparent and reliable logistics traceability system.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How to Contribute](#how-to-contribute)
- [Development Process](#development-process)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Pull Request Process](#pull-request-process)
- [Reporting Issues](#reporting-issues)

## Code of Conduct

This project adheres to the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you agree to abide by these rules.

## How to Contribute

### Reporting Bugs
- Use GitHub Issues to report bugs
- Check if the bug has already been reported
- Use the bug report template when available
- Provide detailed information: environment, reproduction steps, expected results

### Suggesting New Features
- Create a GitHub Issue with the "enhancement" label
- Clearly describe the feature and why it's needed
- Discuss with maintainers before implementing

### Code Contributions
1. Fork the repository
2. Create a feature branch from `develop`
3. Implement changes
4. Write/update tests
5. Ensure coding standards are met
6. Submit a pull request

## Development Process

### Branch Strategy
- `main`: Production-ready code
- `develop`: Development branch
- `feature/*`: Feature branches
- `hotfix/*`: Emergency fixes
- `release/*`: Release preparation

### Setup Development Environment

#### Backend (Node.js/Express)
```bash
cd back-end
npm install
cp .env.example .env
# Configure database and blockchain settings
npm run dev
```

#### Frontend (React)
```bash
cd front-end
npm install
npm start
```

#### Docker Development
```bash
docker-compose -f docker-compose.dev.yml up
```

## Coding Standards

### JavaScript/TypeScript
- Use ESLint and Prettier
- Follow Airbnb JavaScript Style Guide
- Use TypeScript for type safety
- Write meaningful variable and function names
- Comment complex logic

### React Components
- Use functional components with hooks
- Follow React best practices
- Component composition over inheritance
- Prop validation with PropTypes or TypeScript

### Smart Contracts (Solidity)
- Follow Solidity style guide
- Comprehensive testing
- Gas optimization
- Security best practices
- Documentation with NatSpec

### Database
- Use proper indexing
- Follow normalization rules
- Write efficient queries
- Database migrations

## Testing Guidelines

### Required Tests
- Unit tests: minimum 80% coverage
- Integration tests for API endpoints
- Smart contract tests with Hardhat/Truffle
- E2E tests for critical paths

### Running Tests
```bash
# Backend tests
cd back-end
npm test
npm run test:coverage

# Frontend tests
cd front-end
npm test
npm run test:coverage

# Smart contract tests
cd blockchain
npm test
```

## Pull Request Process

### Before Submitting
- [ ] Tests pass
- [ ] Code follows style guidelines
- [ ] Documentation updated
- [ ] No merge conflicts
- [ ] Feature branch up to date

### PR Template
- Clear description of changes
- Link to related issues
- Screenshots for UI changes
- Breaking changes noted
- Testing instructions

### Review Process
1. Code review by maintainers
2. Automated CI/CD checks
3. Security scan
4. Performance impact assessment
5. Documentation review

## Reporting Issues

### Security Issues
- **DO NOT** create public issues for security vulnerabilities
- Email: tracepost.pro@gmail.com
- Use GitHub Security Advisory for responsible disclosure

### Bug Reports
Template:
```markdown
**Bug Description**
Clear and concise description

**Steps to Reproduce**
1. Step 1
2. Step 2
3. See error

**Expected Behavior**
What should happen

**Actual Behavior**
What actually happens

**Environment**
- OS: [e.g. Windows 10]
- Browser: [e.g. Chrome 96]
- Node.js version: [e.g. 16.14.0]
- Contract network: [e.g. Ethereum Mainnet]

**Additional Context**
Screenshots, logs, etc.
```

## Development Best Practices

### Git Commit Messages
```
type(scope): short description

Longer description if needed

- List changes
- Reference issues (#123)
```

Types: feat, fix, docs, style, refactor, test, chore

### Documentation
- Update README for new features
- API documentation with Swagger/OpenAPI
- Inline code documentation
- Architecture decisions in ADR format

### Performance
- Monitor gas usage for smart contracts
- Optimize database queries
- Frontend performance best practices
- Load testing for production

## Getting Help

- GitHub Discussions for questions
- Discord/Slack community
- Weekly community calls
- Documentation wiki

## Recognition

Contributors are recognized in:
- CONTRIBUTORS.md file
- Release notes
- Project website
- Annual contributor rewards

Thank you for contributing to the project! 🚀 