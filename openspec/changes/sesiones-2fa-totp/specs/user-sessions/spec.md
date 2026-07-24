## ADDED Requirements

### Requirement: Session Management
The system SHALL track active sessions for each user, recording metadata such as user agent, IP approximate location, and last activity time. The system SHALL allow users to list their own active sessions, and to terminate them individually or all at once (optionally keeping the current session active).

#### Scenario: List active sessions
- **WHEN** user makes `GET /sessions` with valid auth
- **THEN** system returns list of active sessions with metadata

#### Scenario: Terminate specific session
- **WHEN** user makes `DELETE /sessions/:id` for a valid session they own
- **THEN** the session is marked as revoked/deleted and subsequent requests using that session token return 401

#### Scenario: Terminate all except current
- **WHEN** user makes `DELETE /sessions` with `except_current=true`
- **THEN** all other active sessions for that user are revoked, while the current one remains active

#### Scenario: Session expired edge case
- **WHEN** a session has exceeded its maximum idle time or absolute TTL
- **THEN** the system SHALL reject it on auth check, and it SHALL NOT be listed as an active session in `GET /sessions`
