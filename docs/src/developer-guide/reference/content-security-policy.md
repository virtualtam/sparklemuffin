# Content Security Policy

[Content Security Policy (CSP)](https://en.wikipedia.org/wiki/Content_Security_Policy) is a security
mechanism that helps prevent cross-site scripting (XSS), clickjacking, and other code injection attacks
by controlling which resources can be loaded and executed on a Web page.

SparkleMuffin uses CSP headers to restrict inline scripts, styles, and external resources to trusted sources,
improving the security posture of the application.

## Specifications and Resources

- [Wikipedia - Content Security Policy](https://en.wikipedia.org/wiki/Content_Security_Policy)
- [MDN - Content-Security-Policy header reference](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Content-Security-Policy)
- [web.dev - Content Security Policy](https://web.dev/articles/csp)
- [web.dev - Mitigate cross-site scripting (XSS) with a strict Content Security Policy (CSP)](https://web.dev/articles/strict-csp)
- [Google - CSP Evaluator](https://csp-evaluator.withgoogle.com/) - tool to evaluate CSP policies
- [OWASP - Content Security Policy Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Content_Security_Policy_Cheat_Sheet.html)
