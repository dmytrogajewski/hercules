# UI Tests for UAST Development Service

This directory contains Playwright UI tests for the UAST Development Service.

## Test Files

- **`simple.spec.js`** - Basic functionality tests (3 tests)
- **`basic.spec.js`** - Comprehensive UI tests (10 tests)
- **`api-simple.spec.js`** - API integration tests (4 tests)
- **`app.spec.js`** - Original comprehensive tests (may have selector issues)
- **`api.spec.js`** - Original API tests (may have selector issues)

## Running Tests

### Prerequisites

1. Make sure the development environment is running:
   ```bash
   make uast-dev
   ```

2. Install Playwright dependencies:
   ```bash
   cd web
   npm install
   npx playwright install
   ```

### Running Tests

#### From project root:
```bash
# Run basic tests
make uast-test

# Run all tests
cd web && npm test

# Run specific test file
cd web && npm test -- tests/simple.spec.js
```

#### From web directory:
```bash
# Run all tests
npm test

# Run specific test file
npm test -- tests/simple.spec.js

# Run tests with UI
npm run test:ui

# Run tests in headed mode
npm run test:headed

# Run tests in debug mode
npm run test:debug
```

## Test Configuration

Tests are configured in `playwright.config.js`:
- Single browser (Chromium) for faster execution
- Single worker to avoid conflicts
- Base URL: `http://localhost:3000`
- HTML reporter enabled

## Test Results

Test results are available in the HTML report:
```bash
npx playwright show-report
```

## Test Coverage

The tests cover:

### Basic Functionality
- ✅ Page loading and title
- ✅ Tab navigation
- ✅ Code editor visibility
- ✅ Language selector
- ✅ Button states

### UI Interactions
- ✅ Tab switching
- ✅ Language selection
- ✅ Code input
- ✅ Responsive design

### API Integration
- ✅ Code parsing
- ✅ Query execution
- ✅ Error handling (partial)

## Troubleshooting

### Tests hanging
- Ensure only one test process is running
- Check that the development server is running
- Restart the development environment if needed

### Selector issues
- Use role-based selectors when possible
- Avoid generic text selectors that match multiple elements
- Use specific CSS classes or IDs

### API test failures
- Ensure the backend server is running on port 8080
- Check network connectivity
- Verify API endpoints are responding

## Adding New Tests

1. Create a new test file in the `tests/` directory
2. Use descriptive test names
3. Use role-based selectors when possible
4. Add appropriate timeouts for async operations
5. Test both success and error scenarios 