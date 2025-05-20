# API Documentation
# Users and Auth
## Overview
This document outlines the "Users and Auth" API endpoints, detailing their purpose, parameters, responses, and authentication requirements. All request and response data is formatted in JSON for uniformity.

## Endpoints

### Register User
- **URL**: `/api/v1/register`
- **Method**: `POST`
- **Description**: Registers a new user by storing their email and password in the database. The password is hashed before storage, and a unique UUID is assigned to the user.
- **Parameters**:
  - **Request Body** (JSON):
    ```json
    {
      "email": "string",
      "password": "string"
    }
    ```
- **Response**:
  - **Status Codes**:
    - `201 Created`: User successfully registered.
    - `400 Bad Request`: If required fields are missing or the request body is malformed.
    - `424 Failed Dependency`: If password hashing or user creation fails.
  - **Error Responses** (JSON):
    ```json
    {"error": "[something] is required"}
    ```
    ```json
    {"error": "Invalid request body"}
    ```
    ```json
    {"error": "Failed to hash password"}
    ```
    ```json
    {"error": "Failed to create user"}
    ```
- **Authentication**: None required.

### Login User
- **URL**: `/api/v1/login`
- **Method**: `POST`
- **Description**: Authenticates a user by validating their email and password. Upon success, it generates a JWT access token and a refresh token, stores the refresh token in the database, and returns user information along with both tokens.
- **Parameters**:
  - **Request Body** (JSON):
    ```json
    {
      "email": "string",
      "password": "string"
    }
    ```
- **Response**:
  - **Status Codes**:
    - `200 OK`: Successful login.
    - `400 Bad Request`: Invalid request body.
    - `401 Unauthorized`: Invalid credentials.
    - `500 Internal Server Error`: Server-side issues.
  - **Response Body** (JSON):
    ```json
    {
      "id": "uuid",
      "created_at": "timestamp",
      "updated_at": "timestamp",
      "email": "string",
      "token": "string",
      "refresh_token": "string",
      "has_notes_premium": false
    }
    ```
- **Authentication**: None required.

### Logout User
- **URL**: `/api/v1/logout`
- **Method**: `POST`
- **Description**: Invalidates the user's refresh token by removing it from the system. The JWT access token must be provided in the request header.
- **Parameters**: None
- **Response**:
  - **Status Codes**:
    - `204 No Content`: Successfully logged out.
    - `401 Unauthorized`: Invalid or missing JWT.
    - `500 Internal Server Error`: Database failures.
- **Authentication**: Requires a valid JWT in the `Authorization` header.

### Refresh Token
- **URL**: `/api/v1/token/refresh`
- **Method**: `POST`
- **Description**: Generates a new JWT access token using a valid refresh token, which must be provided in the request header. The refresh token is verified for existence, revocation, and expiration before issuing a new access token.
- **Parameters**: None
- **Response**:
  - **Status Codes**:
    - `200 OK`: New access token generated.
    - `401 Unauthorized`: Invalid, revoked, or expired refresh token.
    - `500 Internal Server Error`: Server-side issues.
  - **Response Body** (JSON):
    ```json
    {
      "token": "string"
    }
    ```
- **Authentication**: Requires a valid refresh token in the `Authorization` header.

### Update User
- **URL**: `/api/v1/user/me`
- **Method**: `PUT`
- **Description**: Updates the authenticated user's email and password. The request must include a valid JWT in the header. The new password is hashed before storage, and updated user details are returned.
- **Parameters**:
  - **Request Body** (JSON):
    ```json
    {
      "email": "string",
      "password": "string"
    }
    ```
- **Response**:
  - **Status Codes**:
    - `200 OK`: User information updated successfully.
    - `400 Bad Request`: Invalid request body.
    - `401 Unauthorized`: Invalid or missing JWT.
    - `500 Internal Server Error`: Database issues.
  - **Response Body** (JSON):
    ```json
    {
      "id": "uuid",
      "created_at": "timestamp",
      "updated_at": "timestamp",
      "email": "string",
      "has_notes_premium": false
    }
    ```
- **Authentication**: Requires a valid JWT in the `Authorization` header.

# Notes
## Overview
This document outlines the "Notes" API endpoints, detailing their purpose, parameters, responses, and authentication requirements. All request and response data is formatted in JSON for uniformity.

## Endpoints

### Update Note
- **URL**: `/api/v1/notes/{noteID}`
- **Method**: `PUT`
- **Description**: Updates the content of a specific note after verifying ownership. Authenticates the user, validates the note ID and ownership, retrieves the existing note, replaces its content with the new body text, and updates the database record.
- **Parameters**:
  - **Path Parameters**: `noteID` (UUID)
  - **Request Body** (JSON):
    ```json
    {
      "noteID": "uuid",
      "body": "string"
    }
    ```
- **Response**:
  - **Status Codes**:
    - `204 No Content`: Note updated successfully.
- **Authentication**: Requires a valid JWT in the `Authorization` header.

### Delete Note
- **URL**: `/api/v1/notes/{noteID}`
- **Method**: `DELETE`
- **Description**: Deletes a specific note owned by the authenticated user. Validates the user's authentication token, parses the note ID from the path parameters, and removes the note from the database only if it belongs to the requesting user.
- **Parameters**:
  - **Path Parameters**: `noteID` (UUID)
  - **Request Body**: None
- **Response**:
  - **Status Codes**:
    - `204 No Content`: Note deleted successfully.
- **Authentication**: Requires a valid JWT in the `Authorization` header.

### Get Note
- **URL**: `/api/v1/notes/{noteID}`
- **Method**: `GET`
- **Description**: Retrieves a specific note by ID for the authenticated user. Validates the note ID from path parameters, authenticates the user, verifies ownership of the note, retrieves the note data from the database, and returns the complete note object.
- **Parameters**:
  - **Path Parameters**: `noteID` (UUID)
  - **Request Body**: None
- **Response**:
  - **Status Codes**:
    - `200 OK`: Note retrieved successfully.
  - **Response Body** (JSON):
    ```json
    {
      "id": "uuid",
      "created_at": "timestamp",
      "updated_at": "timestamp",
      "body": "string",
      "user_id": "uuid"
    }
    ```
- **Authentication**: Requires a valid JWT in the `Authorization` header.

### Get Notes by Author
- **URL**: `/api/v1/notes`
- **Method**: `GET`
- **Description**: Retrieves all notes for a specific author ID. Parses the author ID from query parameters, validates it's not empty, fetches all notes associated with that author from the database, and returns them as an array of note objects.
- **Parameters**:
  - **Query Parameters**: `authorId` (UUID)
  - **Request Body**: None
- **Response**:
  - **Status Codes**:
    - `200 OK`: Notes retrieved successfully.
  - **Response Body** (JSON):
    ```json
    [
      {
        "id": "uuid",
        "created_at": "timestamp",
        "updated_at": "timestamp",
        "body": "string",
        "user_id": "uuid"
      },
      ...
    ]
    ```
- **Authentication**: None required.

### Create Note
- **URL**: `/api/v1/notes`
- **Method**: `POST`
- **Description**: Creates a new note for an authenticated user. Decodes the request body, authenticates the user via token, verifies the requested user ID matches the authenticated user, creates a new note record in the database with the provided content, and assigns it to the specified user.
- **Parameters**:
  - **Request Body** (JSON):
    ```json
    {
      "body": "string",
      "user_id": "uuid"
    }
    ```
- **Response**:
  - **Status Codes**:
    - `201 Created`: Note created successfully.
- **Authentication**: Requires a valid JWT in the `Authorization` header.