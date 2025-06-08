# Frontend Layout Documentation

> This document describes the root layout component that provides the consistent application shell, including navigation, metadata, and footer for the NIVAI frontend.

## Architecture

```mermaid
classDiagram
    class RootLayout {
        +Metadata metadata
        +ReactNode children
        +render()
    }

    class Navigation {
        +Link[] navItems
        +render()
    }

    class Footer {
        +Link[] quickLinks
        +Link[] legalLinks
        +render()
    }

    class ThemeProvider {
        +Font inter
        +String className
    }

    RootLayout --> Navigation : contains
    RootLayout --> Footer : contains
    RootLayout --> ThemeProvider : uses
```

## Component Structure

```mermaid
graph TB
    subgraph RootLayout["RootLayout (layout.tsx)"]
        HTML[HTML Element]
        Body[Body Element]
        Header[Header Section]
        Main[Main Content]
        Footer[Footer Section]

        subgraph Header
            Nav[Navigation]
            Logo[NIVAI Logo]
            Upload[Upload Button]
            Profile[Profile Button]
        end

        subgraph Footer
            About[About Section]
            Quick[Quick Links]
            Legal[Legal Links]
            Copyright[Copyright Notice]
        end
    end

    HTML --> Body
    Body --> Header
    Body --> Main
    Body --> Footer

    classDef section fill:#e1f5fe,stroke:#4fc3f7,stroke-width:2px;
    classDef component fill:#f3e5f5,stroke:#ab47bc,stroke-width:2px;

    class Header,Footer section;
    class Nav,Logo,Upload,Profile,About,Quick,Legal,Copyright component;
```

## Navigation Schema

```mermaid
graph LR
    subgraph MainNav[Main Navigation]
        Dashboard["/dashboard"]
        Matches["/matches"]
        Teams["/teams"]
        Players["/players"]
        Analytics["/analytics"]
    end

    subgraph FooterNav[Footer Navigation]
        About["/about"]
        Contact["/contact"]
        Support["/support"]
        Privacy["/privacy"]
        Terms["/terms"]
    end

    classDef primary fill:#e3f2fd,stroke:#1e88e5,stroke-width:2px;
    classDef secondary fill:#f3e5f5,stroke:#8e24aa,stroke-width:2px;

    class Dashboard,Matches,Teams,Players,Analytics primary;
    class About,Contact,Support,Privacy,Terms secondary;
```

## Configuration

### Metadata

```typescript
export const metadata: Metadata = {
  title: "NIVAI - Football Analytics Platform",
  description:
    "Advanced football tracking data visualization and analysis platform",
  applicationName: "NIVAI Football Analytics",
};
```

### Font Configuration

```typescript
const inter = Inter({ subsets: ["latin"] });
```

## Layout Components

### Header Navigation

- Main navigation menu
- Upload action button
- User profile access
- Responsive design

### Footer Structure

1. **Company Information**

   - Platform description
   - Contact details

2. **Quick Links**

   - About Us
   - Contact
   - Support

3. **Legal Information**
   - Privacy Policy
   - Terms of Service
   - Copyright notice

## Styling

### Theme Colors

```css
Primary:
- Blue-800: #1E40AF (Navigation, buttons)
- Gray-700: #374151 (Text)
- White: #FFFFFF (Background)

Hover States:
- Blue-700: #1D4ED8 (Button hover)
- Blue-800: #1E40AF (Link hover)
```

### Layout Classes

```css
Container:
- max-w-7xl
- mx-auto
- px-4 sm:px-6 lg:px-8

Spacing:
- py-12 (Footer padding)
- space-y-2 (Link spacing)
- gap-8 (Grid gaps)
```

## Responsiveness

### Breakpoints

```css
sm: 640px  // Small devices
md: 768px  // Medium devices
lg: 1024px // Large devices
xl: 1280px // Extra large devices
```

## Usage Example

```tsx
// Page component using the layout
export default function HomePage() {
  return (
    // Layout automatically wraps the page content
    <div>
      <h1>Welcome to NIVAI</h1>
      <p>Your football analytics platform</p>
    </div>
  );
}
```

## Related Files

- `globals.css`: Global styles
- `page.tsx`: Home page component
- `components/Navigation.tsx`: Navigation component
- `components/Footer.tsx`: Footer component
