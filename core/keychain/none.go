//go:build android || !(linux || windows || darwin)

package keychain

// System returns nil: no keychain here (on Android the app seals the
// password in the Android Keystore behind a fingerprint instead).
func System() Keychain { return nil }
