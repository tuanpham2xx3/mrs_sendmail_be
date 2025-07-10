package services

import (
	"fmt"

	"gopkg.in/gomail.v2"
	"mrs_sendemail_be/internal/config"
)

type SMTPService struct {
	config *config.Config
	dialer *gomail.Dialer
}

func NewSMTPService(cfg *config.Config) *SMTPService {
	dialer := gomail.NewDialer(
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.Username,
		cfg.SMTP.Password,
	)

	return &SMTPService{
		config: cfg,
		dialer: dialer,
	}
}

// TestConnection kiểm tra kết nối SMTP
func (s *SMTPService) TestConnection() error {
	sender, err := s.dialer.Dial()
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer sender.Close()
	return nil
}

// SendVerificationEmail gửi email chứa mã xác thực
func (s *SMTPService) SendVerificationEmail(email, code, system string, customData map[string]interface{}) error {
	if system == "" {
		system = s.config.Code.DefaultSystemName
	}

	subject := fmt.Sprintf("Mã xác thực cho %s", system)
	body := s.generateEmailBody(code, system, customData)

	message := gomail.NewMessage()
	message.SetHeader("From", message.FormatAddress(s.config.SMTP.Username, s.config.SMTP.FromName))
	message.SetHeader("To", email)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", body)

	if err := s.dialer.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// generateEmailBody tạo nội dung HTML cho email
func (s *SMTPService) generateEmailBody(code, system string, customData map[string]interface{}) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="vi">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Mã xác thực</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f4f4f4;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 10px;
            box-shadow: 0 0 20px rgba(0,0,0,0.1);
        }
        .header {
            text-align: center;
            margin-bottom: 30px;
        }
        .logo {
            font-size: 24px;
            font-weight: bold;
            color: #2c3e50;
            margin-bottom: 10px;
        }
        .code-container {
            background: #f8f9fa;
            border: 2px solid #e9ecef;
            border-radius: 8px;
            padding: 20px;
            text-align: center;
            margin: 20px 0;
        }
        .verification-code {
            font-size: 32px;
            font-weight: bold;
            color: #007bff;
            letter-spacing: 8px;
            margin: 10px 0;
        }
        .warning {
            background: #fff3cd;
            border: 1px solid #ffeaa7;
            border-radius: 5px;
            padding: 15px;
            margin: 20px 0;
            color: #856404;
        }
        .footer {
            margin-top: 30px;
            padding-top: 20px;
            border-top: 1px solid #eee;
            text-align: center;
            color: #666;
            font-size: 14px;
        }
        .highlight {
            color: #007bff;
            font-weight: bold;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">%s</div>
            <h2>Mã xác thực đăng nhập</h2>
        </div>
        
        <p>Xin chào,</p>
        <p>Bạn đã yêu cầu mã xác thực để đăng nhập vào hệ thống <span class="highlight">%s</span>.</p>
        
        <div class="code-container">
            <p><strong>Mã xác thực của bạn là:</strong></p>
            <div class="verification-code">%s</div>
            <p><small>Mã này có hiệu lực trong vòng <span class="highlight">%d phút</span></small></p>
        </div>
        
        <div class="warning">
            <strong>⚠️ Lưu ý bảo mật:</strong>
            <ul style="margin: 10px 0; padding-left: 20px;">
                <li>Không chia sẻ mã này với bất kỳ ai</li>
                <li>Mã chỉ sử dụng một lần và sẽ hết hạn sau %d phút</li>
                <li>Nếu bạn không yêu cầu mã này, vui lòng bỏ qua email</li>
            </ul>
        </div>
        
        <p>Nếu bạn gặp khó khăn trong việc đăng nhập, vui lòng liên hệ với đội ngũ hỗ trợ.</p>
        
        <div class="footer">
            <p>Email này được gửi tự động từ hệ thống <strong>%s</strong></p>
            <p>Vui lòng không trả lời email này.</p>
        </div>
    </div>
</body>
</html>
    `,
		system,                          // Logo/header
		system,                          // Tên hệ thống trong nội dung
		code,                            // Mã xác thực
		s.config.Code.ExpireMinutes,     // Thời gian hết hạn (1)
		s.config.Code.ExpireMinutes,     // Thời gian hết hạn (2)
		system,                          // Footer
	)
} 