## ADDED Requirements

### Requirement: TOTP Setup and Enrolment
The system SHALL provide endpoints for a user to enrol in 2FA via TOTP. The system SHALL generate a secret, provide a QR-compatible URI, and require a valid code to finalize enrolment.

#### Scenario: Enrol MFA
- **WHEN** user makes `POST /auth/mfa/enroll`
- **THEN** system generates a pending TOTP secret and returns it (with QR-compatible URI)

#### Scenario: Verify and activate MFA
- **WHEN** user makes `POST /auth/mfa/verify` with correct 6-digit code for pending secret
- **THEN** MFA becomes active for the user account

#### Scenario: Failed verify attempt
- **WHEN** user provides incorrect 6-digit code during verify
- **THEN** system returns 400 and MFA remains inactive

### Requirement: TOTP Deactivation
The system SHALL allow users to disable 2FA after providing their current password.

#### Scenario: Disable MFA
- **WHEN** user makes `POST /auth/mfa/disable` providing their current valid password
- **THEN** MFA is disabled for the user
