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

Delete a pet

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

##### 404 Response

Pet not found

## orders

### GET /v3/users/{userId}/orders

Get user orders

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

##### 404 Response

User not found

### GET /v3/orders

List all orders

Returns a list of all orders in the system

#### Query Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
after | Cursor for pagination - returns orders after this cursor | false | string
first | Number of orders to return (cursor-based pagination) | false | integer

##### 200 Response

Successful response

### POST /v3/orders

Create a new order

Place a new order for pets

##### 201 Response

Order created successfully

##### 400 Response

Invalid order data

### GET /v3/orders/{orderId}

Get order by ID

Returns detailed order information

#### Path Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
orderId | Order identifier | true | string

##### 200 Response

Successful response

##### 404 Response

Order not found

## pets

### GET /v3/pets

List all pets

Returns a paginated list of all pets in the store

#### Query Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
after | Cursor for pagination - returns pets after this cursor | false | string
first | Number of pets to return (cursor-based pagination) | false | integer
tag | Filter by tag | false | string

##### 200 Response

Successful response with pet list

##### 400 Response

Invalid request parameters

### POST /v3/pets

Create a new pet

Adds a new pet to the store inventory

#### Header Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
X-Request-ID | Unique request identifier | true | string

##### 201 Response

Pet created successfully

##### 400 Response

Invalid pet data

### POST /v3/pets.delete

Delete a pet

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

##### 404 Response

Pet not found

### GET /v3/pets/{petId}

Get a pet by ID

Returns detailed information about a specific pet

#### Path Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
petId | ID of pet to return | true | string

##### 200 Response

Successful response

##### 404 Response

Pet not found

## users

### GET /v3/users

List all users

Returns a list of registered users

#### Query Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
active | Filter by active status | false | boolean

##### 200 Response

Successful response

### POST /v3/users

Create a new user

Register a new user account

##### 201 Response

User created successfully

##### 400 Response

Invalid user data

### GET /v3/users/{userId}

Get user by ID

Returns detailed user information

#### Path Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
userId | User identifier | true | string

##### 200 Response

Successful response

##### 404 Response

User not found

### GET /v3/users/{userId}/orders

Get user orders

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

##### 404 Response

User not found

## Default APIs

### GET /v3/health

Health check endpoint

Returns the health status of the API

##### 200 Response

Service is healthy

### GET /v3/metrics

Get API metrics

Returns usage metrics and statistics

#### Header Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
X-Admin-Token | Admin authorization token | true | string

##### 200 Response

Metrics data

##### 403 Response

Unauthorized

