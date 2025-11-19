# Pet Store API

A comprehensive API for managing a pet store with users, pets, and orders

## Table of Contents

HTTP Request | Description
-------------|------------
GET [/pets](#getpets) | List all pets
POST [/pets](#postpets) | Create a new pet
GET [/pets/{petId}](#getpetspetid) | Get a pet by ID
DELETE [/pets/{petId}](#deletepetspetid) | Delete a pet
GET [/users](#getusers) | List all users
POST [/users](#postusers) | Create a new user
GET [/users/{userId}](#getusersuserid) | Get user by ID
GET [/users/{userId}/orders](#getusersuseridorders) | Get user orders
GET [/orders](#getorders) | List all orders
POST [/orders](#postorders) | Create a new order
GET [/orders/{orderId}](#getordersorderid) | Get order by ID
GET [/health](#gethealth) | Health check endpoint
GET [/metrics](#getmetrics) | Get API metrics

## admin

### DELETE /pets/{petId}

Delete a pet

Removes a pet from the store inventory

#### Path Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
petId | ID of pet to delete | true | string

#### Header Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
X-Admin-Token | Admin authorization token | true | string

##### 200 Response

Pet deleted successfully

##### 403 Response

Unauthorized

##### 404 Response

Pet not found

## orders

### GET /users/{userId}/orders

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

### GET /orders

List all orders

Returns a list of all orders in the system

#### Query Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
limit | Maximum number of orders to return | false | integer

##### 200 Response

Successful response

### POST /orders

Create a new order

Place a new order for pets

##### 201 Response

Order created successfully

##### 400 Response

Invalid order data

### GET /orders/{orderId}

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

### GET /pets

List all pets

Returns a paginated list of all pets in the store

#### Query Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
limit | Maximum number of pets to return | false | integer
offset | Number of pets to skip | false | integer
tag | Filter by tag | false | string

##### 200 Response

Successful response with pet list

##### 400 Response

Invalid request parameters

### POST /pets

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

### GET /pets/{petId}

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

### DELETE /pets/{petId}

Delete a pet

Removes a pet from the store inventory

#### Path Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
petId | ID of pet to delete | true | string

#### Header Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
X-Admin-Token | Admin authorization token | true | string

##### 200 Response

Pet deleted successfully

##### 403 Response

Unauthorized

##### 404 Response

Pet not found

## users

### GET /users

List all users

Returns a list of registered users

#### Query Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
active | Filter by active status | false | boolean

##### 200 Response

Successful response

### POST /users

Create a new user

Register a new user account

##### 201 Response

User created successfully

##### 400 Response

Invalid user data

### GET /users/{userId}

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

### GET /users/{userId}/orders

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

### GET /health

Health check endpoint

Returns the health status of the API

##### 200 Response

Service is healthy

### GET /metrics

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

