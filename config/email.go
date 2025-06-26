package config

type EmailConfig struct {
	SenderName  string
	SenderEmail string
}

func NewEmailConfig(name, email string) EmailConfig {
	return EmailConfig{
		SenderName:  name,
		SenderEmail: email,
	}
}
