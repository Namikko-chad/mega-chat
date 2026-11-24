# MegaChat

One messenger client for **VK Teams** (VK WorkSpace), **Slack** and **Telegram**, with any number
of accounts per messenger, on Linux, Windows and Android. All chats are in one list; each chat and
message knows its account.

Chats: the chat list, history, sending and replying, read marks, typing, edits and deletions.
**Mail and calendars** of Google and Mail.ru / VK WorkSpace accounts: all inboxes (and all sent,
drafts, …) merged into one list, all calendars of all accounts in one month view. Group calls are
planned. Providers declare
capabilities (`Capability` in `proto/megachat.proto`), so these can be added to the providers
that support them without touching the rest.

## Architecture

```
   ┌──────────────── one process: the app ─────────────────┐
   │ Flutter UI (app/): unlock · chats · chat · mail · log │       mchat (CLI)
   │        │ gRPC (proto/megachat.proto), bearer token    │          │
   │        │ unix socket / localhost tcp (Windows)        │          │
   │ core/embed ── desktopbind (FFI) | mobilebind (AAR) ◄─┼── endpoint.json
   └────────┼──────────────────────────────────────────────┘
            │
   core/rpc ── core/hub ── core/provider ── providers/telegram (MTProto, gotd/td)
                  │                     ├── providers/slack    (Web API + RTM / Socket Mode / polling)
                  │                     ├── providers/vkteams  (WIM/RAPI client protocol)
                  │                     ├── providers/google   (Gmail + Calendar REST, OAuth)
                  │                     └── providers/mailru   (IMAP/SMTP + CalDAV)
             core/store (bbolt, encrypted)    core/vault (password, Argon2id + XChaCha20)
```

- **vault** (`vault.mc`): master key and accounts with their credentials, encrypted with a key
  derived from the password. Unlocking starts all accounts. On desktops "Remember the password
  on this machine" keeps that derived key (not the password) in the system keychain — Secret
  Service (GNOME Keyring, KWallet) on Linux, Credential Manager on Windows — and the vault opens
  by itself when the app starts; on Android the password is sealed behind a fingerprint instead.
- **store** (`cache-<id>.db`): chats, messages and each account's session state. Every value is
  sealed and every key is an HMAC, so the file shows nothing without the vault.
- **hub**: one session per enabled account, reconnecting with backoff. Session events go into the
  store and then to subscribers. History is read from the cache, and gaps are filled from the
  provider.
- **provider**: what a messenger plugs into. `Info` + `Login` (a step-by-step sign-in dialog that
  the app renders generically) + `Session` (`Run` for live events, plus capability interfaces such
  as `Messenger`).

There is no separate daemon: the core runs inside the app's process on every platform. On Linux
it is linked into the `megachat` executable (`desktopbind` built as a C archive), on Windows it is
`megachat_core.dll` next to `megachat.exe`, both called through Dart FFI; on Android it is a
gomobile AAR kept alive by a foreground service that also shows notifications. The accounts stay
connected while the app runs, also when its window is hidden in the tray; quitting disconnects
them. A second instance of the app uses the core of the first one.

## Messengers

| | Sign-in | Live events | Notes |
|---|---|---|---|
| Telegram | QR code scanned in Telegram on the phone (Settings → Devices → Link Desktop Device), or phone → code; then the cloud password if set | MTProto updates with gap recovery (getDifference survives restarts) | needs app credentials from my.telegram.org: build with `TG_APP_ID`/`TG_APP_HASH`, otherwise the sign-in asks for them |
| Slack — browser session | workspace + `xoxc-` token + `d` cookie from the web client | RTM websocket | works in any workspace without admin approval; unofficial |
| Slack — OAuth app | your app's client id/secret, approval in the browser | Socket Mode with an `xapp-` token; without one, the 20 most active chats are polled | redirect to `http://localhost:53682/slack/callback`, or paste the address you land on |
| VK Teams — login/password | server, login, password, client key | fetchEvents long-poll | the password is used once; a token renews sessions |
| VK Teams — web session | the address of a `fetchEvents` request of the web client (DevTools → Network; it carries server, API version, node and `aimsid`) | fetchEvents long-poll | ends when the web session ends; the desktop client's logs mask the aimsid |

**VK Teams caveat:** VK publishes no documentation for the user-level client protocol. The
implementation follows what VK WorkSpace client logs show (`/api/vNN/rapi/<method>`, aimsid,
`status.code 20000`) and the ICQ/AIM WIM design that VK Teams descends from. All wire details
(paths, field names, signing) are in `core/providers/vkteams/wire.go` and must be checked against
real client traffic before use. The server host and API version can be set per account
(on-premises installations).

