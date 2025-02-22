package services

import (
    "fmt"
    "net/smtp"
)

type EmailService struct {
    smtpHost     string
    smtpPort     string
    smtpUsername string
    smtpPassword string
    fromEmail    string
}

func NewEmailService(host, port, username, password, fromEmail string) *EmailService {
    return &EmailService{
        smtpHost:     host,
        smtpPort:     port,
        smtpUsername: username,
        smtpPassword: password,
        fromEmail:    fromEmail,
    }
}

func (s *EmailService) SendPasswordResetEmail(toEmail, resetLink string) error {
    auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)
    
    subject := "Password Reset Request"
    body := fmt.Sprintf(`
Hello,

A password reset has been requested for your account. To reset your password, click the link below:

%s

This link will expire in 1 hour. If you didn't request this reset, please ignore this email.

Best regards,
Your Blog Platform Team
`, resetLink)

    msg := fmt.Sprintf("From: %s\r\n"+
        "To: %s\r\n"+
        "Subject: %s\r\n"+
        "Content-Type: text/plain; charset=UTF-8\r\n"+
        "\r\n"+
        "%s", s.fromEmail, toEmail, subject, body)

    addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
    return smtp.SendMail(addr, auth, s.fromEmail, []string{toEmail}, []byte(msg))
}
