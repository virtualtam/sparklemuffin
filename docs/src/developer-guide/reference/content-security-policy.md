# Content Security Policy

[Content Security Policy (CSP)](https://en.wikipedia.org/wiki/Content_Security_Policy) is a security
mechanism. It helps prevent cross-site scripting (XSS), clickjacking, and other code-injection attacks.
CSP controls which resources a Web page can load and run.

SparkleMuffin uses CSP headers to restrict inline scripts, styles, and external resources to trusted
sources. This makes the application more secure.

## Specifications and Resources

- [Wikipedia - Content Security Policy](https://en.wikipedia.org/wiki/Content_Security_Policy)
- [MDN - Content-Security-Policy header reference](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Content-Security-Policy)
- [web.dev - Content Security Policy](https://web.dev/articles/csp)
- [web.dev - Mitigate cross-site scripting (XSS) with a strict Content Security Policy (CSP)](https://web.dev/articles/strict-csp)
- [Google - CSP Evaluator](https://csp-evaluator.withgoogle.com/) - tool to evaluate CSP policies
- [OWASP - Content Security Policy Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Content_Security_Policy_Cheat_Sheet.html)