## Mail and calendars

| | Sign-in | Mail | Calendar | Live changes |
|---|---|---|---|---|
| Google | OAuth in the browser (loopback redirect + PKCE); refresh token in the vault | Gmail API (labels as folders, All mail as archive) | Google Calendar API | Gmail history every minute, calendars every 3 minutes |
| Mail.ru / VK WorkSpace | address + **app password** (mailbox settings → Security → passwords for external apps) | IMAP `imap.mail.ru:993`, SMTP `smtp.mail.ru:465` | CalDAV `calendar.mail.ru` (discovered) | inbox counters every minute, calendar etags every 5 minutes |
| Own servers (on-premises VK WorkSpace, any other) | address, password, IMAP/SMTP hosts, optional CalDAV URL | IMAP/SMTP (993/465 TLS, 587 STARTTLS) | CalDAV | same |

- Nothing of the mail is cached on disk: lists and bodies come from the providers on every read,
  several accounts at once, and the hub merges them (newest first; an opaque cursor pages through
  all accounts together). Attachments go through `Media` and its sealed cache. Accounts that fail
  are reported next to the results; the others still show.
- Messages are shown as text; an HTML-only message is turned into text in the core, remote
  content is never loaded.
- Repeating CalDAV events are expanded by the core (RRULE, EXDATE, RDATE, moved occurrences).
  Single occurrences of a repeating CalDAV event cannot be edited or deleted yet; Google's can.
- Google needs an OAuth client of type "Desktop app" (Gmail API and Calendar API enabled; for an
  unverified app add yourself as a test user): build with `GOOGLE_CLIENT_ID`/`GOOGLE_CLIENT_SECRET`,
  otherwise the sign-in asks for them.
- **Mail.ru caveat:** the CalDAV address and that Mail.ru keeps a copy of SMTP-sent mail in Sent
  follow Mail.ru's help pages; the tests run against local IMAP, SMTP and CalDAV servers, not the
  live service.

## Build

Go 1.27, Flutter 3.41, protoc + protoc-gen-go/-go-grpc + protoc_plugin (Dart), gomobile and an
Android NDK for Android.

```bash
make build-linux                             # builds/linux/megachat: the app with the core in it
make build-linux TG_APP_ID=123 TG_APP_HASH=abc   # with built-in Telegram app credentials
make build-linux GOOGLE_CLIENT_ID=… GOOGLE_CLIENT_SECRET=…   # with a built-in Google OAuth client
make build                                   # builds/mchat (CLI)
make build-windows                           # builds/windows/mchat.exe; the app: flutter build windows
                                             # on Windows (Go + MSYS2 gcc build megachat_core.dll)
make build-apk                               # builds/android/*.apk (arm64, x86_64)
make test race vet app-test                  # tests are capped at 1 GB, one run at a time
make proto                                   # after changing proto/megachat.proto
```

## Use

`mchat` talks to the core of the running app (the cache belongs to one process):

```bash
builds/linux/megachat &
mchat create                  # new vault (asks for a password); later: mchat unlock [remembered]
mchat providers
mchat login telegram          # the same sign-in dialog the app shows
mchat login slack session
mchat accounts
mchat chats -n 20
mchat history ACCOUNT CHAT -n 50
mchat send ACCOUNT CHAT hello
mchat sendfile ACCOUNT CHAT ./photo.jpg caption...
mchat login google            # or: mchat login mailru [server]
mchat folders
mchat mail -n 20              # all inboxes; mchat mail ACCOUNT FOLDER for one folder
mchat mailread ACCOUNT ID
mchat sendmail ACCOUNT to@example.com "Subject" text...
mchat events -n 7             # the next 7 days of every calendar
mchat watch                   # live events
```

`make run-demo` runs the app with the **Demo** messenger (`MEGACHAT_DEMO=1`), an in-memory
provider with generated chats. Use it to work on the app without real accounts.

Files (desktop): `~/.config/megachat/` holds `vault.mc`, `cache-*.db`, `endpoint.json` (while the
app runs) and `megachat.log`.

## License

MegaChat is free software: you can redistribute it and/or modify it under the terms of the
GNU General Public License, version 3 or (at your option) any later version. See
[LICENSE](LICENSE). The bundled fonts (JetBrains Mono, Syne) are under the SIL Open Font
License (`app/assets/fonts/OFL-*.txt`).
