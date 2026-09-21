# Go Forum

<img src="https://i.ibb.co/bg0JQ4vD/image-2026-09-21-21-26-01.png" alt="Home page" width="500" height="300">
<img src="https://i.ibb.co/0RNcRM0T/image-2026-09-21-21-23-45.png" alt="Home page" width="500" height="300">
<img src="https://i.ibb.co/GQ0yJ7y3/image-2026-09-21-21-30-51.png" alt="Home page" width="500" height="300">

A server-rendered forum application built with Go, Gorilla Mux, SQLite, and HTML templates. The project demonstrates the core workflows of a classic discussion board: account registration, login/logout, forum categories, topics, replies, likes, editable content, profiles, user groups, and an admin control panel.

Current version: **2.1.1**  
See [Changelog.txt](Changelog.txt) for version history.

## Why this project exists

I built this project to practice full-stack web development in Go without hiding the application behind a large framework. The code shows routing, request handling, HTML template rendering, SQL persistence, session cookies, role-based behavior, and Docker-based deployment in a compact codebase that reviewers can inspect quickly.

## Features

- User registration and login
- HTTPS local server with secure, HTTP-only session cookies
- Password validation and hashed password storage
- Forum categories, topics, and posts
- Post creation, editing, deletion, and likes
- Public user profiles and "all posts by user" pages
- Personal profile page with password change flow
- Admin control panel
- User search and user group management
- Admin moderation for categories, topics, and posts
- SQLite database stored in `db/forum.db`
- Dockerfile for containerized runs

## Tech Stack

- **Language:** Go 1.20
- **Router:** Gorilla Mux
- **Database:** SQLite via `modernc.org/sqlite`
- **Frontend:** HTML, CSS, static assets
- **Deployment:** Docker

## Quick Start

### Prerequisites

- Go 1.20 or newer
- A browser that can open a local HTTPS development certificate

### Run locally

```bash
go run .
```

Then open:

```text
https://localhost:8080/
```

The app uses the included local certificate files:

- `localhost+1.pem`
- `localhost+1-key.pem`

Your browser may show a local certificate warning. This is expected for a development certificate.

## Docker

Build the image:

```bash
docker build -t go-forum .
```

Run the container:

```bash
docker run --name go-forum -p 8080:8080 go-forum
```

Then visit:

```text
https://localhost:8080/
```

Useful Docker commands:

```bash
docker stop go-forum
docker start go-forum
docker rm go-forum
docker rmi go-forum
```

## Project Structure

```text
.
|-- server.go              # App entry point, route registration, static files, HTTPS server
|-- models/
|   `-- storedata.go       # Shared data structs and database handle
|-- web/
|   |-- handlers.go        # Main forum, auth, profile, post, and like handlers
|   |-- admin.go           # Admin panel and moderation handlers
|   `-- functions.go       # Shared helper functions
|-- templates/             # Go HTML templates
|-- css/                   # Stylesheets
|-- img/                   # Images and favicon
|-- fonts/                 # Local fonts
|-- static/                # Additional static assets
|-- db/
|   `-- forum.db           # SQLite database
|-- Dockerfile
|-- go.mod
|-- go.sum
`-- Changelog.txt
```

## Application Flow

1. `server.go` opens the SQLite database and registers all HTTP routes.
2. Static assets are served from `css`, `img`, `fonts`, and `static`.
3. Handlers in `web/handlers.go` manage forum browsing, registration, login, profiles, posts, and likes.
4. Handlers in `web/admin.go` manage admin-only moderation and user group actions.
5. Templates in `templates/` render the server-side HTML responses.

## Reviewer Notes

Good places to start when reviewing the code:

- [server.go](server.go) for routing and app startup
- [web/handlers.go](web/handlers.go) for the main user-facing workflows
- [web/admin.go](web/admin.go) for admin and moderation behavior
- [models/storedata.go](models/storedata.go) for the main data structures
- [templates/header.html](templates/header.html) for navigation state based on login/admin status

The project intentionally keeps the architecture simple so the request flow is easy to follow. It is a practical portfolio project rather than a production-ready forum platform.

## Security-Related Work Included

- HTTPS local development server
- Secure and HTTP-only cookies
- Cryptographically random session token generation
- Case-insensitive username login handling
- Generic login failure message
- Password complexity checks
- Hashed passwords
- Authorization checks for editing posts

## Current Limitations and Future Improvements

The next improvements I would prioritize are:

- Add automated tests for handlers, auth flows, and admin actions
- Move session and login state away from package-level global variables
- Add database migrations instead of creating tables inside request handlers
- Improve SQL consistency and reduce repeated query logic
- Add CSRF protection for form submissions
- Add environment-based configuration for port, database path, and certificate files
- Improve Docker image size with a multi-stage build

## Version History

- **v2.1.1** - new icons and headers were added
- **v2.1** - code cleanup and refactoring
- **v2.0** - admin control panel, moderation functions, user groups, profile password changes, and permission fixes
- **v1.4** - HTTPS support, secure cookies, improved session token generation
- **v1.2** - password requirements, hashed passwords, and SQLite dependency cleanup

Full details are available in [Changelog.txt](Changelog.txt).
