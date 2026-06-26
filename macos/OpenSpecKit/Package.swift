// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "OpenSpecKit",
    platforms: [.macOS(.v13)],
    products: [
        .library(name: "OpenSpecKit", targets: ["OpenSpecKit"]),
    ],
    dependencies: [
        .package(url: "https://github.com/jpsim/Yams.git", from: "5.1.0"),
    ],
    targets: [
        // The UI-free domain port (mirrors Go internal/openspec).
        .target(name: "OpenSpecKit", dependencies: ["Yams"]),
        // Shared golden-verification logic, importable by both the executable
        // runner (CLT-friendly) and the XCTest target (needs Xcode).
        .target(name: "OpenSpecKitGolden", dependencies: ["OpenSpecKit"]),
        // CLT-runnable verifier: `swift run oskgolden` (exit 1 on any mismatch).
        .executableTarget(name: "oskgolden", dependencies: ["OpenSpecKitGolden"]),
        // Standard `swift test` lane (requires Xcode's XCTest).
        .testTarget(name: "OpenSpecKitTests", dependencies: ["OpenSpecKitGolden"]),
    ]
)
