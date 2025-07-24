# Implementation Plan

## Development Workflow

**Recommended approach: Feature branches with pull requests**

1. Create feature branch for each task (e.g., `feature/goose-migration-setup`)
2. Implement the task including any unit tests
3. Run tests to ensure functionality works
4. Create pull request with clear description
5. Review and merge to main branch
6. Move to next task

**Task Dependencies:**
- Tasks 1-3 are foundational and should be completed first
- Unit test tasks (X.1) can be included in the same branch as their parent task
- Each task builds on previous ones, so complete in order

- [x] 1. Set up Goose migration tooling and replace Atlas
  - Remove Atlas configuration files (atlas.hcl)
  - Install and configure Goose for database migrations
  - Create initial migration structure with embedded migrations
  - Update database initialization code to use Goose
  - _Requirements: 5.1, 5.2, 5.3, 5.4_

- [x] 2. Create new database schema with separate tables
  - Write Goose migration to create gauge_templates table
  - Write Goose migration to create gauge_instances table  
  - Update gauge_values table to reference gauge_instances
  - Create necessary indexes for performance
  - _Requirements: 2.1, 3.2_

- [x] 3. Update SQLC queries for new table structure
  - Create queries for gauge_templates CRUD operations
  - Create queries for gauge_instances CRUD operations
  - Update existing gauge queries to work with new schema
  - Generate new Go code with SQLC
  - _Requirements: 5.5_

- [x] 4. Implement period calculation utilities
  - Create time utility functions for current period calculation
  - Create time utility functions for next period calculation
  - Handle edge cases (month boundaries, leap years)
  - _Requirements: 3.1, 3.3_

- [x] 4.1 Write unit tests for period calculation utilities
  - Test weekly period calculations for all days of week
  - Test bi-weekly period calculations for different dates
  - Test monthly period calculations including month boundaries
  - Test edge cases like leap years and year boundaries
  - _Requirements: 3.1, 3.3_

- [x] 5. Create scheduling service for automated gauge instance creation
  - Implement SchedulingService interface with CreateInstancesForActiveTemplates method
  - Add logic to check if next period instances already exist
  - Add logic to create new gauge instances for next periods
  - _Requirements: 3.1, 3.2, 3.3, 3.4_

- [x] 5.1 Write unit tests for scheduling service
  - Test CreateInstancesForActiveTemplates with mock database
  - Test duplicate instance prevention logic
  - Test instance creation for different frequencies
  - Test error handling when database operations fail
  - _Requirements: 3.1, 3.2, 3.3, 3.4_

- [x] 6. Update gauge handlers to work with template/instance model
  - Modify gauge creation handler to create gauge templates
  - Update gauge form validation to include active field
  - Modify gauge editing handler to work with templates
  - Update admin page to show gauge templates instead of instances
  - _Requirements: 2.1, 2.2, 2.5_

- [x] 6.1 Write unit tests for gauge template handlers
  - Test gauge template creation with valid data
  - Test form validation for required fields (frequency, active status)
  - Test gauge template editing and updates
  - Test error handling for invalid template data
  - _Requirements: 2.1, 2.2, 2.5_

- [x] 6.2 Create database seeding command for development
  - Create seed command to populate database with sample gauge templates
  - Create sample gauge instances for current periods
  - Add Makefile target for easy seeding (`make seed`)
  - Include variety of frequencies and realistic sample data
  - _Development tool for easier testing and demo purposes_

- [x] 7. Implement dashboard filtering for current period gauges
  - Update dashboard handler to query current period gauge instances
  - Implement filtering logic based on current time and frequency
  - Update dashboard template to show only current period gauges
  - Add "no active gauges" message when no instances exist
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 4.1, 4.2, 4.4_

- [x] 7.1 Write unit tests for dashboard filtering logic
  - Test dashboard shows only current period gauge instances
  - Test filtering works correctly for weekly, bi-weekly, and monthly frequencies
  - Test "no active gauges" message displays when no instances exist
  - Test dashboard handler error handling
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 4.1, 4.2, 4.4_

- [x] 8. Add background scheduling service to main application
  - Integrate scheduling service into main server startup
  - Configure scheduling service to run as background goroutine
  - Set appropriate interval for scheduling checks (daily)
  - Add graceful shutdown handling for background service
  - _Requirements: 3.1, 3.5_

- [ ] 9. Update gauge increment/decrement handlers for instances
  - Modify increment handler to work with gauge instances
  - Modify decrement handler to work with gauge instances
  - Update HTMX responses to work with new data structure
  - Ensure gauge values are properly updated in instances
  - _Requirements: 1.1, 4.3_

- [ ] 9.1 Write unit tests for gauge instance value handlers
  - Test increment handler updates gauge instance values correctly
  - Test decrement handler prevents negative values
  - Test HTMX responses return correct updated gauge data
  - Test error handling for invalid gauge instance IDs
  - _Requirements: 1.1, 4.3_

- [ ] 10. Add current period indicator to dashboard
  - Create component to show current time period context
  - Display current week/bi-weekly/monthly period information
  - Add simple styling to make period context clear
  - _Requirements: 4.5_

- [ ] 11. Update historical data queries for new structure
  - Modify gauge history queries to work with instances
  - Group historical data by time periods based on frequency
  - Update trends page to show data from gauge instances
  - Ensure historical data displays correctly for all frequencies
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

- [ ] 11.1 Write unit tests for historical data functionality
  - Test gauge history queries return correct data for instances
  - Test data grouping by time periods for each frequency type
  - Test trends page displays historical data correctly
  - Test edge cases with missing or incomplete historical data
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

- [ ] 12. Write integration tests for complete workflow
  - Test gauge template creation and activation
  - Test automated instance creation by scheduling service
  - Test dashboard filtering shows correct instances
  - Test gauge value updates work with instances
  - _Requirements: All requirements integration testing_