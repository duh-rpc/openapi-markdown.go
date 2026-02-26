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

## Admin

### POST /v3/pets.delete

Removes a pet from the store inventory. Requires admin privileges to perform this operation

#### Headers

Name | Description | Required | Type
-----|-------------|----------|-----
X-Admin-Token | Admin authorization token required for this operation | true | string

### Request

```json
{
   "petId": "dl2INvNSQT"
}
```

#### Field Definitions

- `petId` *(string, required)* Unique identifier of the pet to delete from inventory

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

- `pet` *(object)* Information about the deleted pet

#### 403 Response

Unauthorized - invalid or missing admin token

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

Returns usage metrics and statistics including request counts, throughput, and uptime. Requires admin privileges

#### Headers

Name | Description | Required | Type
-----|-------------|----------|-----
X-Admin-Token | Admin authorization token required for accessing metrics | true | string

### Responses

#### 200 Response

Metrics data retrieved successfully

```json
{
   "requestsPerSecond": 19.31071700388466,
   "requestsTotal": 49,
   "uptime": 54
}
```

#### Field Definitions

- `requestsTotal` *(integer)* Total number of API requests processed since service start
- `requestsPerSecond` *(number)* Current request throughput measured in requests per second
- `uptime` *(integer)* Server uptime in seconds since last restart

#### 403 Response

Unauthorized - invalid or missing admin token

```json
{
   "code": 6,
   "error": "An error occurred"
}
```

## Orders

### GET /v3/users/{userId}/orders

Returns all orders placed by a specific user with optional filtering by order status

#### Path Parameters

- `userId` *(string, required)* Unique identifier of the user whose orders to retrieve

#### Query Parameters

- `status` *(string)* Filter orders by their current status (placed, approved, or delivered)

### Responses

#### 200 Response

Successful response with user's orders

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

- `orders` *(array of Order)* Array of order objects matching the query criteria
- `cursor` *(string)* Pagination cursor for fetching the next page of results

**Order**
- `id` *(string)*: Unique identifier for the order
- `userId` *(string)*: Identifier of the user who placed the order
- `petId` *(string)*: Identifier of the pet being ordered
- `status` *(string)*: Current status of the order in the fulfillment process. Enums: `placed`, `approved`, `delivered`
- `quantity` *(integer)*: Number of pets being ordered

#### 404 Response

User not found

```json
{
   "code": 6,
   "error": "An error occurred"
}
```

### GET /v3/orders

Returns a paginated list of all orders in the system with cursor-based pagination support

#### Query Parameters

- `after` *(string)* Cursor for pagination - returns orders after this cursor

- `first` *(integer)* Number of orders to return (cursor-based pagination)

### Responses

#### 200 Response

Successful response with order list

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

- `orders` *(array of Order)* Array of order objects matching the query criteria
- `cursor` *(string)* Pagination cursor for fetching the next page of results

**Order**
- `id` *(string)*: Unique identifier for the order
- `userId` *(string)*: Identifier of the user who placed the order
- `petId` *(string)*: Identifier of the pet being ordered
- `status` *(string)*: Current status of the order in the fulfillment process. Enums: `placed`, `approved`, `delivered`
- `quantity` *(integer)*: Number of pets being ordered

### POST /v3/orders

Place a new order for pets with specified quantity and delivery details

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

- `id` *(string)* Unique identifier assigned to the newly created order
- `message` *(string)* Confirmation message about the order placement

#### 400 Response

Invalid order data or insufficient inventory

```json
{
   "code": 6,
   "error": "An error occurred"
}
```

### GET /v3/orders/{orderId}

Returns detailed information about a specific order including items, status, and delivery information

#### Path Parameters

- `orderId` *(string, required)* Unique identifier of the order to retrieve

### Responses

#### 200 Response

Successful response with order details

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

- `id` *(string)* Unique identifier for the order
- `userId` *(string)* Identifier of the user who placed the order
- `petId` *(string)* Identifier of the pet being ordered
- `status` *(string)* Current status of the order in the fulfillment process Enums: `placed`, `approved`, `delivered`
- `quantity` *(integer)* Number of pets being ordered

#### 404 Response

Order not found

```json
{
   "code": 6,
   "error": "An error occurred"
}
```

## Pets

### GET /v3/pets

Returns a paginated list of all pets in the store with support for filtering by tags and cursor-based pagination

#### Query Parameters

- `after` *(string)* Cursor for pagination - returns pets after this cursor

- `first` *(integer)* Number of pets to return (cursor-based pagination)

- `tag` *(string)* Filter by tag

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

- `pets` *(array of Pet)* Array of pet objects matching the query criteria
- `cursor` *(string)* Pagination cursor for fetching the next page of results

