# Release v0.1.0 - Initial Release

**Release Date:** February 18, 2024

## 🎉 First Public Release

This is the initial public release of go-NMOS, a production-oriented NMOS management stack built with Go and Svelte.

## ✨ Features

### Core Functionality
- **JWT Authentication** - Secure login with role-based access control (admin, editor, viewer)
- **Flow Management** - Complete CRUD operations for multicast flows
- **NMOS Integration** - IS-04 discovery and IS-05 connection management
- **NMOS Patch Panel** - Visual router-style interface for sender/receiver patching
- **NMOS Registry (RDS) Support** - Connect to IS-04 Query API registries
- **Collision Detection** - Automatic detection of IP/port conflicts
- **Automation Jobs** - Scheduled tasks for flow management
- **Address Planning** - Hierarchical address bucket management

### User Interface
- Modern dark theme with Tailwind CSS
- Responsive design
- Real-time updates via MQTT
- Intuitive NMOS Patch Panel
- Comprehensive flow search and filtering

### Infrastructure
- Docker Compose setup for easy deployment
- PostgreSQL database
- MQTT broker integration (Mosquitto)
- RESTful API with Go backend
- Svelte 5 frontend with Vite

## 🚀 Quick Start

```bash
# Clone the repository
git clone https://github.com/mos1907/GO-NMOS.git
cd GO-NMOS

# Configure backend
cp backend/env.example backend/.env
# Edit backend/.env with your settings

# Start services
make up
```

Access:
- **UI**: http://localhost:4173
- **API**: http://localhost:9090/api/health
- **Default credentials**: admin / change-this-password

## 📚 Documentation

- [README.md](README.md) - Complete setup and usage guide
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guidelines
- [GITHUB_SETUP.md](GITHUB_SETUP.md) - GitHub repository setup guide

## 🔧 Requirements

- Docker Desktop / Docker Engine + Docker Compose
- Go 1.22+ (for local development)
- Node.js 18+ (for local frontend development)

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

Built with:
- [Go](https://golang.org/)
- [Svelte](https://svelte.dev/)
- [Tailwind CSS](https://tailwindcss.com/)
- [PostgreSQL](https://www.postgresql.org/)
- [Mosquitto](https://mosquitto.org/)

## 🔗 Links

- **Repository**: https://github.com/mos1907/GO-NMOS
- **Issues**: https://github.com/mos1907/GO-NMOS/issues
- **Discussions**: https://github.com/mos1907/GO-NMOS/discussions

---

**Note**: This is the initial release. We welcome feedback, bug reports, and contributions!
