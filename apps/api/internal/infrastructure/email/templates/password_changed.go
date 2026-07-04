package templates

// PasswordChangedEmail is the HTML template for password changed confirmation.
const PasswordChangedEmail = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Password Changed</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #1a1a1a; background-color: #f4f4f5; }
        .container { max-width: 560px; margin: 40px auto; padding: 0 20px; }
        .email-wrapper { background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
        .header { background: linear-gradient(135deg, #28a745 0%, #20c997 100%); padding: 40px 40px 32px; text-align: center; }
        .logo { font-size: 28px; font-weight: 700; color: #ffffff; letter-spacing: -0.5px; }
        .icon { font-size: 48px; margin-bottom: 16px; }
        .content { padding: 40px; }
        h1 { font-size: 24px; font-weight: 600; margin-bottom: 16px; color: #1a1a1a; }
        p { font-size: 16px; color: #4a4a4a; margin-bottom: 24px; }
        .success-box { background: #d4edda; border: 1px solid #c3e6cb; border-radius: 8px; padding: 20px; margin: 24px 0; text-align: center; }
        .success-icon { font-size: 48px; margin-bottom: 12px; }
        .security-tip { background: #e7f3ff; border: 1px solid #b3d7ff; border-radius: 8px; padding: 16px; margin: 24px 0; }
        .security-title { font-weight: 600; color: #0056b3; margin-bottom: 8px; }
        .security-text { font-size: 14px; color: #666; }
        .footer { background: #f9f9f9; padding: 24px 40px; text-align: center; border-top: 1px solid #eeeeee; }
        .footer p { font-size: 13px; color: #888888; margin-bottom: 8px; }
        .footer a { color: #28a745; text-decoration: none; }
    </style>
</head>
<body>
    <div class="container">
        <div class="email-wrapper">
            <div class="header">
                <div class="icon">✓</div>
                <div class="logo">Vyzorix</div>
            </div>
            <div class="content">
                <h1>Password Changed Successfully</h1>
                <p>Hi {{.Name}},</p>
                <p>Your password has been changed successfully.</p>
                <div class="success-box">
                    <div class="success-icon">✓</div>
                    <p style="margin-bottom: 0;">Your account is secure</p>
                </div>
                <div class="security-tip">
                    <div class="security-title">Security Tip</div>
                    <div class="security-text">If you didn't change your password, please contact our support team immediately.</div>
                </div>
            </div>
            <div class="footer">
                <p>Vyzorix Update Server</p>
                <p><a href="{{.BaseURL}}">Visit your dashboard</a></p>
            </div>
        </div>
    </div>
</body>
</html>`
