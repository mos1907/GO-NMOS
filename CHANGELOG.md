# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- NMOS Registry (RDS) support for discovering nodes from IS-04 Query API
- Modern dark theme UI redesign with Tailwind CSS
- NMOS Patch Panel with visual router-style interface
- Comprehensive documentation in README.md
- Contributing guidelines (CONTRIBUTING.md)
- Issue and Pull Request templates
- GitHub Actions workflow for PR checks

### Changed
- Complete UI redesign: dark theme throughout all views
- Improved text input styling and consistency
- Enhanced NMOS Patch Panel with better UX

### Fixed
- CORS configuration for local development
- Database connection handling

## [0.1.0] - 2024-02-18

### Added
- Initial release of go-NMOS
- JWT authentication system
- Flow CRUD operations
- NMOS IS-04 discovery and IS-05 connection management
- Collision detection (Checker)
- Automation jobs system
- Address planning (Planner)
- User management
- Settings management
- Log viewing and download
- MQTT real-time updates
- Docker Compose setup
- PostgreSQL database integration
- Svelte 5 frontend with Vite
- Go backend with chi router
- MIT License

[Unreleased]: https://github.com/mos1907/GO-NMOS/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/mos1907/GO-NMOS/releases/tag/v0.1.0
