package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"fmt"
	"github.com/shr-go/bili_live_tui/api"
	"github.com/shr-go/bili_live_tui/internal/live_room"
	"net/http"
	"os"
	"time"
)

type loginStep uint8

const (
	loginStepConfirmLogin loginStep = iota
	loginStepWaitLogin
	loginStepLoginNeedRefresh
	loginStepLoginSuccess
	loginStepDone
)

type loginModel struct {
	step        loginStep
	client      *http.Client
	loginData   *api.QRCodeLoginData
	room        *api.LiveRoom
	cookies     string
	chooseLogin bool
	localCookie bool
	quit        bool
	errText     string
	width       int
	height      int
}

func newLoginModel(client *http.Client) loginModel {
	return loginModel{
		step:        loginStepConfirmLogin,
		client:      client,
		loginData:   nil,
		room:        nil,
		cookies:     "",
		chooseLogin: true,
		localCookie: false,
		quit:        false,
		width:       windowWidth,
		height:      windowHeight,
	}
}

type waitScanMsg struct{}

type TickMsg time.Time

type loginDataMsg struct {
	data *api.QRCodeLoginData
	err  error
}
type loginPollMsg struct {
	data    api.QRCodeLoginData
	cookies string
	err     error
}
type loginRoomMsg struct {
	room *api.LiveRoom
	err  error
}

func (m *loginModel) loadLoginData() tea.Cmd {
	client := m.client
	return func() tea.Msg {
		data, err := live_room.QRCodeLogin(client)
		return loginDataMsg{data, err}
	}
}

