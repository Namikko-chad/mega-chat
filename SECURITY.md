# Security policy

Parley holds the credentials of your messenger, mail and calendar accounts, so security reports
are taken seriously.

## Reporting a vulnerability

Please report privately through GitHub:
**[Report a vulnerability](https://github.com/SelfHold/parley/security/advisories/new)**.
Do not open a public issue for a security problem.

Include what is affected, how to reproduce it, and what an attacker gains. You will get an answer
as soon as possible; this is a personal project, so there is no fixed response time. Please give
a reasonable time for a fix before disclosing.

## Scope

In scope:

- Credentials and tokens: anything that leaks account passwords, tokens, cookies, session keys
  or app passwords (to disk, logs, error messages, RPC replies, the UI, other hosts than the
  provider's own, or other apps).
- The vault: weaknesses in its encryption, key derivation, the remembered key in the keychain,
  the Android fingerprint unlock, or merging synced copies.
- The cache: anything readable in `cache-*.db` without the vault, or media written to disk
  without the user's action.
- The local RPC endpoint: reaching the core without its token, or from another user.
- Content handling: remote content loaded from mail, unsafe links opened, link previews reaching
  private addresses.
- Sync through Shard: Parley's side of the Connector API.

Out of scope:

- Vulnerabilities in the messengers' or mail providers' own services.
- The VK and VK Teams providers' use of unofficial client protocols as such (documented in the
  README).
- Attacks that need root or the unlocked user session on the device.
- Shard itself: report those at [SelfHold/shard](https://github.com/SelfHold/shard/security/advisories/new).

## Supported versions

Only the latest `main` is supported; there are no releases yet.
