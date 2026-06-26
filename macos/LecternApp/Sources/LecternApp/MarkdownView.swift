import SwiftUI
import Markdown

// MarkdownView renders an OpenSpec artifact with swift-markdown. Apple's
// AttributedString(markdown:) is single-paragraph and can't do tables / fenced
// code / nested lists, so we walk the AST into SwiftUI views ourselves (the
// decision recorded in design.md).

struct MarkdownView: View {
    private let document: Document
    // Larger-than-default base that still scales with the system text size
    // (Dynamic Type), instead of a fixed point size.
    @ScaledMetric(relativeTo: .body) private var bodySize: CGFloat = 15

    init(_ text: String) {
        self.document = Document(parsing: text)
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            ForEach(Array(document.children.enumerated()), id: \.offset) { _, child in
                BlockView(markup: child)
            }
        }
        .font(.system(size: bodySize))
        .lineSpacing(6)
        .textSelection(.enabled)
    }
}

/// Renders a single block-level Markup node (and recurses for containers).
struct BlockView: View {
    let markup: Markup

    var body: some View {
        switch markup {
        case let heading as Heading:
            Text(InlineText.attributed(heading))
                .font(headingFont(heading.level))
                .bold()
                .padding(.top, heading.level <= 2 ? 12 : 6)
                .padding(.bottom, 2)

        case let paragraph as Paragraph:
            Text(InlineText.attributed(paragraph))
                .fixedSize(horizontal: false, vertical: true)

        case let code as CodeBlock:
            CodeBlockView(code: code.code)

        case let list as UnorderedList:
            ListView(items: Array(list.listItems), ordered: false)

        case let list as OrderedList:
            ListView(items: Array(list.listItems), ordered: true)

        case let quote as BlockQuote:
            BlockQuoteView(quote: quote)

        case is ThematicBreak:
            Divider().padding(.vertical, 4)

        case let table as Markdown.Table:
            MarkdownTableView(table: table)

        default:
            // Unknown container: render its block children.
            VStack(alignment: .leading, spacing: 8) {
                ForEach(Array(markup.children.enumerated()), id: \.offset) { _, child in
                    BlockView(markup: child)
                }
            }
        }
    }

    private func headingFont(_ level: Int) -> Font {
        switch level {
        case 1: return .title
        case 2: return .title2
        case 3: return .title3
        default: return .headline
        }
    }
}

struct CodeBlockView: View {
    let code: String

    var body: some View {
        Text(code.hasSuffix("\n") ? String(code.dropLast()) : code)
            .font(.system(.callout, design: .monospaced))
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(10)
            .background(Color.secondary.opacity(0.12))
            .clipShape(RoundedRectangle(cornerRadius: 6))
            .textSelection(.enabled)
    }
}

struct ListView: View {
    let items: [ListItem]
    let ordered: Bool

    var body: some View {
        VStack(alignment: .leading, spacing: 7) {
            ForEach(Array(items.enumerated()), id: \.offset) { index, item in
                HStack(alignment: .top, spacing: 8) {
                    Text(ordered ? "\(index + 1)." : "•")
                        .monospacedDigit()
                        .foregroundStyle(.secondary)
                    VStack(alignment: .leading, spacing: 4) {
                        ForEach(Array(item.children.enumerated()), id: \.offset) { _, child in
                            BlockView(markup: child)
                        }
                    }
                }
            }
        }
        .padding(.leading, 4)
    }
}

struct BlockQuoteView: View {
    let quote: BlockQuote

    var body: some View {
        HStack(alignment: .top, spacing: 8) {
            RoundedRectangle(cornerRadius: 1.5)
                .fill(Color.secondary.opacity(0.5))
                .frame(width: 3)
            VStack(alignment: .leading, spacing: 8) {
                ForEach(Array(quote.children.enumerated()), id: \.offset) { _, child in
                    BlockView(markup: child)
                }
            }
        }
    }
}

struct MarkdownTableView: View {
    let table: Markdown.Table

    var body: some View {
        let headCells = Array(table.head.cells)
        let rows = Array(table.body.rows)
        Grid(alignment: .leadingFirstTextBaseline, horizontalSpacing: 16, verticalSpacing: 6) {
            GridRow {
                ForEach(Array(headCells.enumerated()), id: \.offset) { _, cell in
                    Text(InlineText.attributed(cell)).bold()
                }
            }
            Divider()
            ForEach(Array(rows.enumerated()), id: \.offset) { _, row in
                GridRow {
                    ForEach(Array(row.cells.enumerated()), id: \.offset) { _, cell in
                        Text(InlineText.attributed(cell))
                    }
                }
            }
        }
        .padding(10)
        .background(Color.secondary.opacity(0.06))
        .clipShape(RoundedRectangle(cornerRadius: 6))
    }
}

/// Builds an AttributedString from a node's inline children, honoring emphasis,
/// strong, inline code, and links (rendered via SwiftUI Text's support for
/// inlinePresentationIntent).
enum InlineText {
    static func attributed(_ markup: Markup) -> AttributedString {
        var result = AttributedString()
        for child in markup.children {
            result += render(child)
        }
        return result
    }

    private static func render(_ markup: Markup) -> AttributedString {
        switch markup {
        case let text as Markdown.Text:
            return AttributedString(text.string)

        case let emphasis as Emphasis:
            return applying(attributed(emphasis), .emphasized)

        case let strong as Strong:
            return applying(attributed(strong), .stronglyEmphasized)

        case let code as InlineCode:
            return applying(AttributedString(code.code), .code)

        case let link as Markdown.Link:
            var inner = attributed(link)
            if let dest = link.destination, let url = URL(string: dest) {
                inner.link = url
                inner.underlineStyle = .single
            }
            return inner

        case is LineBreak:
            return AttributedString("\n")

        case is SoftBreak:
            return AttributedString(" ")

        default:
            return attributed(markup)
        }
    }

    /// Merges an inline presentation intent across all runs, preserving any
    /// nested intents (e.g. bold inside italic) instead of overwriting them.
    private static func applying(_ input: AttributedString, _ intent: InlinePresentationIntent) -> AttributedString {
        var copy = input
        for run in copy.runs {
            var current = copy[run.range].inlinePresentationIntent ?? []
            current.formUnion(intent)
            copy[run.range].inlinePresentationIntent = current
        }
        return copy
    }
}
