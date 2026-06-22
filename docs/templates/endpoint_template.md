## [HTTP_METHOD] [PATH]

[Provide a 1-2 sentence description of what the endpoint does, triggers, or creates.]

### Details
- **Authentication**: `Required (Bearer JWT) | Public`
- **Scope Limitations**: [e.g. Workspace tenants only]

### Parameters

#### Path Parameters
| Name | Type | Required | Description |
|------|------|----------|-------------|
| [param] | string | Yes | [Description] |

#### Query Parameters
| Name | Type | Required | Description |
|------|------|----------|-------------|
| [query_param] | string | No | [Description] |

### Request Payload (`application/json`)
```json
{
  "field": "type"
}
```

### Responses

#### 200 OK / 201 Created
```json
{
  "message": "success",
  "data": {}
}
```

#### 400 Bad Request
```json
{
  "message": "error details"
}
```
