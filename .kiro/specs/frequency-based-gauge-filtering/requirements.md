# Requirements Document

## Introduction

This feature enhances the dashboard to display only gauges that are relevant to the current time period based on their frequency settings. Instead of showing all gauges regardless of their timing, the dashboard will intelligently filter and show only gauges that are actionable or relevant for the current time period (weekly, bi-weekly, monthly). Additionally, gauges that are not yet relevant or have passed their time window will be visually disabled to provide clear user feedback.

**Project Context:** This is a work-in-progress application that is not yet deployed to production. Therefore, we can be aggressive with database schema changes, tooling improvements, and structural modifications without concern for backward compatibility or migration complexity.

## Requirements

### Requirement 1

**User Story:** As a user, I want to see only gauges relevant to the current time period on my dashboard, so that I can focus on actionable items without being distracted by irrelevant gauges.

#### Acceptance Criteria

1. WHEN the dashboard loads THEN the system SHALL display only gauges that are relevant to the current time period based on their frequency
2. WHEN a gauge has weekly frequency THEN the system SHALL show it only during its designated week
3. WHEN a gauge has bi-weekly frequency THEN the system SHALL show it only during its designated bi-weekly period based on ISO calendar weeks starting on Sunday (e.g., weeks 1-2 form one bi-weekly period, weeks 3-4 form another, with each week running Sunday through Saturday)
4. WHEN a gauge has monthly frequency THEN the system SHALL show it only during its designated month
5. WHEN a gauge is not relevant to the current time period THEN the system SHALL either hide it or display it in a disabled state

### Requirement 2

**User Story:** As a user, I want to configure my gauges with frequency and activation settings, so that the automated system can manage them properly.

#### Acceptance Criteria

1. WHEN creating a gauge THEN the system SHALL require frequency to be specified (weekly, bi-weekly, or monthly)
2. WHEN creating a gauge THEN the system SHALL allow me to set it as active or inactive
3. WHEN a gauge is set as active THEN the automated process SHALL create instances based on its frequency
4. WHEN creating a gauge THEN the system SHALL set the start date to determine when the automated process begins
5. WHEN editing a gauge THEN the system SHALL allow me to change its active status

### Requirement 3

**User Story:** As a system administrator, I want an automated process to manage gauge schedules, so that new gauge instances are created automatically based on their frequency settings.

#### Acceptance Criteria

1. WHEN a gauge's time period is about to end THEN the system SHALL automatically create a new instance for the next period
2. WHEN creating a new gauge instance THEN the system SHALL reset its value to 0 while preserving the configuration
3. WHEN a gauge is set as active by the user THEN the automated process SHALL create instances as needed based on frequency
4. IF a gauge instance already exists for a time period THEN the system SHALL NOT create a duplicate
5. WHEN the automated process runs THEN the system SHALL optionally log the creation of new gauge instances

### Requirement 4

**User Story:** As a user, I want to see only active gauges relevant to the current time period on my dashboard, so that I can focus on actionable items.

#### Acceptance Criteria

1. WHEN viewing the dashboard THEN the system SHALL display only gauge instances that are active for the current time period
2. WHEN a gauge instance is not relevant to the current time period THEN the system SHALL hide it from the dashboard
3. WHEN viewing the dashboard THEN the system SHALL show gauge instances with their current values and progress
4. WHEN no gauge instances are active for the current period THEN the system SHALL display a message indicating no active gauges
5. WHEN viewing the dashboard THEN the system SHALL provide a simple indicator showing the current time period

### Requirement 5

**User Story:** As a developer, I want to improve the project's tooling and structure to better support the new frequency-based features, so that development and maintenance are more efficient.

#### Acceptance Criteria

1. WHEN implementing database changes THEN the system SHALL migrate from Atlas to Goose for database migrations to provide better Go integration and simpler migration management
2. WHEN setting up the new database schema THEN the system SHALL use Goose migration files for all schema changes including the new frequency-based fields
3. WHEN running database migrations THEN the system SHALL support both up and down migrations for easy rollback during development
4. WHEN managing database versions THEN the system SHALL use Goose's versioning system to track migration state
5. WHEN developing new features THEN the system SHALL maintain the existing SQLC integration for type-safe database queries

### Requirement 6

**User Story:** As a user, I want to view historical gauge data across different time periods, so that I can track my progress over time.

#### Acceptance Criteria

1. WHEN viewing gauge history THEN the system SHALL group data by time periods based on frequency
2. WHEN a gauge has weekly frequency THEN the system SHALL show weekly historical data
3. WHEN a gauge has bi-weekly frequency THEN the system SHALL show bi-weekly historical data
4. WHEN a gauge has monthly frequency THEN the system SHALL show monthly historical data
5. WHEN viewing trends THEN the system SHALL maintain the ability to see data across multiple time periods