# Roadmap

All high-level goals and planned work for this project will be documented in this file.

The roadmap is based on the Now / Next / Later format to communicate current focus, upcoming work and longer-term ideas.

## Now
- www: Improve responsiveness using partial reloads with HTMX and Alpine.js
- www: Review CSRF protection for HTML forms
- www: SameSite cookie policy
- www: Content Security Policy (CSP)
- www: Review OWASP Top 10 checklist

## Next
- Bookmark: Handle conflict with an existing bookmark (URL)
- Feed: Add entry tags, with auto-tagging rules
- Feed: Bookmark entry
- Internal: Rework error flow (logging, metadata)
    - www: Improve error messages

## Later
### Content & Features
- Bookmark: Sanitize URLs to remove tracking parameters
- Bookmark: Detect link rot
- Bookmark, Feed: Store site favicon
- Bookmark, Feed: Store site domain
- Feed: Adapt fetch frequency to entry publication frequency
- Feed: Improve duplicate entry detection
- Search: Query language
- Taxonomy: Tag hierarchy

### Users
- Authentication: Password reset
- Authentication: OAuth2/OpenID
- Authentication: Two-factor authentication
- Documentation: Add a user guide with screenshots
- Users: Audit log

### www
- www: Display curated content on the home page
- www: Responsive design
- www: Internationalization (i18n)
- www: Dark mode / theme switching

### Command-line
- Database: Review connection pool transaction and timeout usage

### API
- API: /health endpoint
- API: /healthcheck endpoint
- API: OpenAPI or gRPC?
- API: Authentication flow

### Integrations
- Integration: Browser extension
- Integration: Archive.org
- Integration: Self-hosted archive
- Integration: News (HN, Lobste.rs)
- Integration: Forges (Github, Gitlab, Gitea/Forgejo)
