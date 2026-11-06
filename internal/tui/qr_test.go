package tui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/shr-go/bili_live_tui/api"
	"github.com/skip2/go-qrcode"
)

func TestQRRowsStayAligned(t *testing.T) {
	q, err := qrcode.New("https://passport.bilibili.com/h5-app/passport/login/scan?navhide=1&qrcode_key=example", qrcode.Low)
	if err != nil {
		t.Fatal(err)
	}
	modules := q.Bitmap()
	rendered := renderQR(modules)
	lines := strings.Split(rendered, "\n")
	if len(lines) < 8 {
		t.Fatal("qr too short")
	}
	want := len(modules[0]) * lipgloss.Width("▀")
	for i, line := range lines {
		if lipgloss.Width(line) != want {
			t.Fatalf("line %d width %d, want %d", i, lipgloss.Width(line), want)
		}
		plain := stripANSI(line)
		if strings.ContainsAny(plain, " █▄") {
			t.Fatalf("line %d mixes narrow spaces with blocks: %q", i, plain)
		}
	}

	placed := lipgloss.Place(want+20, len(lines)+4, lipgloss.Center, lipgloss.Center, rendered)
	col := -1
	for _, line := range strings.Split(placed, "\n") {
		i := strings.Index(line, "▀")
		if i < 0 {
			continue
		}
		got := ansi.StringWidth(line[:i])
		if col < 0 {
			col = got
		}
		if got != col {
			t.Fatalf("placed qr shifted from column %d to %d", col, got)
		}
	}
	if col < 0 {
		t.Fatal("placed qr missing")
	}
}

func TestLoginQRViewFallsBackWhenTerminalIsTooSmall(t *testing.T) {
	q, err := qrcode.New("https://passport.bilibili.com/h5-app/passport/login/scan?navhide=1&qrcode_key=example", qrcode.Low)
	if err != nil {
		t.Fatal(err)
	}
	m := newLoginModel(nil)
	m.step = loginStepWaitLogin
	m.loginData = &api.QRCodeLoginData{Modules: q.Bitmap(), Status: api.QRLoginNotScan}
	m.width, m.height = 120, 60
	wide := m.View().Content
	if !qrRowsAligned(wide) {
		t.Fatal("login qr rows are not aligned")
	}
	m.width, m.height = 20, 10
	narrow := m.View().Content
	if strings.Contains(narrow, "▀") {
		t.Fatal("small terminal still draws a clipped qr")
	}
	if !strings.Contains(narrow, "login.png") {
		t.Fatal("small terminal missing login.png hint")
	}
}

func qrRowsAligned(view string) bool {
	col := -1
	found := false
	for _, line := range strings.Split(view, "\n") {
		i := strings.Index(line, "▀")
		if i < 0 {
			continue
		}
		found = true
		got := ansi.StringWidth(line[:i])
		if col < 0 {
			col = got
		}
		if got != col {
			return false
		}
	}
	return found
}

func stripANSI(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && s[i] != 'm' {
				i++
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}
