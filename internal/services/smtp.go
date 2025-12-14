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

// SendActivationEmail gửi email chứa liên kết kích hoạt
func (s *SMTPService) SendActivationEmail(email, activationURL, action, system string, customData map[string]interface{}) error {
	if system == "" {
		system = s.config.Code.DefaultSystemName
	}

	var subject, body string
	switch action {
	case "registration":
		subject = fmt.Sprintf("Kích hoạt tài khoản %s", system)
		body = s.generateActivationEmailBody(activationURL, system, "registration", customData)
	case "password_reset":
		subject = fmt.Sprintf("Đặt lại mật khẩu %s", system)
		body = s.generateActivationEmailBody(activationURL, system, "password_reset", customData)
	default:
		subject = fmt.Sprintf("Xác thực email cho %s", system)
		body = s.generateActivationEmailBody(activationURL, system, "verification", customData)
	}

	message := gomail.NewMessage()
	message.SetHeader("From", message.FormatAddress(s.config.SMTP.Username, s.config.SMTP.FromName))
	message.SetHeader("To", email)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", body)

	if err := s.dialer.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to send activation email: %w", err)
	}

	return nil
}

// generateActivationEmailBody tạo nội dung HTML cho activation email
func (s *SMTPService) generateActivationEmailBody(activationURL, system, action string, customData map[string]interface{}) string {
	var title, message, buttonText, buttonColor string
	
	switch action {
	case "registration":
		title = "Kích hoạt tài khoản"
		message = "Cảm ơn bạn đã đăng ký tài khoản. Vui lòng click vào nút bên dưới để kích hoạt tài khoản của bạn:"
		buttonText = "Kích Hoạt Tài Khoản"
		buttonColor = "#28a745"
	case "password_reset":
		title = "Đặt lại mật khẩu"
		// Check if temp password is provided in customData
		if tempPassword, exists := customData["temp_password"]; exists {
			message = fmt.Sprintf("Mật khẩu tạm thời của bạn là: <strong style='font-size: 18px; color: #dc3545; background: #f8f9fa; padding: 8px 12px; border-radius: 4px; font-family: monospace;'>%v</strong><br><br>Click vào nút bên dưới để kích hoạt mật khẩu này. Sau khi đăng nhập, bạn cần vào trang cài đặt để đổi mật khẩu mới:", tempPassword)
			buttonText = "Kích Hoạt Mật Khẩu Tạm Thời"
		} else {
			message = "Bạn đã yêu cầu đặt lại mật khẩu. Click vào nút bên dưới để tạo mật khẩu mới:"
			buttonText = "Đặt Lại Mật Khẩu"
		}
		buttonColor = "#dc3545"
	default:
		title = "Xác thực email"
		message = "Vui lòng click vào nút bên dưới để xác thực địa chỉ email của bạn:"
		buttonText = "Xác Thực Email"
		buttonColor = "#007bff"
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="vi">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
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
        .button-container {
            text-align: center;
            margin: 30px 0;
        }
        .activation-button {
            display: inline-block;
            background: %s;
            color: white;
            padding: 15px 30px;
            text-decoration: none;
            border-radius: 8px;
            font-weight: bold;
            font-size: 16px;
            transition: background-color 0.3s;
        }
        .activation-button:hover {
            opacity: 0.9;
        }
        .url-fallback {
            background: #f8f9fa;
            border: 1px solid #e9ecef;
            border-radius: 5px;
            padding: 15px;
            margin: 20px 0;
            word-break: break-all;
            font-size: 14px;
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
            <h2>%s</h2>
        </div>
        
        <p>Xin chào,</p>
        <p>%s</p>
        
        <div class="button-container">
            <a href="%s" class="activation-button">%s</a>
        </div>
        
        <p><strong>Liên kết này sẽ hết hạn sau <span class="highlight">30 phút</span>.</strong></p>
        
        <p>Nếu bạn không thể click vào nút trên, vui lòng copy và paste URL bên dưới vào trình duyệt:</p>
        <div class="url-fallback">
            %s
        </div>
        
        <div class="warning">
            <strong>⚠️ Lưu ý bảo mật:</strong>
            <ul style="margin: 10px 0; padding-left: 20px;">
                <li>Liên kết này chỉ sử dụng một lần và sẽ hết hạn sau 30 phút</li>
                <li>Không chia sẻ liên kết này với bất kỳ ai</li>
                <li>Nếu bạn không yêu cầu email này, vui lòng bỏ qua</li>
            </ul>
        </div>
        
        <p>Nếu bạn gặp khó khăn, vui lòng liên hệ với đội ngũ hỗ trợ.</p>
        
        <div class="footer">
            <p>Email này được gửi tự động từ hệ thống <strong>%s</strong></p>
            <p>Vui lòng không trả lời email này.</p>
        </div>
    </div>
</body>
</html>
    `,
		title,        // Page title
		buttonColor,  // Button color
		system,       // Logo/header
		title,        // H2 title
		message,      // Main message
		activationURL, // Button URL
		buttonText,   // Button text
		activationURL, // Fallback URL
		system,       // Footer system name
	)
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