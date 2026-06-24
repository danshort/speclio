## 1. Arrow-key tab navigation

- [x] 1.1 In `updateViewer`, add `right` to the `Tab` case and `left` to the `Shift+Tab` case (reuse `nextAvailableTab`)
- [x] 1.2 Confirm `h`/`l` (change navigation) are unchanged

## 2. Surfacing

- [x] 2.1 Update the normal and tasks footer hints to `1-4/Tab/←→: artifact`
- [x] 2.2 Add `Tab`/`Shift+Tab` and `←`/`→` rows to the README keyboard reference (EN + ES)

## 3. Tests

- [x] 3.1 Add tests asserting `right` advances to the next available tab and `left` to the previous, skipping disabled tabs

## 4. Verification

- [x] 4.1 `gofmt`, `go vet`, `go test -race ./...` pass
