# Note-Sharing App
# API Overview

This API powers a note-taking application, offering endpoints for user authentication and note management. It uses JSON for data exchange, ensuring lightweight and structured communication between client and server. Authentication is secured with JWT (JSON Web Tokens), providing a robust mechanism for protecting routes and managing user sessions efficiently.

## Users and Auth

Handles user registration, login, and session management with the following endpoints:

- **`/api/v1/register` (POST)**: Allows new users to sign up by submitting an email and password, creating a secure account in the system.
- **`/api/v1/login` (POST)**: Authenticates users with their credentials, returning access and refresh tokens for secure session management.
- **`/api/v1/logout` (POST)**: Invalidates the user's refresh token, ensuring a clean and secure logout process.
- **`/api/v1/token/refresh` (POST)**: Refreshes an expired access token using a valid refresh token, keeping users logged in seamlessly.
- **`/api/v1/user/me` (PUT)**: Lets authenticated users update their email or password, maintaining control over their account details.

## Notes

Manages CRUD (Create, Read, Update, Delete) operations for notes, tied to authenticated users where specified:

- **`/api/v1/notes/{noteID}` (PUT)**: Updates the content of an existing note, allowing the authenticated owner to modify it as needed.
- **`/api/v1/notes/{noteID}` (DELETE)**: Removes a specific note from the system, restricted to the authenticated owner for security.
- **`/api/v1/notes/{noteID}` (GET)**: Retrieves a single note by its ID, accessible only to the authenticated owner.
- **`/api/v1/notes` (GET)**: Fetches all notes associated with a specified author ID, publicly accessible without authentication for flexibility.
- **`/api/v1/notes` (POST)**: Creates a new note for the authenticated user, enabling them to build their collection.

**Note**: This documentation is actively evolving and will be refined as the project progresses.
## Database Schema
<img src="./doc_assets/db_diagram.png" alt="Database Diagram" height ="70%" width="70%">
