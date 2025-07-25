# Mobile UX Design Guidelines - Simplified for iPhone 16 Plus

## Visual Design Specifications

### Simple Color System for Status Indicators

```css
/* Status Colors - Simple Green/Red System */
:root {
  /* Success States */
  --status-success: #22c55e;      /* Bright Green - Meeting target */
  --status-success-border: #16a34a; /* Darker green for borders */
  
  /* Error States */  
  --status-error: #ef4444;        /* Bright Red - Not meeting target */
  --status-error-border: #dc2626; /* Darker red for borders */
  
  /* Frequency Colors */
  --freq-weekly: #3b82f6;         /* Blue */
  --freq-biweekly: #8b5cf6;       /* Purple */
  --freq-monthly: #059669;        /* Green */
}
```

### Simple Frequency Badges

#### Weekly Gauges
- **Badge**: "Weekly" with blue background (`#3b82f6`)
- **Text**: White text on blue background

#### Bi-weekly Gauges  
- **Badge**: "Bi-weekly" with purple background (`#8b5cf6`)
- **Text**: White text on purple background

#### Monthly Gauges
- **Badge**: "Monthly" with green background (`#059669`)
- **Text**: White text on green background

### iPhone 16 Plus Layout Specifications

#### Target Screen Size
```css
/* iPhone 16 Plus - 430px width */
.iphone-16-plus { width: 430px; }
```

#### Simplified Header Layout
```
┌─────────────────────────────────────┐
│          Today: Friday, July 25      │ ← Simple, clean header
└─────────────────────────────────────┘
Total: ~60px height (simplified)
```

#### Responsive Gauge Card Layout  
```
┌─────────────────────────────────────┐
│ 🏃‍♂️ [Blue Badge: Weekly]        [⚙️] │ ← Better icon + frequency badge
├─────────────────────────────────────┤
│               47                    │ ← Large number
│              hours                  │ ← Unit text
│    ████████░░ 82% to target        │ ← Progress bar + status border
│         [➖]    [➕]                 │ ← 56px buttons
└─────────────────────────────────────┘
iPhone 16 Plus: Full width (430px)
Desktop: Multi-column responsive grid
```

### Simple Touch Target Specifications

```css
/* 56px Touch Targets for iPhone 16 Plus */
.gauge-button {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
}

/* Status Border Styling */
.gauge-success-border {
  border-left: 4px solid var(--status-success-border);
}

.gauge-error-border {
  border-left: 4px solid var(--status-error-border);
}
```

## Simple Implementation Approach

### 1. Status Calculation (Green/Red Only)
```go
func getGaugeStatus(value, target int64, direction string) string {
  if direction == "under" {
    if value <= target {
      return "success" // Green
    }
    return "error" // Red
  } else {
    if value >= target {
      return "success" // Green  
    }
    return "error" // Red
  }
}
```

### 2. Responsive Grid for iPhone 16 Plus and Desktop
```css
/* Single column layout for iPhone 16 Plus */
@media (max-width: 430px) {
  .gauge-grid {
    grid-template-columns: 1fr;
    gap: 1rem;
    padding: 1rem;
  }
  
  .gauge-card {
    width: 100%;
    min-height: auto; /* Remove fixed sizing */
  }
}

/* Multi-column layout for desktop */
@media (min-width: 768px) {
  .gauge-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 1.5rem;
  }
}

@media (min-width: 1024px) {
  .gauge-grid {
    grid-template-columns: repeat(3, 1fr);
    gap: 2rem;
  }
}
```

### 3. Sticky Footer
```css
/* Simple sticky footer */
.drawer-content {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

.container {
  flex-grow: 1;
}

footer {
  margin-top: auto;
}
```

### 4. Better Gauge Icons
```css
/* Modern gauge icons with better visual design */
.gauge-icon {
  font-size: 2rem; /* 32px - larger, more prominent */
  filter: drop-shadow(0 1px 2px rgba(0, 0, 0, 0.1));
}

/* Icon suggestions for common gauge types */
.exercise-icon { content: "🏃‍♂️"; } /* Running person */
.sleep-icon { content: "😴"; }     /* Sleeping face */
.water-icon { content: "💧"; }     /* Water drop */
.food-icon { content: "🍎"; }      /* Apple/healthy food */
.weight-icon { content: "⚖️"; }    /* Scale */
.steps-icon { content: "👟"; }     /* Shoe */
.meditation-icon { content: "🧘‍♀️"; } /* Meditation */
.reading-icon { content: "📚"; }   /* Books */
```

This simplified approach focuses on immediate visual impact for iPhone 16 Plus and desktop without complex features.