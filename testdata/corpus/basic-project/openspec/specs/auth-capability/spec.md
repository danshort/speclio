## Purpose

Authentication capability.

## Requirements

### Requirement: Login
Users SHALL log in.

#### Scenario: Valid credentials
- **WHEN** correct credentials are supplied
- **THEN** the user is authenticated

### Requirement: Logout
Users SHALL log out.

#### Scenario: Active session
- **WHEN** a logged-in user logs out
- **THEN** the session ends
