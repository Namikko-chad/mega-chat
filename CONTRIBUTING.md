# Contributing to Parley

Thanks for your interest. Parley is a personal project at an early stage, so please open an issue
to discuss a larger change before writing it. Bug reports with steps to reproduce are welcome.
Security problems go through [SECURITY.md](SECURITY.md), not public issues.

Read [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) first: it lists the rules a change must keep
(accounts, secrets, storage, sync, events).

## Development commands

Everything goes through `make`; outputs land in `builds/`. Prerequisites are listed in the
[README](README.md#build-from-source).

```bash
make test        # Go core tests
make race        # Go core tests with the race detector (skips the packages that link Telegram)
make vet         # go vet
make app-test    # flutter analyze + flutter test
make proto       # regenerate Go and Dart code after changing proto/parley.proto
make run-demo    # the app with the Demo messenger (no real accounts)
make app-run     # flutter run with the build defines
make core-lib    # the desktop core as a C archive (the Linux app build runs it itself)
make aar         # the Android core (gomobile)
make build-linux | make build-apk | make build
```

- Test runs are capped at 1 GB of memory and run one at a time: the `make` test targets wrap the
  cap with `systemd-run --user --scope -p MemoryMax=1G`. Linking a test binary with gotd's
  generated `tg` package only fits with `GOMAXPROCS=1 GOGC=10 go test -p 1 -vet=off -ldflags=-w`,
  which `make test` does. Use the targets rather than bare `go test` / `flutter test`.
- Hub and rpc tests use the in-memory `fake` provider; Slack, VK Teams, VK, Google and the
  IMAP/SMTP/CalDAV provider are tested against local fake servers; Telegram's mapping with unit
  tests. Shard sync is tested against the fake in `core/shard/fakeshard`, never a live Shard.
- Work on screens with `make run-demo` and `app/tool/screenshots.dart`.
- Generated code (`core/rpc/pb`, `core/shard/pb`, `app/lib/data/rpc/gen`) is never edited by
  hand.
- Format Go with `gofmt` and Dart with `dart format`.

## Rules

- **English only** in code, comments, identifiers, logs, error messages, commit messages and
  docs. Other languages appear only as translations:
  - UI strings go through `tr('English', 'Русский')` (`app/lib/core/i18n.dart`); add both.
  - A translated document (e.g. `README.ru.md`) must match its English original.
  - Test data and protocol data in another language are fine (e.g. a mail server's Russian folder
    names); say in an English comment why the string is not English.
- **UI language**: the app starts in English regardless of the system locale; Russian only when
  chosen in Settings. Don't read the platform locale to pick a language. Android-side strings
  (Kotlin) follow the app's choice (`Notifications.tr`).
- **No secrets anywhere** they could leak: logs, errors, RPC replies, plain files, the app. Don't
  commit credentials, tokens or local config.
- **Don't rename persisted identifiers**: the app id `parley`, the record name `vault.parley`, file
  names in the data dir, the Android package `dev.selfhold.parley`, the AIDL package, keychain entries.
  They live on users' devices and in Shard; changing one needs a migration.
- **Connector contract**: `proto/connector.proto` and the AIDL under
  `app/android/app/src/main/aidl/dev/selfhold/shard/connector/` are owned by
  [Shard](https://github.com/SelfHold/shard). Change them there first, then copy byte for byte.
  Parley must keep working with an older Shard and without Shard at all.
- **Platforms**: Linux and Android. Keep platform-specific code behind the existing seams
  (`desktopbind`, `mobilebind`, `core/keychain`, the data-dir lookup, Kotlin under `app/android`)
  so macOS and iOS can be added later. Windows is not a target.
- **UI**: dark theme, design tokens only (`HmColors`, `HmText`), widgets from
  `app/lib/presentation/common/`, no Material elevation; every screen must work in the phone
  layout (below 768 px).
- Keep provider-specific code inside its provider; the hub, the rpc layer and the app stay
  generic (see the architecture document).

## Pull requests

- One topic per pull request, against `main`. Describe what changed and how you tested it.
- Run `make vet test app-test` (one after another) before opening it.
- By contributing you agree that your contribution is licensed under the GNU GPL v3 or later,
  like the rest of the project.
