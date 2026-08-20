# Roadmap

All high-level goals and planned work for this project will be documented in this file.

The roadmap is based on the Now / Next / Later format to communicate current focus, upcoming work and longer-term ideas.

## Now
- Feed: Tag subscriptions
- Feed: Bookmark entry

## Next
- Feed: Improve duplicate entry detection
- Taxonomy: Tag hierarchy
- Internal: Rework error flow (logging, metadata)
    - www: Improve error messages

## Later
### Content & Features
- Bookmark: Sanitize URLs to remove tracking parameters
- Bookmark: Detect link rot
- Bookmark, Feed: Store site favicon
- Bookmark, Feed: Store site domain
- Feed: Adapt fetch frequency to entry publication frequency
- Search: Query language

### Users
- www: Review OWASP Top 10 checklist
  - Users: Audit log
  - Users: Identify session by user-agent
  - Users: Show active sessions
  - Users: Revoke all sessions
  - Authentication: Password reset
  - Authentication: OAuth2/OpenID
  - Authentication: Two-factor authentication
- Documentation: Add a user guide with screenshots

### www
- www: Display curated content on the home page
- www: Internationalization (i18n)

### Command-line
- Database: Review connection pool transaction and timeout usage

### API
- API: OpenAPI or gRPC?
- API: Authentication flow

### Integrations
- Integration: Browser extension
- Integration: Archive.org
- Integration: Self-hosted archive
- Integration: News (HN, Lobste.rs)
- Integration: Forges (Github, Gitlab, Codeberg, Gitea/Forgejo)
