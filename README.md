# URL Shortener

A high-performance URL shortener built with Go, Gin, and SQLite.

## Features
- Shorten URLs with custom aliases
- Authentication via JWT
- Track access statistics (daily, weekly, monthly)
- User dashboard for managing URLs
- QR Code generation
- First-access Webhook callbacks
- Per-IP Rate Limiting

## API Endpoints

### Auth
```bash
# Register
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com", "password":"password123"}'

# Login
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com", "password":"password123"}'
```

### URLs
```bash
# Shorten a URL (Optional Auth)
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://google.com",
    "custom": "my-google",
    "webhook_url": "https://webhook.site/my-webhook-url"
  }'

# Redirect
curl -L http://localhost:8080/my-google

# View QR Code
curl -o qrcode.png http://localhost:8080/my-google/qr

# View Stats
curl http://localhost:8080/my-google/stats
```

### Management (Auth Required)
```bash
# List User URLs
curl -H "Authorization: Bearer <TOKEN>" http://localhost:8080/me/urls

# Update URL
curl -X PATCH -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://bing.com"}' \
  http://localhost:8080/my-google

# Delete URL
curl -X DELETE -H "Authorization: Bearer <TOKEN>" http://localhost:8080/my-google
```
