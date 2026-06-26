import XCTest
import OpenSpecKitGolden

final class GoldenTests: XCTestCase {
    func testAllGoldenChecks() {
        let failures = runAllGoldenChecks()
        for f in failures {
            XCTFail(f.description)
        }
    }
}
