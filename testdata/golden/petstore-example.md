# Pet Store API

A comprehensive API for managing a pet store with users, pets, and orders

## Table of Contents

HTTP Request | Description
-------------|------------
GET [/v3/pets](#getv3pets) | List all pets
POST [/v3/pets](#postv3pets) | Create a new pet
POST [/v3/pets.delete](#postv3petsdelete) | Delete a pet
GET [/v3/pets/{petId}](#getv3petspetid) | Get a pet by ID
GET [/v3/users](#getv3users) | List all users
POST [/v3/users](#postv3users) | Create a new user
GET [/v3/users/{userId}](#getv3usersuserid) | Get user by ID
GET [/v3/users/{userId}/orders](#getv3usersuseridorders) | Get user orders
GET [/v3/orders](#getv3orders) | List all orders
POST [/v3/orders](#postv3orders) | Create a new order
GET [/v3/orders/{orderId}](#getv3ordersorderid) | Get order by ID
GET [/v3/health](#getv3health) | Health check endpoint
GET [/v3/metrics](#getv3metrics) | Get API metrics

## admin

### POST /v3/pets.delete

Removes a pet from the store inventory

#### Header Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
X-Admin-Token | Admin authorization token | true | string

### Request

```json
{
   "petId": "dl2INvNSQT"
}
```

#### Field Definitions

**petId** (string, required)
- ID of pet to delete

### Responses

#### 200 Response

Pet deleted successfully

```json
{
   "pet": {
      "petId": "Z5zQu9MxNm",
      "success": false
   }
}
```

#### Field Definitions

**pet** (object)

#### 403 Response

Unauthorized

```json
{
   "code": 6,
   "error": "An error occurred"
}
```

#### 404 Response

Pet not found

```json
{
   "code": 6,
   "error": "An error occurred"
}
```

### GET /v3/metrics

Returns usage metrics and statistics

#### Header Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
X-Admin-Token | Admin authorization token | true | string

### Responses

#### 200 Response

Metrics data

```json
{
   "requestsPerSecond": 19.31071700388466,
   "requestsTotal": 49,
   "uptime": 54
}
```

#### Field Definitions

**requestsTotal** (integer)

**requestsPerSecond** (number)

**uptime** (integer)

#### 403 Response

Unauthorized

```json
{
   "code": 6,
   "error": "An error occurred"
}
```

## orders

### GET /v3/users/{userId}/orders

Returns all orders placed by a specific user

#### Path Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
userId | User identifier | true | string

#### Query Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
status | Filter by order status | false | string

### Responses

#### 200 Response

Successful response

```json
{
   "cursor": "dNLyvxI2qwpXy1MzNj",
   "orders": [
      {
         "id": "u3ujt5jSrl",
         "petId": "MmGDTQm9Oy",
         "quantity": 33,
         "status": "placed",
         "userId": "vzERQgJPPe"
      }
   ]
}
```

#### Field Definitions

**orders** (array of objects)

**cursor** (string)

**Order**
- `id` (string)
- `userId` (string)
- `petId` (string)
- `status` (string)
- `quantity` (integer)

#### 404 Response

User not found

```json
{
   "code": 6,
   "error": "An error occurred"
}
```

### GET /v3/orders

Returns a list of all orders in the system

#### Query Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
after | Cursor for pagination - returns orders after this cursor | false | string
first | Number of orders to return (cursor-based pagination) | false | integer

### Responses

#### 200 Response

Successful response

```json
{
   "cursor": "dNLyvxI2qwpXy1MzNj",
   "orders": [
      {
         "id": "u3ujt5jSrl",
         "petId": "MmGDTQm9Oy",
         "quantity": 33,
         "status": "placed",
         "userId": "vzERQgJPPe"
      }
   ]
}
```

#### Field Definitions

**orders** (array of objects)

**cursor** (string)

**Order**
- `id` (string)
- `userId` (string)
- `petId` (string)
- `status` (string)
- `quantity` (integer)

### POST /v3/orders

Place a new order for pets

### Responses

#### 201 Response

Order created successfully

```json
{
   "id": "Nb6pTA5ook",
   "message": "This is a message"
}
```

#### Field Definitions

**id** (string)

**message** (string)

#### 400 Response

Invalid order data

```json
{
   "code": 6,
   "error": "An error occurred"
}
```

### GET /v3/orders/{orderId}

Returns detailed order information

#### Path Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
orderId | Order identifier | true | string

### Responses

#### 200 Response

Successful response

```json
{
   "id": "3tnghwev4q",
   "petId": "KloHxMixtq",
   "quantity": 18,
   "status": "placed",
   "userId": "MlFi3bO9SE"
}
```

#### Field Definitions

**id** (string)

**userId** (string)

**petId** (string)

**status** (string)

**quantity** (integer)

#### 404 Response

Order not found

```json
{
   "code": 6,
   "error": "An error occurred"
}
```

## pets

### GET /v3/pets

Returns a paginated list of all pets in the store

#### Query Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
after | Cursor for pagination - returns pets after this cursor | false | string
first | Number of pets to return (cursor-based pagination) | false | integer
tag | Filter by tag | false | string

### Responses

#### 200 Response

Successful response with pet list

```json
{
   "cursor": "eyJpZCI6InBldC0xMjMifQ==",
   "pets": [
      {
         "id": "pet-123",
         "name": "Fluffy",
         "status": "available",
         "tags": [
            "cat",
            "friendly"
         ]
      }
   ]
}
```

#### Field Definitions

**pets** (array of objects)

**cursor** (string)

**Pet**
- `id` (string)
- `name` (string)
- `status` (string)
- `tags` (string array)

#### 400 Response

Invalid request parameters

```json
{
   "code": 6,
   "error": "An error occurred"
}
```

### POST /v3/pets

Adds a new pet to the store inventory

#### Header Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
X-Request-ID | Unique request identifier | true | string

### Responses

#### 201 Response

Pet created successfully

```json
{
   "id": "kwHUMGhWzG",
   "message": "This is a message"
}
```

#### Field Definitions

**id** (string)

**message** (string)

#### 400 Response

Invalid pet data

```json
{
   "code": 6,
   "error": "An error occurred"
}
```

### POST /v3/pets.delete

Removes a pet from the store inventory

#### Header Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
X-Admin-Token | Admin authorization token | true | string

### Request

```json
{
   "petId": "dl2INvNSQT"
}
```

#### Field Definitions

**petId** (string, required)
- ID of pet to delete

### Responses

#### 200 Response

Pet deleted successfully

```json
{
   "pet": {
      "petId": "Z5zQu9MxNm",
      "success": false
   }
}
```

#### Field Definitions

**pet** (object)

#### 403 Response

Unauthorized

```json
{
   "code": 6,
   "error": "An error occurred"
}
```

#### 404 Response

Pet not found

```json
{
   "code": 6,
   "error": "An error occurred"
}
```

### GET /v3/pets/{petId}

Returns detailed information about a specific pet

#### Path Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
petId | ID of pet to return | true | string

### Responses

#### 200 Response

Successful response

```json
{
   "id": "3gDk8Bg7W9",
   "name": "LLxq2zGNO6",
   "status": "available",
   "tags": [
      "q1Xh3S7gYe"
   ]
}
```

#### Field Definitions

**id** (string)

**name** (string)

**status** (string)

**tags** (string array)

#### 404 Response

Pet not found

```json
{
   "code": 6,
   "error": "An error occurred"
}
```

## users

### GET /v3/users

Returns a list of registered users

#### Query Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
active | Filter by active status | false | boolean

### Responses

#### 200 Response

Successful response

```json
{
   "users": [
      {
         "active": true,
         "email": "user@example.com",
         "id": "pld7aFPfYJ",
         "username": "K6SV75azeo"
      }
   ]
}
```

#### Field Definitions

**users** (array of objects)

**User**
- `id` (string)
- `username` (string)
- `email` (string)
- `active` (boolean)

### POST /v3/users

Register a new user account

### Responses

#### 201 Response

User created successfully

```json
{
   "id": "vf8sRN3aXc",
   "message": "This is a message"
}
```

#### Field Definitions

**id** (string)

**message** (string)

#### 400 Response

Invalid user data

```json
{
   "code": 6,
   "error": "An error occurred"
}
```

### GET /v3/users/{userId}

Returns detailed user information

#### Path Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
userId | User identifier | true | string

### Responses

#### 200 Response

Successful response

```json
{
   "active": true,
   "email": "user@example.com",
   "id": "0L8r30xvTn",
   "username": "j31WE1Wf9y"
}
```

#### Field Definitions

**id** (string)

**username** (string)

**email** (string)

**active** (boolean)

#### 404 Response

User not found

```json
{
   "code": 6,
   "error": "An error occurred"
}
```

### GET /v3/users/{userId}/orders

Returns all orders placed by a specific user

#### Path Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
userId | User identifier | true | string

#### Query Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
status | Filter by order status | false | string

### Responses

#### 200 Response

Successful response

```json
{
   "cursor": "dNLyvxI2qwpXy1MzNj",
   "orders": [
      {
         "id": "u3ujt5jSrl",
         "petId": "MmGDTQm9Oy",
         "quantity": 33,
         "status": "placed",
         "userId": "vzERQgJPPe"
      }
   ]
}
```

#### Field Definitions

**orders** (array of objects)

**cursor** (string)

**Order**
- `id` (string)
- `userId` (string)
- `petId` (string)
- `status` (string)
- `quantity` (integer)

#### 404 Response

User not found

```json
{
   "code": 6,
   "error": "An error occurred"
}
```

## Default APIs

### GET /v3/health

Returns the health status of the API

### Responses

#### 200 Response

Service is healthy

```json
{
   "status": "healthy",
   "timestamp": "2024-01-15T10:30:00Z"
}
```

#### Field Definitions

**status** (string)

**timestamp** (string)

