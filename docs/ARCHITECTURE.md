# Parley architecture

Parley is a Go core (`core/`) that runs inside the app's process, a
Flutter app (`app/`) that only renders, and a debug CLI (`cmd/parley`). This document lists how
the parts fit together and the rules that changes must keep. The overview is in the
[README](../README.md).

## Layout

| Path | What |
|---|---|
| `proto/parley.proto` | The gRPC API between the core and the app (and the CLI). |
| `core/embed` | Starts the core in a process: vault, store, hub, gRPC server on a private socket, `endpoint.json`. |
| `desktopbind` | The core as a C archive linked into the Linux executable, called through Dart FFI. Also desktop notifications (D-Bus). |
| `mobilebind` | The core as a gomobile AAR for Android. |
| `core/rpc` | The gRPC service: converts between the proto and the model, publishes events. |
| `core/hub` | One session per enabled account; history, mail and calendar reads over all accounts. |
| `core/provider` | The plugin interfaces: `Provider`, `Login`, `Session` and capability interfaces. |
| `core/providers/*` | Telegram, Slack, VK Teams, VK, Google, Mail.ru / IMAP+CalDAV, and the in-memory Demo (`fake`). |
| `core/vault` | The encrypted vault: accounts and credentials; merging synced copies. |
| `core/store` | The encrypted cache (bbolt): chats, messages, media, provider session state. |
| `core/keychain` | The OS keychain (go-keyring) for "remember on this machine". |
| `core/shard` | The client of Shard's Connector API (optional vault sync). |
| `core/richtext`, `core/recur`, `core/linkpreview` | Formatting spans, repeating events, link previews. |
| `app/lib` | The Flutter app: `data/` (RPC client, native core), `presentation/` (screens, `common/` widget kit), `core/` (theme, i18n, settings). |
| `app/android` | Kotlin glue: core service, notifications, fingerprint, Shard binding, clipboard files. |

Generated code (`core/rpc/pb`, `core/shard/pb`, `app/lib/data/rpc/gen`) comes from
`make proto` and is never edited by hand.

## Rules

### Accounts and providers

- An account is identified by its vault account id, never by provider type: several accounts of
  one provider are normal. Chat and message ids are the provider's own and unique only inside an
  account, so every API call, event, store key and UI key carries the account id too.
- One messenger account is added once (`vault.Config.SameLogin`: provider + login, e.g. the
  Telegram user id, Slack `team:user`). A re-login must sign in as the same login and keeps the
  account id.
- Providers are plugins behind `core/provider`. The hub knows only `Provider` / `Session` /
  `Login` and capability interfaces checked by type assertion (messages, threads, reactions,
  media, forwarding, files, edit, delete, mentions, profiles, chat folders, chat management,
  calls, mail, calendars). A new feature is a new capability interface; provider-specific code
  never goes into the hub, the rpc layer or the app. The hub lists an account's capabilities in
  `Account.capabilities`, and the app shows a feature only when it is there.
- Sign-in is a dialog of `provider.Step`s driven by the core (fields, a URL to open, errors in the
  step). The app renders steps generically; it has no per-provider login screens.
- Protocol details that cannot be checked against documentation stay in one file per provider
  (`vkteams/wire.go`, `vk/wire.go`) and are marked unverified there.

### Secrets and storage

- Credentials never leave the core: the account conversion to the proto copies no parameters;
  logs, errors and RPC responses carry no tokens or passwords. Passwords used only to obtain a
  token (VK Teams) are not stored. Credentials are sent only to the messenger's own hosts.
- Nothing readable on disk. The vault is Argon2id + XChaCha20-Poly1305. The cache has HMAC bucket
  and key names, and values sealed with their location as associated data. Provider session state
  (Telegram auth key, update state, access hashes) goes into the account's key-value space in the
  store.
- "Remember the password on this machine" (desktops) keeps the vault's password-derived key,
  never the password, in the OS keychain; a marker file in the data dir only says that it is
  there. Changing the password re-keeps it; a key that no longer opens the vault is forgotten.
  The app tries it once per start, not after "Lock". Android has no keychain: the password is
  sealed in the Android Keystore behind a fingerprint instead.
- The vault file may live anywhere on desktops: `<data dir>/vault.path` holds its path (the only
  plain file, no secret), else `<data dir>/vault.parley`. The cache stays in the data dir, keyed by
  the vault id.