**Pet**
- `id` *(string)*: Unique identifier for the pet
- `name` *(string)*: Name of the pet
- `status` *(string)*: Current availability status of the pet in the store. Enums: `available`, `pending`, `sold`
- `tags` *(string array)*: Tags for categorizing and filtering pets (e.g., species, characteristics)

#### 400 Response

Invalid request parameters

```json
{
   "code": 6,
   "error": "An error occurred"
}
```

### POST /v3/pets

Adds a new pet to the store inventory with the provided details

#### Headers

Name | Description | Required | Type
-----|-------------|----------|-----
X-Request-ID | Unique request identifier for tracking and debugging | true | string

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

- `id` *(string)* Unique identifier assigned to the newly created pet
- `message` *(string)* Confirmation message about the pet creation

#### 400 Response

Invalid pet data

```json
{
   "code": 6,
   "error": "An error occurred"
}
```

### POST /v3/pets.delete

Removes a pet from the store inventory. Requires admin privileges to perform this operation

#### Headers

Name | Description | Required | Type
-----|-------------|----------|-----
X-Admin-Token | Admin authorization token required for this operation | true | string

### Request

```json
{
   "petId": "dl2INvNSQT"
}
```

#### Field Definitions

- `petId` *(string, required)* Unique identifier of the pet to delete from inventory

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

- `pet` *(object)* Information about the deleted pet

#### 403 Response

Unauthorized - invalid or missing admin token

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

Returns detailed information about a specific pet including its status, tags, and availability

#### Path Parameters

- `petId` *(string, required)* Unique identifier of the pet to retrieve

### Responses

#### 200 Response

Successful response with pet details

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

- `id` *(string)* Unique identifier for the pet
- `name` *(string)* Name of the pet
- `status` *(string)* Current availability status of the pet in the store Enums: `available`, `pending`, `sold`
- `tags` *(string array)* Tags for categorizing and filtering pets (e.g., species, characteristics)

#### 404 Response

Pet not found

```json
{
   "code": 6,
   "error": "An error occurred"
}
```

## Users

### GET /v3/users

Returns a list of registered users with optional filtering by active status

#### Query Parameters

- `active` *(boolean)* Filter users by their active status (true for active users, false for inactive)

### Responses

#### 200 Response

Successful response with user list

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

- `users` *(array of User)* Array of user objects matching the query criteria

**User**
- `id` *(string)*: Unique identifier for the user account
- `username` *(string)*: User's chosen username for login and display
- `email` *(string)*: User's email address for communication and account recovery
- `active` *(boolean)*: Indicates whether the user account is currently active

### POST /v3/users

Register a new user account in the system with username and email

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

- `id` *(string)* Unique identifier assigned to the newly created user
- `message` *(string)* Confirmation message about the user registration

#### 400 Response

Invalid user data or duplicate email/username

```json
{
   "code": 6,
   "error": "An error occurred"
}
```

### GET /v3/users/{userId}

Returns detailed information about a specific user including their profile and account status

#### Path Parameters

- `userId` *(string, required)* Unique identifier of the user to retrieve

### Responses

#### 200 Response

Successful response with user details

```json
{
   "active": true,
   "email": "user@example.com",
   "id": "0L8r30xvTn",
   "username": "j31WE1Wf9y"
}
```

#### Field Definitions

- `id` *(string)* Unique identifier for the user account
- `username` *(string)* User's chosen username for login and display
- `email` *(string)* User's email address for communication and account recovery
- `active` *(boolean)* Indicates whether the user account is currently active

#### 404 Response

User not found

```json
{
   "code": 6,
   "error": "An error occurred"
}
```

### GET /v3/users/{userId}/orders

Returns all orders placed by a specific user with optional filtering by order status

#### Path Parameters

- `userId` *(string, required)* Unique identifier of the user whose orders to retrieve

#### Query Parameters

- `status` *(string)* Filter orders by their current status (placed, approved, or delivered)

### Responses

#### 200 Response

Successful response with user's orders

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

- `orders` *(array of Order)* Array of order objects matching the query criteria
- `cursor` *(string)* Pagination cursor for fetching the next page of results

**Order**
- `id` *(string)*: Unique identifier for the order
- `userId` *(string)*: Identifier of the user who placed the order
- `petId` *(string)*: Identifier of the pet being ordered
- `status` *(string)*: Current status of the order in the fulfillment process. Enums: `placed`, `approved`, `delivered`
- `quantity` *(integer)*: Number of pets being ordered

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

Returns the health status of the API and its dependencies for monitoring purposes

### Responses

#### 200 Response

Service is healthy and operational

```json
{
   "status": "healthy",
   "timestamp": "2024-01-15T10:30:00Z"
}
```

#### Field Definitions

- `status` *(string)* Overall health status of the API and its dependencies Enums: `healthy`, `degraded`, `down`
- `timestamp` *(string)* ISO 8601 timestamp of when the health check was performed

