import Foundation
import OpenSpecKitGolden

let failures = runAllGoldenChecks()
if failures.isEmpty {
    print("✓ all golden checks passed")
    exit(0)
}
for f in failures {
    print("✗ \(f)")
}
print("\(failures.count) golden check(s) failed")
exit(1)
