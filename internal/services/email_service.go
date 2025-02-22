package services

type EmailService struct{}

func NewEmailService() *EmailService {
	return &EmailService{}
}

func (s *EmailService) SendPasswordResetEmail(email, token string) error {
	// TODO: Implement email sending
	return nil
}
