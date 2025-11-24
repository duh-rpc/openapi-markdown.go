# Example API API

A long description describing the api intent

## Table of Contents

HTTP Request | Description
-------------|------------
GET [/v1/authorization/check](#postv2AuthorizationCheck) | List all pets

### POST /v1/authorization/check

This is a long description of this endpoint with details and expressed intent of the endpoint.

#### Headers

Name | Description | Required | Type
-----|-------------|----------|-----
X-Admin-Token | Admin authorization token | true | string

### Request

```json
{
  "principal": {
    "type": "API_KEY",
    "id": "ak_1234567890abcdef"
  },
  "resource": {
    "type": "VAULT",
    "id": "vault_9876543210fedcba"
    "nested": {
      "type": "THING",
      "name": "the thing"
    }
  },
  "actions": [
    "VAULTS_DETAILS_VIEW",
    "WITHDRAW_INITIATE"
  ]
}
```

#### Field Definitions

**principal**
- `type` (string, required): The type of principal. Enums: `API_KEY`, `USER`, `SERVICE_ACCOUNT`, `OAUTH_TOKEN`
- `id` (string, required): The identifier for the principal

**resource**
- `type` (string, required): The type of resource. Enums: `VAULT`, `ORGANIZATION`, `ACCOUNT`, `TRANSACTION`
- `id` (string, required): The identifier for the resource
- `nested` (object, required): A description of the nested thing

**nested**
- `type` (string, required): A description of this field. Enums: `YOUR`, `MOM`
- `name` (string): A description of this field.

**actions** (string enum, required) A description of this field. With multiple sentences. Enums: `VAULTS_DETAILS_VIEW`, `WITHDRAW_INITIATE`, `DEPOSIT_INITIATE`, `ORGANIZATION_SETTINGS_UPDATE`, `USER_INVITE`

### Responses

#### 200 Response

Description of the response

```json
{
  "allowed": true,
  "results": [
    {
      "action": "VAULTS_DETAILS_VIEW",
      "allowed": true
    },
    {
      "action": "WITHDRAW_INITIATE",
      "allowed": true
    }
  ]
}
```

#### Field Definitions

**allowed** (boolean)
- `true` if ALL actions are allowed
- `false` if ANY action is denied (atomic behavior)

**results** (array of objects)
- One entry per requested action, in the same order as the request
- `action` (string): The action that was checked
- `allowed` (boolean): Whether this specific action is allowed

#### 400 Response
```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "Invalid request format",
    "details": {
      "field": "actions",
      "reason": "must contain at least one action"
    }
  }
}
```

#### 404 Response
```json
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "Resource not found",
    "details": {
      "resource_type": "VAULT",
      "resource_id": "vault_invalid"
    }
  }
}
```