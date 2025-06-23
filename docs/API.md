# API Documentation

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

All API endpoints (except auth and webhooks) require JWT authentication.

### Headers

```
Authorization: Bearer <jwt_token>
```

## Endpoints

### Authentication

#### Login

```http
POST /auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "User Name",
    "role": "admin"
  }
}
```

#### Register

```http
POST /auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "name": "User Name"
}
```

#### Refresh Token

```http
POST /auth/refresh
Authorization: Bearer <jwt_token>
```

#### Get Current User

```http
GET /auth/me
Authorization: Bearer <jwt_token>
```

### AmoCRM Integration

#### Get Authorization URL

```http
GET /amocrm/auth?state=optional_state
```

Response:
```json
{
  "auth_url": "https://www.amocrm.ru/oauth?..."
}
```

#### OAuth Callback

```http
GET /amocrm/auth/callback?code=authorization_code
```

#### Check Status

```http
GET /amocrm/status
```

Response:
```json
{
  "service": "AmoCRM",
  "initialized": true,
  "authorized": true,
  "token_expires_at": "2024-01-15T10:30:00Z"
}
```

#### Get Fields

```http
GET /amocrm/fields?entity_type=leads
```

Query parameters:
- `entity_type`: `leads` or `contacts`

Response:
```json
[
  {
    "id": 123456,
    "name": "Телефон",
    "type": "multitext",
    "entity_type": "leads"
  }
]
```

#### Sync Fields

```http
POST /amocrm/fields/sync?entity_type=leads
```

#### Get Leads

```http
GET /amocrm/leads?page=1&limit=50
```

Query parameters:
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 50, max: 250)
- `query`: Search query
- `status_id`: Filter by status
- `pipeline_id`: Filter by pipeline
- `responsible_user_id`: Filter by responsible user

Response:
```json
{
  "data": [
    {
      "id": 123456,
      "name": "Test Lead",
      "price": 100000,
      "status_id": 142,
      "pipeline_id": 123,
      "responsible_user_id": 456
    }
  ],
  "page": 1,
  "limit": 50,
  "count": 25
}
```

#### Get Lead by ID

```http
GET /amocrm/leads/123456
```

#### Update Lead

```http
PUT /amocrm/leads/123456
Content-Type: application/json

{
  "name": "Updated Lead Name",
  "price": 150000,
  "status_id": 143
}
```

### Flows

#### Get All Flows

```http
GET /flows
```

Response:
```json
[
  {
    "id": "uuid",
    "name": "Main Flow",
    "is_active": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

#### Get Flow by ID

```http
GET /flows/{id}
```

Response includes full flow configuration with nodes and edges.

#### Create Flow

```http
POST /flows
Content-Type: application/json

{
  "name": "New Flow",
  "flow_data": {
    "nodes": [...],
    "edges": [...]
  },
  "is_active": false
}
```

#### Update Flow

```http
PUT /flows/{id}
Content-Type: application/json

{
  "name": "Updated Flow",
  "flow_data": {...},
  "is_active": true
}
```

#### Delete Flow

```http
DELETE /flows/{id}
```

### Dialer

#### Get Schedulers

```http
GET /dialer/schedulers
```

#### Get Campaigns

```http
GET /dialer/campaigns
```

#### Get Buckets

```http
GET /dialer/buckets?campaign_id=uuid
```

### Webhooks (No authentication required)

#### AmoCRM Lead Webhooks

```http
POST /webhooks/amocrm/lead/add
POST /webhooks/amocrm/lead/update
POST /webhooks/amocrm/lead/delete
POST /webhooks/amocrm/lead/status
POST /webhooks/amocrm/lead/responsible
```

Webhook payload varies by event type. See AmoCRM documentation for details.

## Error Responses

All errors follow the same format:

```json
{
  "error": "Error message",
  "code": 400,
  "details": "Additional error details (optional)"
}
```

Common error codes:
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `500` - Internal Server Error
- `503` - Service Unavailable

## Rate Limiting

- AmoCRM API: 7 requests per second (handled automatically)
- General API: 100 requests per minute per IP

## Pagination

For endpoints that return lists, use these query parameters:
- `page`: Page number (starts from 1)
- `limit`: Items per page (default varies by endpoint)

## Webhooks Security

For production, configure webhook signature verification:

1. Set webhook secret in AmoCRM
2. Add to `.env`: `AMOCRM_WEBHOOK_SECRET=your_secret`
3. System will verify `X-Signature` header

## Examples

### Complete Flow: Create and Activate Integration

1. **Login**
```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@example.com","password":"demo123"}' \
  | jq -r .token)
```

2. **Check AmoCRM Status**
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/amocrm/status
```

3. **Sync Fields**
```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/amocrm/fields/sync?entity_type=leads
```

4. **Create Flow**
```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  http://localhost:8080/api/v1/flows \
  -d '{
    "name": "Auto Dialer Flow",
    "flow_data": {
      "nodes": [...],
      "edges": [...]
    },
    "is_active": true
  }'
```

5. **Test Webhook**
```bash
curl -X POST http://localhost:8080/api/v1/webhooks/amocrm/lead/add \
  -H "Content-Type: application/json" \
  -d '{
    "leads": {
      "add": [{
        "id": 123456,
        "name": "Test Lead"
      }]
    }
  }'
```