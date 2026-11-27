# Parley

One client for **Telegram**, **Slack**, **VK Teams** and **VK**, plus **mail and calendars**
(Google, Mail.ru / VK WorkSpace, your own IMAP/SMTP/CalDAV servers). Any number of accounts per
provider. All chats are in one list, all inboxes in one mailbox, all calendars in one month view;
every chat, message and event knows its account.

Parley was formerly called MegaChat.

Parley is part of [SelfHold](https://github.com/SelfHold), a set of self-hosted apps that share
one architecture: a Go core inside the app's process and a Flutter UI. It can keep its vault in
sync between devices through [Shard](https://github.com/SelfHold/shard) (optional).

## Status

A personal project at an early stage (version 0.1.0). It is used daily by its author, but expect
rough edges, missing features and protocol details that are not verified against every server
(see the caveats below). There are no releases yet: build from source.

## Platforms

- **Linux** desktop and **Android** (arm64, x86_64; Android 8.0+).
- **macOS and iOS** are planned, not built yet.
- Windows is not a target. Leftover Windows build files are in the tree but are not maintained.

## Features

Chats (all messengers): chat list with unread counters, history (cached and gap-filled), sending
and replying, read marks, typing, Markdown formatting in the composer, link previews (optional),
playing audio, voice and video in the app, desktop and Android notifications. What else a chat
offers depends on the messenger; the app shows a feature only when the account supports it:

| | Telegram | Slack | VK Teams | VK |
|---|---|---|---|---|
| Edit / delete own messages | yes / yes (also "for me" outside channels) | yes / yes (for everyone only) | yes / yes (also "for me") | yes / yes (also "for me") |
| Send files and pictures | yes (albums) | yes | yes | yes |
| Reactions | yes | yes | yes | yes |
| Forward (inside one account) | yes | yes (as a permalink) | yes | yes |
| Threads / channel comments | yes | yes | yes | – |
| Forum topics | yes | – | – | – |
| Mentions with "@" | yes | yes | yes | yes |
| Chat info (profile, members) | yes | yes | yes | yes |
| Chat folders | yes | yes (sidebar sections, browser session only) | – | – |
| Mute / pin / archive / leave | yes | leave | mute, archive, leave | yes |
| Presence (online dot) | yes | yes (browser session) | yes | yes |
| 1:1 audio and video calls | yes (WebRTC, end-to-end encrypted) | – | – | – |

Group calls are planned. Forwarding to another account, or forwarding a thread reply or a text
fragment, sends the text ("Forwarded from …").

### Sign-in

| | Sign-in | Live events | Notes |
|---|---|---|---|
| Telegram | QR code scanned in Telegram on the phone (Settings → Devices → Link Desktop Device), or phone → code; then the cloud password if set | MTProto updates with gap recovery | needs app credentials from [my.telegram.org](https://my.telegram.org): build with `TG_APP_ID` / `TG_APP_HASH`, otherwise the sign-in asks for them |
| Slack: browser session | workspace + `xoxc-` token + `d` cookie from the web client | RTM websocket | works in any workspace without admin approval; unofficial |
| Slack: OAuth app | your app's client id/secret, approval in the browser | Socket Mode with an `xapp-` token; without one, the 20 most active chats are polled | redirect to `http://localhost:53682/slack/callback`, or paste the address you land on |
| VK Teams: login/password | server, login, password, client key | `fetchEvents` long-poll | the password is used once; a token renews sessions |
| VK Teams: web session | the address of a `fetchEvents` request of the web client (DevTools → Network) | `fetchEvents` long-poll | ends when the web session ends |
| VK | QR code scanned in VK on the phone | long-poll | see the caveat below |

**VK Teams caveat:** VK publishes no documentation for the user-level client protocol. The
implementation follows what VK WorkSpace client logs show and the ICQ/AIM WIM design that VK
Teams descends from. All wire details are in `core/providers/vkteams/wire.go`; several features
(threads, forwarding, quotes, files, edits) are marked unverified there. The server host and API
version can be set per account (on-premises installations).

**VK caveat:** VK's messages API is open only to its own applications. The VK provider signs in
and talks to the API as the official VK Messenger desktop client does (`core/providers/vk/wire.go`).
This is against VK's terms of service, and VK may change the client key at any time, after which
sign-in stops working until the provider is updated. Use it at your own risk.

### Mail and calendars

| | Sign-in | Mail | Calendar | Live changes |
|---|---|---|---|---|
| Google | OAuth in the browser (loopback redirect + PKCE); refresh token in the vault | Gmail API (labels as folders, All mail as archive) | Google Calendar API | Gmail history every minute, calendars every 3 minutes |
| Mail.ru / VK WorkSpace | address + **app password** (mailbox settings → Security → passwords for external apps) | IMAP `imap.mail.ru:993`, SMTP `smtp.mail.ru:465` | CalDAV `calendar.mail.ru` (discovered) | inbox counters every minute, calendar ETags every 5 minutes |
| Own servers (on-premises VK WorkSpace, any other) | address, password, IMAP/SMTP hosts, optional CalDAV URL | IMAP/SMTP (993/465 TLS, 587 STARTTLS) | CalDAV | same |

- All inboxes (and all sent, drafts, …) are merged into one list, newest first. Messages can be
  marked, moved and deleted in bulk, across accounts.
- Nothing of the mail is cached on disk: lists and bodies come from the providers on every read.
  Accounts that fail are reported next to the results; the others still show.
- Messages are shown as text with styles; an HTML-only message is reduced to text in the core,
  and remote content is never loaded. Outgoing mail is sent as plain text + HTML from Markdown.
- Events can be created, edited and deleted, including repeating ones (this occurrence / this
  and following / the whole series), in Google Calendar and CalDAV.
- Google needs an OAuth client of type "Desktop app" with the Gmail API and the Calendar API
  enabled (for an unverified app, add yourself as a test user): build with `GOOGLE_CLIENT_ID` /
  `GOOGLE_CLIENT_SECRET`, otherwise the sign-in asks for them.
- **Mail.ru caveat:** the CalDAV address, and that Mail.ru keeps a copy of SMTP-sent mail in
  Sent, follow Mail.ru's help pages; the tests run against local IMAP, SMTP and CalDAV servers,
  not the live service.

### Language

The UI is in English. Russian can be chosen in Settings; the choice is kept per device and is
never taken from the system locale.

## Security model

- **Vault** (`vault.parley`): the accounts and their credentials (tokens, session keys, app
  passwords), encrypted with a key derived from your password (Argon2id + XChaCha20-Poly1305).
  Unlocking it starts all accounts. The password itself is never stored.
- **Remember on this machine** (desktop, optional): keeps the password-derived key, not the
  password, in the system keychain (Secret Service: GNOME Keyring, KWallet). On Android the vault
  can instead be unlocked by fingerprint: the password is sealed in the Android Keystore.
- **Cache** (`cache-<id>.db`): chats, messages, media and each account's session state. Every
  value is sealed and every key is an HMAC, so the file shows nothing without the vault. Mail and
  calendars are not cached on disk at all.
- Credentials never leave the core: the UI never holds them, and logs, errors and RPC responses
  carry no tokens or passwords. Credentials are only sent to the messenger's own hosts.
- The UI talks to the core over a private socket with a bearer token.
- Link previews are fetched by the core only from public addresses; the site sees your address,
  so they can be turned off in Settings.

See [SECURITY.md](SECURITY.md) to report a vulnerability.

## Architecture

```
   ┌──────────────── one process: the app ─────────────────┐
   │ Flutter UI (app/): unlock · chats · mail · calendar   │       parley (CLI)
   │        │ gRPC (proto/parley.proto), bearer token    │          │
   │        │ unix socket                                  │          │
   │ core/embed ── desktopbind (FFI) | mobilebind (AAR) ◄─┼── endpoint.json
   └────────┼──────────────────────────────────────────────┘
            │
   core/rpc ── core/hub ── core/provider ── providers/telegram (MTProto, gotd/td)
                  │                     ├── providers/slack    (Web API + RTM / Socket Mode / polling)
                  │                     ├── providers/vkteams  (WIM/RAPI client protocol)
                  │                     ├── providers/vk       (VK API as the official client)
                  │                     ├── providers/google   (Gmail + Calendar REST, OAuth)
                  │                     └── providers/mailru   (IMAP/SMTP + CalDAV)
             core/store (bbolt, encrypted)    core/vault (password, Argon2id + XChaCha20)
```

- **hub**: one session per enabled account, reconnecting with backoff. Session events go into the
  store and then to subscribers. History is read from the cache, gaps are filled from the provider.
- **provider**: what a messenger or mail service plugs into: `Info`, `Login` (a step-by-step
  sign-in dialog that the app renders generically) and `Session` with capability interfaces
  (messages, threads, files, calls, mail, calendar, …). The hub and the app know no provider.
- There is no daemon: the core runs inside the app's process. On Linux it is linked into the
  executable (`desktopbind` built as a C archive, called through Dart FFI); on Android it is a
  gomobile AAR kept alive by a foreground service that also shows notifications. Accounts stay
  connected while the app runs, also when its window is hidden in the tray. A second instance of
  the app uses the core of the first one.

More detail: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

## Build from source

Prerequisites:

- Go 1.27 and Flutter 3.41 (Dart 3.11) with the Linux desktop toolchain (clang, CMake, Ninja,
  pkg-config, GTK 3 headers).
- Linux libraries used by the app's plugins: libmpv (media playback, `media_kit`) and
  libayatana-appindicator3 (tray icon, `tray_manager`). A Secret Service provider (GNOME Keyring,
  KWallet) is needed only for "remember on this machine".
- Android: the Android SDK and NDK (the Makefile takes the newest NDK under `~/Android/Sdk/ndk`;
  override with `NDK=…`) and gomobile (`go install golang.org/x/mobile/cmd/gomobile@latest`,
  then `gomobile init`).
- Only to regenerate code after changing a `.proto`: protoc, protoc-gen-go, protoc-gen-go-grpc
  and the Dart plugin (`dart pub global activate protoc_plugin`).

```bash
make build-linux                                             # builds/linux/parley: the app with the core in it
make build-linux TG_APP_ID=123 TG_APP_HASH=abc               # with built-in Telegram app credentials
make build-linux GOOGLE_CLIENT_ID=… GOOGLE_CLIENT_SECRET=…   # with a built-in Google OAuth client
make build-apk                                               # builds/android/*.apk (arm64, x86_64)
make build                                                   # builds/parley: the debug CLI
make run-demo                                                # the app with the Demo messenger, no real accounts
```

All outputs go to `builds/`. Development commands (tests, code generation) are in
[CONTRIBUTING.md](CONTRIBUTING.md).

## Use

Start the app, create a vault with a password, then add accounts in Accounts. `make run-demo`
(or `PARLEY_DEMO=1`) adds the **Demo** messenger, an in-memory provider with generated chats,
mail and a call peer.

`parley` is a debug CLI that talks to the core of the running app (the cache belongs to one
process):

```bash
builds/linux/parley &
parley create                  # new vault (asks for a password); later: parley unlock [remembered]
parley providers
parley login telegram          # the same sign-in dialog the app shows
parley login slack session
parley accounts
parley chats -n 20
parley history ACCOUNT CHAT -n 50
parley send ACCOUNT CHAT hello
parley sendfile ACCOUNT CHAT ./photo.jpg caption...
parley login google            # or: parley login mailru [server]
parley folders
parley mail -n 20              # all inboxes; parley mail ACCOUNT FOLDER for one folder
parley mailread ACCOUNT ID
parley sendmail ACCOUNT to@example.com "Subject" text...
parley events -n 7             # the next 7 days of every calendar
parley watch                   # live events
```

Files on Linux: `~/.config/selfhold/parley/` (or `$XDG_CONFIG_HOME/selfhold/parley/`) holds `vault.parley`,
`cache-*.db`, `app.json` (UI settings of this device), `endpoint.json` (while the app runs) and
`parley.log`.

## Sync through Shard (optional)

Parley works on its own. With [Shard](https://github.com/SelfHold/shard) installed and running,
the "Sync via Shard" switch in Settings keeps the vault file in your Shard vault, so all your devices get the same
accounts. The local file stays the working copy; copies from different devices are merged
(union by account, newest change wins, removals are kept), and the first sync lets you choose
when the two vaults differ. Per-device state (pairing, last synced revision) is never synced.
On desktops Parley reaches Shard through its local Connector socket, on Android through Shard's
bound service.

The protocol is described in the
[Connector guide](https://github.com/SelfHold/shard/blob/main/docs/CONNECTOR.md). Parley's copy
of `proto/connector.proto` predates the newer scopes (SSH keys, servers), which it does not use.

## License

Parley is free software: you can redistribute it and/or modify it under the terms of the GNU
General Public License, version 3 or (at your option) any later version. See [LICENSE](LICENSE).
The bundled fonts (JetBrains Mono, Syne) are under the SIL Open Font License
(`app/assets/fonts/OFL-*.txt`).
