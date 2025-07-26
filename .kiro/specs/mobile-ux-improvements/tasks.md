# Mobile UX Improvements - Implementation Tasks

## Implementation Plan

- [x] 1. Simplify Dashboard Header
  - Remove complex frequency date ranges from dashboard header component
  - Replace with simple "Today: [Current Date]" display
  - Reduce header height by 50% for mobile screens
  - Update dashboard.templ to show clean date format
  - _Requirements: R1.1, R1.2, R1.3, R1.4_

- [x] 2. Implement Gauge Status Visual Indicators
  - Add status calculation function for green/red status determination
  - Implement colored left border (4px) on gauge cards based on status
  - Add progress bar component showing percentage to target
  - Update gauge.templ with visual status indicators
  - _Requirements: R2.1, R2.2_

- [x] 3. Optimize Layout for iPhone 16 Plus
  - Remove fixed card sizing (16rem x 16rem) from gauge components
  - Implement responsive card layout with single column on 430px width
  - Update CSS for multi-column layout on desktop screens
  - Improve card spacing for touch interaction
  - Fix HTMX button targets and functionality
  - _Requirements: R4.1, R4.4_

- [x] 4. Add Frequency Visual Differentiation
  - Implement colored badge system for frequency types
  - Add Weekly (blue), Bi-weekly (purple), Monthly (green) badges
  - Update gauge components to display frequency badges
  - Add "NEW" badge for gauges created within 24 hours
  - _Requirements: R3.1, R3.2_

- [ ] 5. Implement Sticky Footer
  - Modify layout structure to support sticky footer
  - Use CSS flexbox approach for footer positioning
  - Ensure footer remains visible at bottom of viewport
  - Update base layout template with proper CSS classes
  - _Requirements: R4.2_

- [ ] 6. Optimize Touch Targets for Mobile
  - Increase increment/decrement buttons to 56px circular targets
  - Improve button spacing and visual feedback on touch
  - Ensure all interactive elements meet 56px minimum touch target
  - Test touch interactions on iPhone 16 Plus (430px width)
  - _Requirements: R5.1, R5.2, R5.3_

## Key Implementation Files

- `internal/views/pages/dashboard.templ` - Main dashboard layout and header
- `internal/views/components/gauge.templ` - Individual gauge card components
- `internal/views/components/gauge_list.templ` - Gauge list layout
- `internal/views/layouts/` - Base layout templates for sticky footer
- `cmd/server/web/static/style.css` - Custom CSS for mobile optimizations

## Testing Strategy

Each task should include:
- Unit tests for any new Go functions or logic
- Component tests for template changes
- Integration tests to verify end-to-end functionality
- Manual testing on iPhone 16 Plus (430px width)

**Test Command:**
```bash
make test  # Runs all test suites
```

## Mobile Testing Checklist

- [ ] Test responsive layout on iPhone 16 Plus (430px width)
- [ ] Verify 56px touch targets are easily tappable
- [ ] Check sticky footer behavior across different content heights
- [ ] Validate color-coded status indicators are clearly visible
- [ ] Test card layout responsiveness from mobile to desktop
- [ ] Ensure all HTMX interactions work properly on touch devices
