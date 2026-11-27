# CloudSync Connector API — integration guide for apps

**API version 1.** This file is self-contained: copy it into an app's repository
and build the integration from it. The contract files are at the end
([Appendix A](#appendix-a-connectorproto), [Appendix B](#appendix-b-aidl-android)).
Their canonical copies live in the CloudSync repository (`proto/connector.proto`,
`app/android/app/src/main/aidl/dev/cloudsync/connector/*.aidl`). They must
match those copies byte for byte.

## 1. What it gives an app

An app that holds its settings in a file (MegaChat: `vault.mc`) can add a
**"Sync via CloudSync"** switch. With the switch on, the app gives the file to
CloudSync, and CloudSync:

- stores it encrypted in **every storage** of the user's vault, with a copy on
  every device;
- gives **every device the same value**. A change made on one device reaches
  the others at their next sync (about 30 s, or right away through `Watch`);
- **never loses a write**. Writes made at the same time on two offline devices
  are both kept, and the app merges them (§ 6);
- protects the file with **the vault's own password**. The app can drop its own
  password, because CloudSync hands it a stable per-app key (`AppKey`, § 8).

The app must keep working **without CloudSync**. Sync is optional, and the
app's existing standalone mode stays as it is. CloudSync is a separate
program: it may be missing, stopped or locked at any moment.

What an app can reach: **only its own records**. A paired app cannot see the
user's files, passwords, storages or other apps' records.

## 2. Concepts

| Term | Meaning |
|---|---|
| **App id** | Stable name of the app and of its data, the same on every platform and device: `^[a-z0-9][a-z0-9._-]{1,63}$`, e.g. `megachat`. Never change it: other ids see other data. |
| **Pairing** | The user's permission for this app **on this device**. It is made once per device and the user can revoke it in CloudSync. |
| **Record** | A named value of the app: key `^[A-Za-z0-9._/-]{1,200}$` (no `..`), value up to **256 KiB**, at most **256 records** per app. Every write keeps the whole value in the vault's journals forever, so records suit settings and vault files. They are not for message caches or media. |
| **Revision** | Opaque string naming the record's current state. Pass it as `base` to the next `Put`/`Delete` (compare-and-swap). `""` means the record was never written. |
| **Versions** | The current values of a record: none (absent or deleted), one (normal), or several (concurrent writes; the app must merge them). |
| **Vault state** | `NO_VAULT`, `LOCKED`, `OFFLINE` (unlocked, no storage reachable: reads work from the device copy, writes fail), `READY`. |

CloudSync stores a record as an operation in its per-device journals, sealed
with a key derived from the vault master key and the app id. Storages see only
ciphertext.

## 3. Finding CloudSync and talking to it

### 3.1 Desktop (Linux, Windows, macOS)

CloudSync's daemon `cloudsyncd` serves the Connector API on a socket of its
own and writes its address into **`connector.json`**:

| OS | Path |
|---|---|
| Linux | `$XDG_CONFIG_HOME/cloudsync/connector.json` (default `~/.config/cloudsync/connector.json`) |
| Windows | `%AppData%\cloudsync\connector.json` |
| macOS | `~/Library/Application Support/cloudsync/connector.json` |

(Go: `os.UserConfigDir()` + `cloudsync/connector.json`.)

```json
{ "address": "unix:///run/user/1000/cloudsync-connector.sock", "api_version": 1 }
```

- `address` is either `unix:///absolute/path` (Linux, macOS) or `host:port` on
  `127.0.0.1` (Windows). It is plain gRPC (HTTP/2) without TLS. Dial the unix
  socket directly (Go: `grpc.NewClient("passthrough:///unix", grpc.WithContextDialer(...))`,
  or the `unix:` scheme of your gRPC library).
- The file exists only while the daemon runs. A file that is missing, or an
  address that refuses connections, means "CloudSync is not running": stay in
  standalone mode and look again later (a file watcher or a retry every 30–60 s).
- The file holds no secret. Authentication is the per-app **bearer token**
  from pairing, sent on every call except `Hello`, `RequestPairing` and
  `AwaitPairing` as gRPC metadata `authorization: Bearer <token>`.
- Check `api_version` (in the file and in `HelloReply`). Version 1 only gains
  fields and methods. A bigger number means an incompatible API: tell the user
  to update the app.

### 3.2 Android

CloudSync (package **`dev.cloudsync.cloudsync`**) exports a bound service.

1. Copy the two AIDL files from [Appendix B](#appendix-b-aidl-android) into
   `src/main/aidl/dev/cloudsync/connector/`, keeping the package name, and
   enable `buildFeatures { aidl = true }`.
2. Declare package visibility in the app's manifest (Android 11+):
   ```xml
   <queries>
     <package android:name="dev.cloudsync.cloudsync"/>
   </queries>
   ```
3. Bind with an explicit intent:
   ```kotlin
   val i = Intent("dev.cloudsync.connector.BIND").setPackage("dev.cloudsync.cloudsync")
   context.bindService(i, conn, Context.BIND_AUTO_CREATE)
   // onServiceConnected: val api = ICloudSyncConnector.Stub.asInterface(binder)
   ```
   If `bindService` returns `false`, CloudSync is not installed: stay standalone.
4. Call methods as `api.call("Get", GetRequest.toByteArray())`. The reply is
   always a serialized **`CallReply`**: `code` 0 means OK and `payload` holds
   the reply message; otherwise `code` is a gRPC status code (§ 9) and
   `message` the text. `call` may block (`AwaitPairing` up to 60 s), so never
   call it on the main thread.
5. `api.watch(watcher)` streams serialized `Change` messages to
   `IConnectorWatcher.onChange`, then one `onEnd(CallReply)`. Stop it with
   `api.unwatch(handle)`. CloudSync stops it by itself when your process dies.
6. There is **no token on Android**. CloudSync identifies the caller through
   Binder (calling package + SHA-256 of its signing certificate). The same
   package signed with another key is another app and has to pair again.
   `AwaitPairing` answers `APPROVED` with an empty token.
7. To let the user unlock CloudSync or answer a request, open it with
   `packageManager.getLaunchIntentForPackage("dev.cloudsync.cloudsync")`.
   CloudSync also posts a notification when a pairing request is waiting.

Generate message classes from `connector.proto` (e.g. protobuf-javalite, or
Wire). Only messages are needed on Android. The `service` block is for the
desktop gRPC client.

## 4. `Hello` and the vault state

Call `Hello{app_id}` first, after every reconnect, and after a `STATE` change.

| `state` | Meaning for the app |
|---|---|
| `NO_VAULT` | CloudSync is installed but not set up. Stay standalone. |
| `LOCKED` | Ask the user to unlock CloudSync (on Android, offer to open it). Get/Put fail with `FAILED_PRECONDITION`. |
| `OFFLINE` | Unlocked, but no storage answers yet. `Get`/`List`/`AppKey` work from this device's copy, `Put`/`Delete` fail with `UNAVAILABLE`. Queue writes and retry. |
| `READY` | Everything works. |

`paired` says whether this caller is paired for `app_id`. `vault_id` changes
when the user opens a different vault. Treat that as a new, empty remote:
never merge data across vault ids.

## 5. Pairing

```
app                                   CloudSync                           user
 │ RequestPairing{app_id,name,desc} ──▶│ keeps the request (5 min)         │
 │◀── PairingTicket{ticket, code}      │── dialog / notification ─────────▶│
 │ shows "Code 123 456 — confirm in CloudSync"                             │
 │ AwaitPairing{ticket, wait 60} ─────▶│◀──── Allow (same code?) ──────────│
 │◀── PairingResult{APPROVED, token}   │                                   │
```

1. `RequestPairing{app_id, app_name, description}`. Put the app name, version
   and platform in `description` (e.g. `"MegaChat 1.4 on linux"`).
2. Show `code` (6 digits, e.g. as `123 456`) and say "Confirm in CloudSync".
   CloudSync shows the same code to the user, so the user can tell that the
   request is really yours.
3. Loop `AwaitPairing{ticket, wait_seconds: 60}` while it returns `PENDING`.
   - `APPROVED`: on desktops, store `token` **on this device only**: the OS
     keychain if the app has one, else a file readable only by the user. Never
     put it into the synced data: it is valid on one device only. The token
     comes back **exactly once**.
   - `DENIED`: tell the user. Do not ask again on your own.
   - `EXPIRED`: the request timed out (5 min), was replaced by a newer request
     of the same app, or was already answered. Offer to try again.
4. A later `UNAUTHENTICATED` from any call, or a `Change{UNPAIRED}`, means the
   pairing is gone (revoked in CloudSync, or CloudSync data reset). Switch sync
   off, keep the local file, and offer to pair again.

Pairing again replaces the app's old pairing on that device. `Unpair` removes
the pairing from the app's side; use it when the user switches sync off.
Records are not deleted by unpairing.

## 6. Reading and writing records

### 6.1 Compare-and-swap

```
item = Get{key}                     // item.revision, item.versions
value = merge(item.versions) ...    // see 6.2
Put{key, data: new, base: item.revision}
  OK       → keep the returned item.revision as your new base
  ABORTED  → someone wrote meanwhile: Get again, merge, Put again
```

- `Put` to a record that does not exist uses `base: ""`. If the record exists
  (or was deleted, which leaves a non-empty revision), you get `ABORTED`: Get
  it first.
- `Delete{key, base}` follows the same rule. After a delete, `Get` returns no
  versions and a non-empty revision. Write again from that revision.
- `ABORTED` covers only writes this device already knows about. Writes from
  other devices that have not synced yet cannot be checked at write time. They
  show up later as **several versions** of the record (§ 6.2).

### 6.2 Several versions: the app merges

`Item.versions` has more than one entry when devices wrote the record from the
same revision (typically offline). CloudSync cannot merge them because it does
not understand the content. The app must:

1. Decode every version.
2. Merge them **deterministically**. Every device may run the merge on the
   same input, so the result must not depend on which device runs it or in
   what order. Sort versions by `id` before folding.
3. `Put` the merged value with `base: item.revision`. That single write
   replaces all the versions on every device.

Merge rules that work for settings files:

- Collections of items with ids (accounts, servers): union by id. Keep
  **tombstones** (`removed: [{id, time}]`) so that a removal on one device is
  not undone by the other device's copy. Drop tombstones older than, say,
  90 days.
- Scalar settings: last writer wins by a per-field timestamp, or else by
  `Version.time_unix_ms`, ties broken by `Version.id`.
- Never discard a version you cannot decode: keep the record as is and tell
  the user.

### 6.3 Watching

`Watch` sends nothing at first. Open it, then `Get`/`List` what you need, then
apply changes as they come:

- `RECORD{key, revision, deleted}`: a record changed (on any device). If
  `revision` differs from what you hold, `Get` it.
- `STATE{state}`: the vault state changed: call `Hello`.
- `UNPAIRED`: see § 5, step 4. The stream ends after it.

The stream also ends when CloudSync stops or locks the connection. Reconnect
with backoff, call `Hello`, then `List` to catch up. Changes are not queued
while you are disconnected.

## 7. Recommended integration: one settings file

The usual case: the app has a file `F` (MegaChat: `vault.mc`), and `F`
becomes record `key = "vault.mc"`.

Keep **local sync state** per device, outside `F`: `enabled`, `token`
(desktop), `vault_id`, `base` (the revision `F` was last synced at) and
`synced_hash` (hash of `F` at that moment).

**Switching sync on** ("Sync via CloudSync" button, shown when CloudSync is
found and `state` is not `NO_VAULT`):

1. Pair (§ 5) if not paired.
2. `Get{F}`:
   - no versions → `Put{F, local bytes, base: revision}`. This device seeds the vault.
   - versions exist and equal the local bytes → nothing to do.
   - versions exist and differ → ask the user: **use the synced settings**
     (replace local, after a backup copy), **upload this device's settings**
     (`Put` with that revision), or, if the app supports it, **merge**.
3. Save `base` and `synced_hash`.

**Running:**

- The local `F` stays the working copy. The app reads and writes it as before,
  so it works when CloudSync is gone.
- Local change (`hash(F) != synced_hash`) → `Put{F, bytes, base}`.
  `ABORTED` → Get, merge, Put. `UNAVAILABLE`/`FAILED_PRECONDITION` → keep it
  pending and retry on the next `STATE` change.
- Remote change (`Watch` RECORD, or `Get` on start shows `revision != base`):
  - local unchanged since the last sync → write the remote bytes to `F`
    atomically (temp file + rename) and reload;
  - local changed too → merge both, write `F`, `Put` with the new revision;
  - several versions → merge (§ 6.2).
- On start: `Hello` → if paired and `vault_id` matches → `Get` and reconcile as above.

**Switching sync off:** keep `F`, call `Unpair`, clear the local sync state.
The record stays in the vault for the other devices. Deleting it is a separate
action ("Remove synced settings from CloudSync": `Delete`).

**`vault_id` changed:** the user opened another vault. Stop syncing and ask
whether to switch sync on for the new vault (as above).

## 8. No password of the app's own (`AppKey`)

`AppKey` returns 32 bytes that are the same for this app on every device of
the vault and never change (HKDF of the vault master key and the app id).
`key_id` is a short fingerprint that tells keys apart.

So the app does not need a password of its own while sync is on:

- **Synced copy.** The record is already encrypted by CloudSync. The app may
  store its plain settings there, or its own file sealed with `AppKey`.
  Sealing with `AppKey` keeps every copy, local and synced, in one format.
- **Local copy.** Seal `F` with `AppKey` (e.g. XChaCha20-Poly1305 with an AD
  of your choice). To open it when CloudSync is not running, keep `AppKey` in
  the OS keychain (desktop) or wrapped by the Android Keystore. Without such a
  cache the app needs CloudSync unlocked to start.
- **Migration from a password-protected file.** When sync is switched on, the
  user enters the old password once. The app re-seals `F` with `AppKey` and
  marks that the file is now "protected by CloudSync". Switching sync off must
  offer to set a password again before the app forgets `AppKey`.
- **Other devices** get `F` from CloudSync and open it with the same `AppKey`,
  without a password.

Never log or upload `AppKey`. It is as sensitive as the data it opens.

## 9. Errors

| gRPC code | When | What to do |
|---|---|---|
| `UNAUTHENTICATED` (16) | not paired, token unknown, pairing revoked | switch sync off, offer to pair again |
| `ABORTED` (10) | `base` is not the current revision | Get, merge, retry |
| `FAILED_PRECONDITION` (9) | CloudSync is locked | wait for `STATE`; offer to open CloudSync |
| `UNAVAILABLE` (14) | unlocked but offline (no storage yet), or the core is not running (Android) | keep writes pending, retry |
| `INVALID_ARGUMENT` (3) | bad app id / key, value > 256 KiB, > 256 records | a bug in the app |
| `RESOURCE_EXHAUSTED` (8) | too many pairing requests waiting | try later |
| `UNIMPLEMENTED` (12) | method unknown to this CloudSync | CloudSync too old: tell the user |

A transport error (connection refused, socket gone, `DeadObjectException`)
means CloudSync stopped: fall back to standalone and reconnect later.

## 10. Security notes

- Desktop: the connector socket is readable only by the user (unix) or bound to
  `127.0.0.1` (Windows). Any local process may *ask* for pairing, so the user
  compares the code before allowing. The app's token is stored by CloudSync
  only as SHA-256.
- Android: the caller is identified by Binder, never by data the caller sends.
- An app sees only records under its own app id. File, password safe, storage
  and vault calls are not part of this API.
- Pairings are per device. Revoking on one device does not affect the others.

## 11. Compatibility

- Devices with an older CloudSync without the connector still sync and keep
  app records (they carry the journal operations without reading them).
  A CloudSync without the connector writes no `connector.json` and has no
  `ConnectorService`: to the app it looks like "CloudSync is not installed or
  not running".
- v1 only gains fields and methods. Unknown fields must be ignored (protobuf
  does this by default). Unknown enum values should be treated as the
  `*_UNSPECIFIED` value.

## Appendix: MegaChat specifics

These notes target the MegaChat code base (`core/vault`, `core/rpc`,
`core/embed`, Flutter `settings`).

- App id `megachat`, record key `vault.mc` (the bytes of the vault file). The
  per-device sync state (token, `base`, `synced_hash`, `vault_id`) goes into
  the data dir (e.g. `<data dir>/cloudsync.json`, token in the OS keychain
  next to the remembered unlock key), never into `vault.Config`.
- `vault.Config.Version` is bumped on every change, but it is not enough to
  merge: two devices both reach version N+1. Merge `Accounts` by `ID`
  (a re-login keeps the id) and add tombstones for removed accounts
  (`Removed []struct{ID string; At time.Time}`, an optional JSON field that
  older builds ignore). The merged config gets `max(Version)+1`.
- `VaultID` and `MasterKey` are the same on every device that got the file
  from CloudSync, so the cache keyed by `VaultID` keeps working.
- Two versions sealed with different passwords (a password change on one
  device while another was offline): if one does not open with the current
  key, ask for the other password before merging, or let the user choose a
  version.
- Without CloudSync, everything stays as today: `vault.mc` with a password,
  the remembered key in the keychain, `SetVaultPath`.
- With "Protected by CloudSync" (§ 8), `vault.mc` is sealed with `AppKey`
  instead of the Argon2id password key. Keep the file header so the two modes
  can be told apart (e.g. a new magic `MCVAULT\x02` = sealed with `AppKey`,
  `key_id` in the header).

## Appendix A: `connector.proto`

Canonical: `proto/connector.proto` in the CloudSync repository. Go users:
`option go_package` points into CloudSync's module. Override it when generating
your own client (`--go_opt=Mconnector.proto=yourmodule/cloudsyncpb`).

```proto
syntax = "proto3";

// The CloudSync Connector API, version 1: what CloudSync offers to other
// applications ("connected apps", e.g. MegaChat). An app keeps a few small
// records (its settings file, its vault) in CloudSync; CloudSync stores them
// encrypted in every storage and gives every device the same value.
//
// This file is the contract. Apps copy it as is and generate their own
// client; see docs/CONNECTOR.md for the transport (unix socket / localhost
// on desktops, a bound service on Android), pairing and the merge rules.
// Only fields and methods are ever added to v1; nothing is renamed or
// renumbered.
package cloudsync.connector.v1;

option go_package = "cloudsync/core/connector/pb";

service Connector {
  // --- no credentials needed ---------------------------------------------

  // Hello says what is there: API version, whether the vault is open and
  // whether this caller is paired. Call it first and after every reconnect.
  rpc Hello(HelloRequest) returns (HelloReply);
  // RequestPairing asks the user (in CloudSync) to let this app keep its
  // data there. Show the returned code: the user sees the same one.
  rpc RequestPairing(PairingRequest) returns (PairingTicket);
  // AwaitPairing waits (up to wait_seconds, at most 60) for the user's
  // answer. The token comes back exactly once, with APPROVED.
  rpc AwaitPairing(AwaitPairingRequest) returns (PairingResult);

  // --- paired apps only (bearer token on desktops, the calling package on
  // Android); every call works on the caller's own app id only -----------

  // Get returns one record. A record nobody wrote has no versions.
  rpc Get(GetRequest) returns (Item);
  // List returns every record of the app, without data.
  rpc List(ListRequest) returns (ListReply);
  // Put writes a record. base is the revision the new value was made from
  // (from Get, Put or List; "" for a record that does not exist yet). If
  // the record has another revision now, Put fails with ABORTED: read it
  // again, merge, retry.
  rpc Put(PutRequest) returns (Item);
  // Delete removes a record, with the same base rule as Put.
  rpc Delete(DeleteRequest) returns (Item);
  // AppKey returns a 32-byte key of this app, the same on every device of
  // the vault and never changing. An app may seal its local copy with it
  // instead of asking the user for a password of its own.
  rpc AppKey(AppKeyRequest) returns (AppKeyReply);
  // Watch streams changes of the app's records (from any device) and of
  // the vault state. It sends nothing at the start: call Get / List after
  // it is open.
  rpc Watch(WatchRequest) returns (stream Change);
  // Unpair forgets this app's pairing on this device (the data stays).
  rpc Unpair(UnpairRequest) returns (UnpairReply);
}

enum VaultState {
  VAULT_STATE_UNSPECIFIED = 0;
  NO_VAULT = 1; // CloudSync is not set up on this device
  LOCKED = 2;   // the user has to unlock CloudSync
  // Unlocked, but no storage answers yet (or too few are configured):
  // Get and List read this device's own copy, Put and Delete fail with
  // UNAVAILABLE.
  OFFLINE = 3;
  READY = 4;
}

message HelloRequest {
  string app_id = 1; // optional: fills HelloReply.paired for this app
}

message HelloReply {
  uint32 api_version = 1; // 1
  string core_version = 2;
  VaultState state = 3;
  // The caller is paired (desktops: the request carries a valid token;
  // Android: the calling package was approved for app_id).
  bool paired = 4;
  string app_id = 5;     // the app id of the pairing, when paired
  string vault_id = 6;   // changes when another vault is opened
  string device_name = 7;
}

message PairingRequest {
  // Stable id of the app, the same on every platform and device: the data
  // namespace. ^[a-z0-9][a-z0-9._-]{1,63}$, e.g. "megachat".
  string app_id = 1;
  string app_name = 2;    // shown to the user, e.g. "MegaChat"
  string description = 3; // shown to the user, e.g. "MegaChat 1.4 on linux"
}

message PairingTicket {
  string ticket = 1;          // secret: only the requester knows it
  string code = 2;            // 6 digits, shown by the app and by CloudSync
  int64 expires_unix_ms = 3;  // the request is dropped then (5 minutes)
}

message AwaitPairingRequest {
  string ticket = 1;
  uint32 wait_seconds = 2; // 0 = answer at once; at most 60
}

enum PairingStatus {
  PAIRING_STATUS_UNSPECIFIED = 0;
  PENDING = 1;
  APPROVED = 2;
  DENIED = 3;
  EXPIRED = 4; // timed out, already answered, or unknown ticket
}

message PairingResult {
  PairingStatus status = 1;
  // Desktops: the bearer token for every later call; returned once. Keep
  // it on this device only (it is per device, never put it in the synced
  // data). Android: empty, the package itself is the credential.
  string token = 2;
  string app_id = 3;
}

message Version {
  string id = 1;          // id of the write that made it
  bytes data = 2;         // as written by the app (empty in List)
  int64 size = 3;
  string device_name = 4; // the device that wrote it
  int64 time_unix_ms = 5;
}

message Item {
  string key = 1;
  // Pass it as base to the next Put / Delete. Opaque; "" only for a key
  // that was never written.
  string revision = 2;
  // The current value. Empty: the record does not exist (never written or
  // deleted). More than one: devices wrote it concurrently (offline); the
  // app merges them and Puts the result with this revision as base.
  repeated Version versions = 3;
}

message GetRequest {
  string key = 1;
}

message ListRequest {
  string prefix = 1; // "" = all
}

message ListReply {
  repeated Item items = 1; // only existing records; versions without data
}

message PutRequest {
  // 1–200 bytes of [A-Za-z0-9._/-], e.g. "vault.mc", "settings/ui.json".
  string key = 1;
  bytes data = 2; // at most 256 KiB
  string base = 3;
}

message DeleteRequest {
  string key = 1;
  string base = 2;
}

message AppKeyRequest {}

message AppKeyReply {
  bytes key = 1;    // 32 bytes
  string key_id = 2; // short fingerprint, to tell keys apart
}

message WatchRequest {}

message Change {
  enum Kind {
    KIND_UNSPECIFIED = 0;
    RECORD = 1; // a record changed: key, revision (Get it)
    STATE = 2;  // vault state changed: state (call Hello)
    // The pairing was revoked in CloudSync: the stream ends after it.
    UNPAIRED = 3;
  }
  Kind kind = 1;
  string key = 2;
  string revision = 3;
  bool deleted = 4; // RECORD: no versions left
  VaultState state = 5;
}

message UnpairRequest {}

message UnpairReply {}

// Android only: the bound service carries the messages above as bytes
// (see docs/CONNECTOR.md). Every reply is wrapped in a CallReply.
message CallReply {
  int32 code = 1;      // a gRPC status code; 0 = OK
  string message = 2;  // error text when code != 0
  bytes payload = 3;   // the serialized reply message when code == 0
}
```

## Appendix B: AIDL (Android)

Canonical: `app/android/app/src/main/aidl/dev/cloudsync/connector/` in the
CloudSync repository. Put both files into `src/main/aidl/dev/cloudsync/connector/`.

`ICloudSyncConnector.aidl`:

```aidl
// The CloudSync Connector API on Android (docs/CONNECTOR.md). Apps copy
// this file and IConnectorWatcher.aidl as is, in the same package
// (dev.cloudsync.connector), and bind with an explicit intent:
// action "dev.cloudsync.connector.BIND", package "dev.cloudsync.cloudsync".
//
// Requests and replies are serialized messages of proto/connector.proto.
// CloudSync knows the caller from Binder (package + signing certificate):
// there is no token on Android.
package dev.cloudsync.connector;

import dev.cloudsync.connector.IConnectorWatcher;

interface ICloudSyncConnector {
    // Version of this binder interface: 1.
    int version();

    // One unary method of the Connector service ("Hello", "RequestPairing",
    // "AwaitPairing", "Get", "List", "Put", "Delete", "AppKey", "Unpair"):
    // request is the serialized request message, the result a serialized
    // CallReply (code 0 = OK, payload = the reply message). It may block
    // (AwaitPairing up to 60 s): never call it on the main thread.
    byte[] call(String method, in byte[] request);

    // Starts a Watch: watcher gets serialized Change messages, then onEnd.
    // Returns a handle for unwatch.
    long watch(IConnectorWatcher watcher);
    void unwatch(long handle);
}
```

`IConnectorWatcher.aidl`:

```aidl
// Gets the changes of one Watch (see ICloudSyncConnector.aidl).
package dev.cloudsync.connector;

oneway interface IConnectorWatcher {
    // A serialized Change message.
    void onChange(in byte[] change);
    // The watch ended: a serialized CallReply with the reason.
    void onEnd(in byte[] reply);
}
```
