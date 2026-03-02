## ADDED Requirements

### Requirement: Server supports user authentication

The Server SHALL support user authentication with username and password.

#### Scenario: Successful login
- **WHEN** user submits valid credentials
- **THEN** Server validates password hash
- **AND** Server generates JWT token
- **AND** Server returns token with expiration

#### Scenario: Failed login
- **WHEN** user submits invalid credentials
- **THEN** Server returns authentication error
- **AND** Server does not reveal which field is incorrect
- **AND** Server logs the failed attempt

#### Scenario: Account lockout
- **WHEN** user fails login 5 times in 5 minutes
- **THEN** Server locks account for 15 minutes
- **AND** Server returns account locked error

### Requirement: Server manages user sessions

The Server SHALL manage user sessions via JWT.

#### Scenario: Token validation
- **WHEN** request includes valid JWT token
- **THEN** Server validates token signature
- **AND** Server checks token expiration
- **AND** Server extracts user identity

#### Scenario: Token expiration
- **WHEN** token is expired
- **THEN** Server returns 401 Unauthorized
- **AND** Client must re-authenticate

#### Scenario: Token refresh
- **WHEN** user requests token refresh with valid token
- **THEN** Server issues new token with extended expiration

### Requirement: Server supports user roles

The Server SHALL support role-based access control (RBAC).

#### Scenario: Admin role
- **WHEN** user has `admin` role
- **THEN** user can perform all operations
- **AND** can manage users and Agents

#### Scenario: Operator role
- **WHEN** user has `operator` role
- **THEN** user can create and manage tasks
- **AND** user can view Agents and metrics
- **AND** user cannot manage users

#### Scenario: Viewer role
- **WHEN** user has `viewer` role
- **THEN** user can only view Agents and tasks
- **AND** user cannot create or modify anything

### Requirement: Server supports user management

The Server SHALL support user CRUD operations.

#### Scenario: Create user
- **WHEN** admin creates new user
- **THEN** Server validates username uniqueness
- **AND** Server hashes password
- **AND** Server assigns role
- **AND** Server stores user in database

#### Scenario: Update user
- **WHEN** admin updates user
- **THEN** Server validates permissions
- **AND** Server updates allowed fields

#### Scenario: Delete user
- **WHEN** admin deletes user
- **THEN** Server soft-deletes user
- **AND** Server invalidates active sessions

#### Scenario: List users
- **WHEN** admin requests user list
- **THEN** Server returns all users (excluding passwords)
- **AND** Server supports pagination

### Requirement: Server audits user actions

The Server SHALL log all user actions for audit.

#### Scenario: Action logging
- **WHEN** user performs any action
- **THEN** Server logs:
  - user ID
  - action type
  - resource type and ID
  - action details (JSON)
  - IP address
  - timestamp

#### Scenario: Audit query
- **WHEN** admin queries audit logs
- **THEN** Server returns logs filtered by:
  - user ID
  - action type
  - time range
- **AND** Server supports pagination

### Requirement: Server secures API endpoints

The Server SHALL secure all API endpoints.

#### Scenario: Unauthenticated access
- **WHEN** request lacks valid token
- **THEN** Server returns 401 Unauthorized
- **AND** Server allows only `/api/v1/auth/*` endpoints

#### Scenario: Unauthorized access
- **WHEN** user lacks required role
- **THEN** Server returns 403 Forbidden
- **AND** Server logs unauthorized attempt
