## ADDED Requirements

### Requirement: Settings window
The app SHALL provide a standard macOS Settings window, opened via the "Settings…" menu item and ⌘,, as the home for user preferences. Preferences set there SHALL persist across launches.

#### Scenario: Open settings
- **WHEN** the user chooses Settings… (or presses ⌘,)
- **THEN** the app opens a Settings window with its preferences

#### Scenario: Preferences persist
- **WHEN** the user changes a preference and relaunches the app
- **THEN** the preference retains its value

### Requirement: Adjustable content font size
The app SHALL let the user adjust the size of rendered content (the artifact/markdown reading pane) via a font-size preference, applied as a multiplier on top of the system Dynamic Type size. It SHALL be adjustable from the Settings window and via keyboard shortcuts (⌘+ increase, ⌘− decrease, ⌘0 reset), and SHALL apply only to rendered content — not the sidebar or app chrome.

#### Scenario: Adjust from settings
- **WHEN** the user changes the content text-size control in Settings
- **THEN** rendered content resizes accordingly and the change persists

#### Scenario: Adjust via keyboard
- **WHEN** the user presses ⌘+, ⌘−, or ⌘0
- **THEN** the rendered content size increases, decreases, or resets, consistent with the Settings control (they share one stored value)

#### Scenario: Content-only scope
- **WHEN** the content font size is changed
- **THEN** only the rendered content resizes; the sidebar, toolbar, and other chrome are unaffected
