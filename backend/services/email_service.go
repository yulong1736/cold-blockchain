package services

import (
	"cold-chain-trace/backend/config"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

// PasswordResetMailer 发送密码找回邮件
type PasswordResetMailer interface {
	SendPasswordResetCode(toEmail, code string) error
}

type EmailService struct{}

func NewEmailService() *EmailService {
	return &EmailService{}
}

// SendPasswordResetCode 发送密码找回验证码邮件
func (s *EmailService) SendPasswordResetCode(toEmail, code string) error {
	cfg := config.AppConfig.SMTP
	if cfg.Host == "" || cfg.Port == 0 || cfg.User == "" || cfg.Password == "" {
		return fmt.Errorf("smtp 配置不完整")
	}

	from := strings.TrimSpace(cfg.From)
	if from == "" {
		from = cfg.User
	}

	subject := "冷链平台密码重置验证码"
	body := fmt.Sprintf("您的验证码是：%s\n\n验证码 10 分钟内有效，仅可使用一次。\n如果这不是您的操作，请忽略本邮件。", code)
	message := buildSMTPMessage(from, toEmail, subject, body)
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	auth := smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Host)

	if cfg.UseTLS {
		return sendMailWithTLS(addr, cfg.Host, from, toEmail, auth, message, cfg.InsecureSkipVerify)
	}
	return sendMailWithStartTLS(addr, cfg.Host, from, toEmail, auth, message, cfg.InsecureSkipVerify)
}

func buildSMTPMessage(from, to, subject, body string) []byte {
	headers := []string{
		fmt.Sprintf("From: %s", from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}
	return []byte(strings.Join(headers, "\r\n"))
}

func sendMailWithTLS(addr, host, from, to string, auth smtp.Auth, message []byte, insecureSkipVerify bool) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: insecureSkipVerify,
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Quit()

	return sendSMTPEnvelope(client, from, to, auth, message)
}

func sendMailWithStartTLS(addr, host, from, to string, auth smtp.Auth, message []byte, insecureSkipVerify bool) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer client.Quit()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{
			ServerName:         host,
			InsecureSkipVerify: insecureSkipVerify,
		}); err != nil {
			return err
		}
	}

	return sendSMTPEnvelope(client, from, to, auth, message)
}

func sendSMTPEnvelope(client *smtp.Client, from, to string, auth smtp.Auth, message []byte) error {
	if ok, _ := client.Extension("AUTH"); ok && auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(message); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return nil
}
