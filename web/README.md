# UAST Mapping Development Service

A modern React application built with Radix UI for interactive UAST (Universal Abstract Syntax Tree) mapping development.

## Features

- **Interactive Code Parsing**: Parse code automatically as you type (with debouncing)
- **Shared UAST Output**: View parsed UAST structure and query results in the same area
- **Interactive Querying**: Execute UAST queries automatically as you type (with debouncing)
- **Floating Examples Panel**: Access example queries via a floating button
- **Dark Theme**: Modern dark theme optimized for code and JSON viewing
- **Full-Screen Layout**: Maximizes available space for code and results
- **Multiple Language Support**: Support for 25+ programming languages
- **Real-time Feedback**: Instant parsing and query results

## Tech Stack

- **React 18**: Modern React with hooks
- **Radix UI**: Accessible, unstyled UI components
- **Tailwind CSS**: Utility-first CSS framework
- **Vite**: Fast build tool and development server
- **Lucide React**: Beautiful icons

## Getting Started

### Prerequisites

- Node.js 16+ 
- npm or yarn
- Go 1.19+ (for backend API)

### Quick Start (Recommended)

From the project root, run:
```bash
make uast-dev
```

This will start both the frontend and backend servers:
- Frontend: http://localhost:3000
- Backend: http://localhost:8080

### Manual Installation

1. Install dependencies:
```bash
cd web
npm install
```

2. Start the development server:
```bash
npm run dev
```

3. In another terminal, start the backend:
```bash
uast server --static . --port 8080
```

4. Open your browser and navigate to `http://localhost:3000`

### Building for Production

```bash
npm run build
```

The built files will be in the `dist` directory.

## Usage

1. **Code Input**: Enter your code in the left panel - it will parse automatically as you type
2. **Language Selection**: Choose your programming language from the dropdown
3. **Query Input**: Enter UAST queries in the right panel - results appear automatically
4. **Examples**: Click the floating button (⚡) to access example queries
5. **Shared Output**: View both parsed UAST and query results in the same area

## API Endpoints

The application expects the following API endpoints to be available:

- `POST /api/parse` - Parse code and return UAST
- `POST /api/query` - Execute UAST queries

## Development

### Project Structure

```
src/
├── components/
│   └── ui/          # Radix UI components
├── lib/
│   └── utils.js     # Utility functions
├── App.jsx          # Main application component
├── main.jsx         # Application entry point
└── index.css        # Global styles
```

### Available Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run preview` - Preview production build
- `npm run lint` - Run ESLint
- `npm run lint:fix` - Fix ESLint issues
- `npm test` - Run UI tests
- `npm run test:ui` - Run tests with UI
- `npm run test:headed` - Run tests in headed mode
- `npm run test:debug` - Run tests in debug mode

### Makefile Commands (from project root)

- `make uast-dev` - Start both frontend and backend servers
- `make uast-dev-stop` - Stop both servers
- `make uast-dev-status` - Check server status
- `make uast-test` - Run UI tests
- `make help` - Show all available commands

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linting
5. Submit a pull request

## License

MIT 