- One process per cache: bbolt locks the file. The core lives in the app's process (no daemon);
  the CLI and a second app instance reach it through `endpoint.json`, never open the store.
- App preferences (language, notifications, tray, media autoload) are per device, in
  `<config dir>/parley/app.json` (Android: app support dir). They are never in the vault and
  never synced.

### Sync through Shard

- Optional and off by default; Parley works without Shard, and Shard may be missing, stopped or
  locked at any time. The protocol is Shard's Connector API, see the
  [Connector guide](https://github.com/SelfHold/shard/blob/main/docs/CONNECTOR.md).
- The vault file's bytes are record `vault.parley` of app id `parley`. Both names are persisted on
  users' devices and in Shard and must never change.
- The local file stays the working copy. Per-device state (`<data dir>/shard.json`: base
  revision, synced hash, Shard vault id; the pairing token in the keychain, else
  `shard.token`) is never synced.
- Desktops dial Shard's `connector.json`; Android binds Shard's service through Kotlin
  (`Shard.kt`, the AIDL copied as is from Shard) and `mobilebind.SetShard`.
- Merging (`vault.Merge`) is deterministic: union by account id, the newer `Account.Updated` wins,
  removals are tombstones, the same result in any order on every device; the latest password
  change seals the result. The first sync compares and lets the user pick on a conflict.
- `proto/connector.proto` and the AIDL files are copies of Shard's and are changed only by copying
  Shard's byte for byte.

### History, events and media

- The history cache keeps contiguity: a message is linked only when the next older cached message
  really is its predecessor (fetched pages). Live messages are stored unlinked; the hub fills gaps
  from the provider. Deleted messages are kept as tombstones.
- Applying events is idempotent: a message seen again (sent here, then echoed by the server) is a
  no-op or an edit, never a second message or a second unread.
- Publishing events never blocks: a subscriber that falls behind is cut off and reloads; the app
  reloads everything after every reconnect.
- Media (avatars, attachments, previews) are opaque per-account refs fetched with the `Media` RPC.
  The core caches them sealed in the store (LRU, capped); the app keeps them in memory and writes
  a file only when the user opens or saves one.
- Thread replies are not part of a chat's history: not cached, fetched with `Thread`, live ones
  come as thread-message events.
- Forwarding natively works only inside one account and only for messages of the chat's history.
  Anything else the app sends as text ("Forwarded from X:").
- Files to send ride in the request (100 MB together). Pasted files and pictures never touch the
  disk.
- Presence is kept in memory per session, never in the store.
- Mail and calendars are not cached on disk: the hub reads every account at once and merges (mail
  by time, then account, then id). A failing account is reported next to the results and never
  fails the whole read. Mail remote content is never fetched; mail HTML is reduced to text and
  styles.
- Repeating events: the series' RRULE is on every occurrence; saves and deletes take a scope
  (this / this and following / all). A series never silently becomes a single event.
- Link previews: the core fetches Open Graph data itself, only from public addresses (checked on
  every dial, redirects too), size- and time-limited, cached in memory, never on disk.

### Calls and notifications

- Call media is a standard WebRTC peer connection in the app; the provider only carries the
  signaling. The app knows no messenger. One live call per device over all accounts: a second
  incoming call is turned down as busy. Telegram calls speak tgcalls 11.0.0.
- The hub decides what deserves a notification (someone else wrote, the chat is not muted or the
  user is mentioned). Desktops show it from the app through `desktopbind`; Android shows it in
  Kotlin (`Notifications.kt`), also while the UI is not running.

## App

- Every UI string goes through `tr('English', 'Русский')` (`app/lib/core/i18n.dart`). The app
  starts in English whatever the system locale; Russian is a choice in Settings, kept in
  `app.json`. Android's Kotlin notification strings follow the same choice: the app passes the
  language over the `dev.selfhold.parley/core` method channel at start and on every change.
- Screens use only the widget kit in `app/lib/presentation/common/` and the design tokens
  (`HmColors`, `HmText`): dark theme only, no raw colours, no Material elevation, bundled fonts.
  Below 768 px of width the phone layout is used, on any platform.
- Linux windows are undecorated: the app's header is the title bar.
- Audio, voice and video play in the app (media_kit / libmpv). `app/packages/wakelock_plus`
  overrides media_kit_video's dependency to keep test builds within the 1 GB memory cap; keep the
  override.