func tickEvery() tea.Cmd {
	return tea.Every(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m *loginModel) pollLoginStatus() tea.Cmd {
	client, data := m.client, *m.loginData
	return func() tea.Msg {
		cookies, err := live_room.PollLogin(client, &data)
		return loginPollMsg{data, cookies, err}
	}
}

func (m *loginModel) enterRoom() tea.Cmd {
	client, cookies, saveCookie, roomID := m.client, m.cookies, m.chooseLogin && !m.localCookie, LiveConfig.RoomID
	return func() tea.Msg {
		if saveCookie {
			if !live_room.CheckCookieValid(client, cookies) {
				return loginRoomMsg{err: fmt.Errorf("登录失效，请重启后重新登录")}
			}
			if err := os.WriteFile("COOKIE.DAT", []byte(cookies), 0o600); err != nil {
				return loginRoomMsg{err: fmt.Errorf("保存登录信息失败: %w", err)}
			}
		}
		room, err := live_room.AuthAndConnect(client, roomID)
		return loginRoomMsg{room, err}
	}
}

func (m *loginModel) Init() tea.Cmd {
	if m.step == loginStepLoginSuccess {
		return m.enterRoom()
	}
	return nil
}

func (m *loginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if keyString(msg) == "ctrl+c" {
			m.quit = true
			return m, tea.Quit
		}
		if m.errText != "" && keyString(msg) == "enter" {
			m.errText = ""
			if m.step == loginStepLoginSuccess {
				return m, m.enterRoom()
			}
			m.step = loginStepWaitLogin
			return m, m.loadLoginData()
		}
		switch m.step {
		case loginStepConfirmLogin:
			switch keyString(msg) {
			case "tab":
				m.chooseLogin = !m.chooseLogin
			case "left":
				m.chooseLogin = true
			case "right":
				m.chooseLogin = false
			case "enter", "space":
				if m.chooseLogin {
					m.step = loginStepWaitLogin
					return m, m.loadLoginData()
				} else {
					m.step = loginStepLoginSuccess
					return m, m.enterRoom()
				}
			}
		case loginStepLoginNeedRefresh:
			if keyString(msg) == "enter" || keyString(msg) == "space" {
				m.step = loginStepWaitLogin
				m.loginData = nil
				return m, m.loadLoginData()
			}
		}
		return m, nil
	case loginDataMsg:
		if msg.err != nil {
			m.errText = msg.err.Error()
			return m, nil
		}
		m.loginData = msg.data
		return m, tickEvery()
	case TickMsg:
		if m.loginData == nil || m.step != loginStepWaitLogin {
			return m, nil
		}
		return m, m.pollLoginStatus()
	case loginPollMsg:
		if msg.err != nil {
			m.errText = msg.err.Error()
			return m, nil
		}
		m.loginData = &msg.data
		switch msg.data.Status {
		case api.QRLoginExpired:
			m.step = loginStepLoginNeedRefresh
		case api.QRLoginSuccess:
			m.step, m.cookies = loginStepLoginSuccess, msg.cookies
			return m, m.enterRoom()
		default:
			return m, tickEvery()
		}
	case loginRoomMsg:
		if msg.err != nil {
			m.errText = msg.err.Error()
			return m, nil
		}
		m.room, m.step = msg.room, loginStepDone
		return m, tea.Quit
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		windowWidth, windowHeight = msg.Width, msg.Height
	case loginStep:
		if msg == loginStepWaitLogin {
			return m, tickEvery()
		} else if msg == loginStepLoginSuccess {
			return m, m.enterRoom()
		} else if msg == loginStepDone {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *loginModel) View() tea.View {
	return screen(m.place(m.loginBody()))
}

func (m *loginModel) textWidth() int {
	w := 50
	if m.width > 4 && m.width-4 < w {
		w = m.width - 4
	}
	if w < 1 {
		w = 1
	}
	return w
}

func (m *loginModel) place(content string) string {
	if content == "" || m.width <= 0 || m.height <= 0 {
		return content
	}
	if lipgloss.Width(content) > m.width || lipgloss.Height(content) > m.height {
		return fitBlock(content, m.width, m.height)
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content,
		lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Foreground(subtle)))
}

func (m *loginModel) loginBody() string {
	if m.errText != "" {
		return dialogBoxStyle.Render(singleLine(m.errText) + " · Enter 重试 · Ctrl+C 退出")
	}
	switch m.step {
	case loginStepConfirmLogin:
		var loginButton, cancelButton string
		if m.chooseLogin {
			loginButton = activeButtonStyle.Render("扫码登录")
			cancelButton = buttonStyle.Render("取消")
		} else {
			loginButton = buttonStyle.Render("扫码登录")
			cancelButton = activeButtonStyle.Render("取消")
		}
		question := lipgloss.NewStyle().Width(m.textWidth()).Align(lipgloss.Center).
			Render("扫码登陆后才能发送弹幕哦！")
		buttons := lipgloss.JoinHorizontal(lipgloss.Top, loginButton, "  ", cancelButton)
		ui := lipgloss.JoinVertical(lipgloss.Center, question, buttons)
		return dialogBoxStyle.Render(ui)
	case loginStepWaitLogin:
		if m.loginData != nil {
			return m.qrDialog()
		}
	case loginStepLoginNeedRefresh:
		question := lipgloss.NewStyle().Width(m.textWidth()).Align(lipgloss.Center).
			Render("二维码已过期，请刷新后再试")
		ui := lipgloss.JoinVertical(lipgloss.Center, question, activeButtonStyle.Render("刷新"))
		return dialogBoxStyle.Render(ui)
	case loginStepLoginSuccess:
		str := "登陆成功，正在连接服务器"
		if !m.chooseLogin {
			str = "以游客身份登录，正在连接服务器"
		}
		text := lipgloss.NewStyle().Width(m.textWidth()).Align(lipgloss.Center).Render(str)
		return dialogBoxStyle.Render(text)
	}
	return ""
}

func (m *loginModel) qrDialog() string {
	tips := "扫描二维码，或扫描目录中的 login.png"
	if m.loginData.Status == api.QRLoginNotConfirm {
		tips = "请在手机上点击确定完成登录"
	}
	fallback := dialogBoxStyle.Render("请扫描 login.png")
	qr := renderQR(m.loginData.Modules)
	if qr == "" {
		return fallback
	}
	ui := lipgloss.JoinVertical(lipgloss.Center, tips, qr)
	box := dialogBoxStyle.Copy().Padding(0, 0).Render(ui)
	if m.width > 0 && m.height > 0 && (lipgloss.Width(box) > m.width || lipgloss.Height(box) > m.height) {
		return fallback
	}
	return box
}
