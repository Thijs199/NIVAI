# Dashboard Page Documentation

> This document describes the main analytics dashboard that serves as the central hub for football performance insights and match analysis in the AIFAA platform.

## Architecture

```mermaid
classDiagram
    class Dashboard {
        +render()
    }

    class StatCards {
        +TotalMatches
        +AverageGoals
        +PossessionAverage
        +TopSpeedRecord
    }

    class FormationAnalysis {
        +InteractivePitch
        +FormationVisualization
    }

    class PlayerPerformance {
        +PlayerList[]
        +PerformanceMetrics
    }

    class MatchAnalysis {
        +RecentMatches[]
        +MatchStats
    }

    class VideoHighlights {
        +MainVideo
        +Thumbnails[]
    }

    Dashboard --> StatCards : contains
    Dashboard --> FormationAnalysis : contains
    Dashboard --> PlayerPerformance : contains
    Dashboard --> MatchAnalysis : contains
    Dashboard --> VideoHighlights : contains
```

## Component Layout

```mermaid
graph TB
    subgraph Dashboard["Dashboard Layout"]
        Header["Season Selector"]

        subgraph Stats["Statistics Section"]
            S1["Total Matches"]
            S2["Average Goals"]
            S3["Possession Average"]
            S4["Top Speed Record"]
        end

        subgraph Analysis["Analysis Section"]
            direction LR
            F["Formation Analysis"]
            P["Player Performance"]
        end

        subgraph Data["Match Data Section"]
            direction LR
            M["Recent Matches"]
            V["Video Highlights"]
        end
    end

    Header --> Stats
    Stats --> Analysis
    Analysis --> Data

    classDef section fill:#e1f5fe,stroke:#4fc3f7,stroke-width:2px;
    classDef widget fill:#f3e5f5,stroke:#ab47bc,stroke-width:2px;

    class Header,Stats,Analysis,Data section;
    class S1,S2,S3,S4,F,P,M,V widget;
```

## Data Flow

```mermaid
sequenceDiagram
    participant U as User
    participant D as Dashboard
    participant A as API
    participant WS as WebSocket

    U->>D: Select Season
    D->>A: Fetch Season Stats
    A-->>D: Return Statistics

    U->>D: View Formation
    D->>A: Get Formation Data
    A-->>D: Return Position Data

    loop Real-time Updates
        WS->>D: Match Events
        D->>D: Update Stats
    end
```

## Components

### 1. Statistics Cards

- Total Matches Counter
- Average Goals per Match
- Team Possession Statistics
- Player Speed Records

### 2. Interactive Formation Analysis

```typescript
interface FormationData {
  positions: PlayerPosition[];
  heatmap: HeatmapData;
  movements: MovementPattern[];
}
```

### 3. Player Performance Tracker

```typescript
interface PlayerMetrics {
  name: string;
  team: string;
  rating: number;
  performance: number; // 0-100
}
```

## State Management

### Season Selection

```typescript
type Season = {
  id: string;
  name: string;
  startDate: Date;
  endDate: Date;
  isActive: boolean;
};
```

### Match Status

```typescript
type MatchStatus = "complete" | "processing" | "pending";
```

## Styling

### Card Components

```css
baseCard: .card {
  @apply bg-white overflow-hidden shadow rounded-lg;
}

statsCard: .stats-card {
  @apply px-4 py-5 sm:p-6;
}

metricDisplay: .metric {
  @apply mt-1 text-3xl font-semibold text-gray-900;
}
```

## Interactive Features

### 1. Formation Visualization

- Pixi.js Canvas Integration
- Player Position Dragging
- Formation Pattern Analysis

### 2. Video Controls

- Playback Controls
- Thumbnail Navigation
- Picture-in-Picture Support

## Performance Optimizations

1. **Data Loading**

   - Lazy video loading
   - Progressive stat updates
   - Cached formation data

2. **Rendering**
   - Virtualized player lists
   - Optimized canvas updates
   - Debounced resize handlers

## Usage Example

```typescript
// Dashboard container with real-time updates
import { useWebSocket } from "@/hooks/useWebSocket";

function DashboardContainer() {
  const { data, isConnected } = useWebSocket("/ws/matches");

  return <Dashboard realtimeData={data} isLive={isConnected} />;
}
```

## Related Files

- `components/StatCard.tsx`: Reusable stat card component
- `components/FormationView.tsx`: Interactive pitch component
- `hooks/useMatchData.ts`: Match data fetching hook
- `services/api.ts`: API service integration
