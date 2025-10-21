# Changelog

All notable changes to this project will be documented in this file.

## [2.0.0] - 2025-10-21

### Added
#### Middleware Enhancements
- ✨ **Enhanced Request Logger** - Now tracks response status, size, and duration
- ✨ **Request Timeout Middleware** - 30-second timeout for all requests
- ✨ **Compression Middleware** - Vary header support for future gzip implementation
- ✨ **Improved Rate Limiter** - JSON error responses with Retry-After header
- ✨ **Enhanced Security Headers** - Added CSP, Referrer-Policy, and Permissions-Policy
- ✨ **Advanced Panic Recovery** - Detailed logging with URL and remote address

#### Server Improvements
- 🚀 **Improved Graceful Shutdown** - Extended timeout to 30s with WebSocket cleanup
- 🔌 **WebSocket Hub Shutdown** - Sends shutdown message to all clients before closing
- ✅ **Input Validation** - Email and name validation for user creation
- 📊 **Response Writer Wrapper** - Tracks HTTP status codes and response sizes
- 💬 **Better Error Messages** - Standardized JSON error responses

#### Frontend Enhancements
- 🎨 **Advanced CSS Variables** - 4-tier shadow system (sm, md, lg, xl)
- ✨ **Sophisticated Gradients** - Subtle linear, radial, and conic gradients
- 🖱️ **Mouse Tracking** - Interactive radial gradients following cursor
- 🔄 **Rotating Glow Effects** - Animated conic gradients on cards
- 💫 **Enhanced Particles** - 60 particles with variable opacity and size
- 🎭 **Elastic Animations** - Custom cubic-bezier timing functions
- 🔘 **Ripple Effects** - Expanding circles on button interactions
- 📱 **Improved Typography** - Text gradients and enhanced readability

### Changed
- 📝 **Logging Format** - Arrows (→ ←) for request start/completion
- ⏱️ **Shutdown Timeout** - Increased from 10s to 30s for better cleanup
- 🎨 **Method Badges** - Now uppercase with letter-spacing
- 🔧 **Container Design** - Added glassmorphism with backdrop blur
- 📊 **Code Blocks** - Improved contrast and subtle top borders

### Enhanced
- 🛡️ **Security** - More comprehensive security headers
- 📈 **Observability** - Better request/response tracking
- 🎯 **Error Handling** - More descriptive error messages
- ⚡ **Performance** - Optimized animations and transitions
- 🎨 **Visual Design** - Professional monochrome aesthetic

## [1.0.0] - 2025-10-20

### Initial Release
- ✅ Basic Go features demonstration
- ✅ HTTP server with REST API
- ✅ WebSocket support
- ✅ SQLite database integration
- ✅ Concurrency patterns
- ✅ Generics and reflection
- ✅ Monochrome frontend design

---

**Legend:**
- ✨ Feature
- 🚀 Performance
- 🐛 Bug Fix
- 📝 Documentation
- 🎨 UI/UX
- 🔧 Configuration
- 🛡️ Security
- 📊 Analytics
- ✅ Addition
- 🔄 Change
- ❌ Removal

