# Mobile UX Improvements - Requirements

## Overview

The current mobile experience needs aggressive improvements focused on iPhone 16 Plus (430px width). This is a personal app in WIP mode, so we can be aggressive with changes and keep implementation simple without complex accessibility features or legacy browser support.

## Current Problems Identified

### 1. **Dashboard Information Overload**

- Complex frequency date ranges (Jul 20 - Jul 26, Jul 13 - Jul 26, July 2025) are confusing
- Too much technical information displayed on small screens
- Current period indicator takes up too much vertical space

### 2. **Gauge Visual Design Issues**

- Gauges don't look appealing or modern
- No visual status indicators (good/bad performance)
- Missing color coding for at-a-glance status understanding
- Poor visual hierarchy and readability

### 3. **Frequency Differentiation Problems**

- No visual distinction between weekly, bi-weekly, and monthly gauges
- Users can't quickly identify gauge types
- Frequency information is buried in small text

### 4. **Layout and Scrolling Issues**

- Footer not visible without scrolling (sticky footer needed)
- Content overflows on small screens
- Poor use of available screen real estate
- Inconsistent spacing and padding

### 5. **General Mobile UX Issues**

- Buttons too small for touch interaction
- Text hierarchy unclear on small screens
- Loading states and transitions missing
- No progressive disclosure of information

## Solution Requirements

### R1. **Simplified Dashboard Header**

**Priority: High**

- **R1.1**: Replace complex date ranges with simple "Today: [Current Date]"
- **R1.2**: Remove technical frequency period displays (Jul 20-26, etc.)
- **R1.3**: Add a clean, minimal current date indicator
- **R1.4**: Reduce header height by 50% on mobile screens
- **R1.5**: Use typography hierarchy to emphasize important information

**Acceptance Criteria:**

- Dashboard header shows only "Today: Friday, July 25, 2025"
- No frequency date ranges visible
- Header takes <25% of viewport height on mobile
- Clean, readable typography with proper contrast

### R2. **Enhanced Gauge Visual Design**

**Priority: High**

- **R2.1**: Implement simple visual status indicators with color coding
  - Green: On track / meeting targets
  - Red: Off track / missing targets  
  - No yellow/warning states - keep it simple
- **R2.2**: Add progress visualization (progress bars or radial indicators)
- **R2.3**: Use modern card design with subtle shadows and borders
- **R2.4**: Implement clear visual hierarchy with larger numbers, smaller labels
- **R2.5**: Add icons or visual elements to make gauges more engaging

**Acceptance Criteria:**

- Immediate visual status recognition (color coding)
- Progress towards target clearly visible
- Modern, card-based design with proper spacing
- Touch-friendly interaction areas (44px minimum)
- Consistent visual design language

### R3. **Frequency Visual Differentiation**

**Priority: Medium**

- **R3.1**: Use simple colored badges for different frequencies:
  - Weekly: Blue badge
  - Bi-weekly: Purple badge
  - Monthly: Green badge
- **R3.2**: No complex icons - just colored text badges

**Acceptance Criteria:**

- Immediate frequency identification through visual cues
- Consistent color coding across the application
- Clear grouping and categorization
- Accessible color combinations (WCAG AA compliance)

### R4. **Mobile Layout Optimization**

**Priority: High**

- **R4.1**: Optimize for iPhone 16 Plus (430px width) - single column layout
- **R4.2**: Add sticky footer that remains visible
- **R4.3**: Larger touch targets (56px buttons) for better mobile interaction
- **R4.4**: Remove card fixed sizing, make fully responsive

**Acceptance Criteria:**

- Optimized for iPhone 16 Plus (430px width)
- Footer always visible and accessible
- 56px touch targets for buttons
- Content fits within viewport without overflow

### R5. **Touch Optimization**

**Priority: Medium**

- **R5.1**: 56px circular buttons for increment/decrement
- **R5.2**: Better visual feedback on button press
- **R5.3**: Immediate response to touch interactions

**Acceptance Criteria:**

- All buttons easily tappable on iPhone 16 Plus
- Clear visual feedback when buttons are pressed

## Technical Implementation Notes

### Simple Approach
- Direct modifications to existing templ components  
- CSS-first approach using DaisyUI utilities
- No complex animations or accessibility features
- Target iPhone 16 Plus specifically (430px width)

### Testing Requirements
- Test primarily on iPhone 16 Plus (430px width)
- Verify 56px touch targets work well
- Check footer stickiness

## Implementation Priority

### Phase 1 (High Priority)
- R1: Simplified Dashboard Header ✅ (Done)
- R2: Enhanced Gauge Visual Design (Green/Red status)
- R4: Mobile Layout Optimization (iPhone 16 Plus)

### Phase 2 (Medium Priority)
- R3: Frequency Visual Differentiation (Simple badges)
- R5: Touch Optimization (56px buttons)

