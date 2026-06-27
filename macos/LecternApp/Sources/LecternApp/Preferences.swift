import SwiftUI

// Preferences shared across the app. `contentFontScale` is a user multiplier
// applied on top of Dynamic Type, to the rendered content only (#63). One
// @AppStorage key is the single source of truth; the Settings slider, the
// ⌘±/⌘0 commands, and MarkdownView all read/write it.
enum ContentFont {
    static let storageKey = "contentFontScale"
    static let defaultScale = 1.0
    static let minScale = 0.8
    static let maxScale = 2.0
    static let step = 0.1

    static func clamp(_ value: Double) -> Double {
        min(max(value, minScale), maxScale)
    }
}

// The Settings (⌘,) "General" pane. One pane for now; wrap in a TabView when a
// second category of preferences exists.
struct GeneralSettingsView: View {
    @AppStorage(ContentFont.storageKey) private var scale = ContentFont.defaultScale

    var body: some View {
        Form {
            Section("Content") {
                VStack(alignment: .leading, spacing: 8) {
                    Text("Text size")
                    HStack(spacing: 12) {
                        Text("A").font(.system(size: 11))
                        Slider(value: $scale, in: ContentFont.minScale...ContentFont.maxScale, step: ContentFont.step)
                        Text("A").font(.system(size: 22))
                    }
                    HStack {
                        Text("\(Int((scale * 100).rounded()))%")
                            .font(.caption).monospacedDigit().foregroundStyle(.secondary)
                        Spacer()
                        Button("Reset") { scale = ContentFont.defaultScale }
                            .disabled(scale == ContentFont.defaultScale)
                    }
                    Text("Adjusts only rendered content (proposals, specs, tasks), on top of the system text size.")
                        .font(.caption).foregroundStyle(.secondary)
                }
                .padding(.vertical, 4)
            }
        }
        .formStyle(.grouped)
        .frame(width: 440)
    }
}
