package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os/exec"
	"strings"
	"syscall"
)

// Windows: a toast through Windows PowerShell's WinRT bridge (no extra
// DLLs or COM registration). It goes under PowerShell's app id, as an
// unpackaged app has none of its own; a click does not reach the app.
const toastScript = `
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] > $null
[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] > $null
$in = [Console]::In.ReadToEnd().Split([char]0)
$x = New-Object Windows.Data.Xml.Dom.XmlDocument
$x.LoadXml($in[1])
$t = New-Object Windows.UI.Notifications.ToastNotification $x
$t.Tag = $in[0]
$t.Group = 'parley'
$id = '{1AC14E77-02E7-4E5D-B744-2EB1AE5198B7}\WindowsPowerShell\v1.0\powershell.exe'
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier($id).Show($t)
`

func xmlText(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(s)
}

// toastTag: Windows caps a tag at 64 characters.
func toastTag(tag string) string {
	h := sha256.Sum256([]byte(tag))
	return hex.EncodeToString(h[:16])
}

func showNotice(n notice) {
	x := `<toast><visual><binding template="ToastGeneric"><text>` + xmlText(n.Title) +
		`</text><text>` + xmlText(n.Body) + `</text></binding></visual></toast>`
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", toastScript)
	cmd.Stdin = strings.NewReader(toastTag(n.Tag) + "\x00" + x)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000} // CREATE_NO_WINDOW
	_ = cmd.Run()
}

func closeNotice(string) {}
