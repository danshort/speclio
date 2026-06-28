## MODIFIED Requirements

### Requirement: Immediate reload after closing the editor
The TUI SHALL reload the content of the edited artifact immediately upon returning from the editor, without waiting for the next polling cycle. On return, the TUI SHALL invalidate the render cache for the current tab unconditionally — including when the editor was closed without saving — so that a viewport resize that occurred while the editor was open cannot leave a stale, mis-wrapped render on screen.

#### Scenario: Reload tasks after editing
- **WHEN** the user edits `tasks.md` in the editor and closes the editor
- **THEN** the TUI shows the updated tasks content instantly, with the cursor restored by text

#### Scenario: Reload markdown artifact after editing
- **WHEN** the user edits `proposal.md`, `design.md`, or a `spec.md` and closes the editor
- **THEN** the TUI invalidates the render cache for that tab and re-renders with the new content

#### Scenario: Re-render after an unsaved exit following a resize
- **WHEN** the user opens the editor on a change artifact, the terminal is resized while the editor is open, and the user exits without saving
- **THEN** the TUI drops the current tab's cached render and re-renders the artifact at the new viewport width, showing correctly wrapped content rather than the cached render from the old width
