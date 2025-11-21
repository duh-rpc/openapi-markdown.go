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

##### 200 Response

Pet deleted successfully

```json
{
   "pet": {
      "petId": "Z5zQu9MxNm",
      "success": false
   }
}
```

##### 403 Response

Unauthorized

```json
{
   "code": 0,
   "error": "qDuuLEkoRU"
}
```

##### 404 Response

Pet not found

```json
{
   "code": 0,
   "error": "qDuuLEkoRU"
}
```

### GET /v3/metrics

Returns usage metrics and statistics

#### Header Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
X-Admin-Token | Admin authorization token | true | string

##### 200 Response

Metrics data

```json
{
   "requestsPerSecond": 0,
   "requestsTotal": 0,
   "uptime": 0
}
```

##### 403 Response

Unauthorized

```json
{
   "code": 0,
   "error": "qDuuLEkoRU"
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

##### 200 Response

Successful response

```json
{
   "cursor": "iehJer4vjD",
   "orders": [
      {
         "id": "vzERQgJPPe",
         "petId": "QDX5B43Vis",
         "quantity": 0,
         "status": "placed",
         "userId": "MmGDTQm9Oy"
      }
   ]
}
```

##### 404 Response

User not found

```json
{
   "code": 0,
   "error": "qDuuLEkoRU"
}
```

### GET /v3/orders

Returns a list of all orders in the system

#### Query Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
after | Cursor for pagination - returns orders after this cursor | false | string
first | Number of orders to return (cursor-based pagination) | false | integer

##### 200 Response

Successful response

```json
{
   "cursor": "iehJer4vjD",
   "orders": [
      {
         "id": "vzERQgJPPe",
         "petId": "QDX5B43Vis",
         "quantity": 0,
         "status": "placed",
         "userId": "MmGDTQm9Oy"
      }
   ]
}
```

### POST /v3/orders

Place a new order for pets

##### 201 Response

Order created successfully

```json
{
   "id": "tNb6pTA5oo",
   "message": "kqNDD63C2T"
}
```

##### 400 Response

Invalid order data

```json
{
   "code": 0,
   "error": "qDuuLEkoRU"
}
```

### GET /v3/orders/{orderId}

Returns detailed order information

#### Path Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
orderId | Order identifier | true | string

##### 200 Response

Successful response

```json
{
   "id": "3tnghwev4q",
   "petId": "KloHxMixtq",
   "quantity": 0,
   "status": "placed",
   "userId": "MlFi3bO9SE"
}
```

##### 404 Response

Order not found

```json
{
   "code": 0,
   "error": "qDuuLEkoRU"
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

##### 200 Response

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

##### 400 Response

Invalid request parameters

```json
{
   "code": 0,
   "error": "qDuuLEkoRU"
}
```

### POST /v3/pets

Adds a new pet to the store inventory

#### Header Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
X-Request-ID | Unique request identifier | true | string

##### 201 Response

Pet created successfully

```json
{
   "id": "q1Xh3S7gYe",
   "message": "kwHUMGhWzG"
}
```

##### 400 Response

Invalid pet data

```json
{
   "code": 0,
   "error": "qDuuLEkoRU"
}
```

### POST /v3/pets.delete

Removes a pet from the store inventory

#### Header Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
X-Admin-Token | Admin authorization token | true | string

##### 200 Response

Pet deleted successfully

```json
{
   "pet": {
      "petId": "Z5zQu9MxNm",
      "success": false
   }
}
```

##### 403 Response

Unauthorized

```json
{
   "code": 0,
   "error": "qDuuLEkoRU"
}
```

##### 404 Response

Pet not found

```json
{
   "code": 0,
   "error": "qDuuLEkoRU"
}
```

### GET /v3/pets/{petId}

Returns detailed information about a specific pet

#### Path Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
petId | ID of pet to return | true | string

##### 200 Response

Successful response

```json
{
   "id": "sLiFD4MY7O",
   "name": "3gDk8Bg7W9",
   "status": "available",
   "tags": [
      "LLxq2zGNO6"
   ]
}
```

##### 404 Response

Pet not found

```json
{
   "code": 0,
   "error": "qDuuLEkoRU"
}
```

## users

### GET /v3/users

Returns a list of registered users

#### Query Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
active | Filter by active status | false | boolean

##### 200 Response

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

### POST /v3/users

Register a new user account

##### 201 Response

User created successfully

```json
{
   "id": "vf8sRN3aXc",
   "message": "u3ujt5jSrl"
}
```

##### 400 Response

Invalid user data

```json
{
   "code": 0,
   "error": "qDuuLEkoRU"
}
```

### GET /v3/users/{userId}

Returns detailed user information

#### Path Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
userId | User identifier | true | string

##### 200 Response

Successful response

```json
{
   "active": true,
   "email": "user@example.com",
   "id": "0L8r30xvTn",
   "username": "j31WE1Wf9y"
}
```

##### 404 Response

User not found

```json
{
   "code": 0,
   "error": "qDuuLEkoRU"
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

##### 200 Response

Successful response

```json
{
   "cursor": "iehJer4vjD",
   "orders": [
      {
         "id": "vzERQgJPPe",
         "petId": "QDX5B43Vis",
         "quantity": 0,
         "status": "placed",
         "userId": "MmGDTQm9Oy"
      }
   ]
}
```

##### 404 Response

User not found

```json
{
   "code": 0,
   "error": "qDuuLEkoRU"
}
```

## Default APIs

### GET /v3/health

Returns the health status of the API

##### 200 Response

Service is healthy

```json
{
   "status": "healthy",
   "timestamp": "2024-01-15T10:30:00Z"
}
```

