// Package templates provides email template definitions.
package templates

// VerificationEmail is the HTML template for email verification.
const VerificationEmail = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verify your email</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #1a1a1a; background-color: #f4f4f5; }
        .container { max-width: 560px; margin: 40px auto; padding: 0 20px; }
        .email-wrapper { background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 40px 40px 32px; text-align: center; }
        .logo { font-size: 28px; font-weight: 700; color: #ffffff; letter-spacing: -0.5px; }
        .content { padding: 40px; }
        h1 { font-size: 24px; font-weight: 600; margin-bottom: 16px; color: #1a1a1a; }
        p { font-size: 16px; color: #4a4a4a; margin-bottom: 24px; }
        .button-wrapper { text-align: center; margin: 32px 0; }
        .button { display: inline-block; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: #ffffff; text-decoration: none; font-size: 16px; font-weight: 600; padding: 16px 32px; border-radius: 8px; transition: transform 0.2s; }
        .button:hover { transform: translateY(-2px); }
        .expiry { text-align: center; font-size: 14px; color: #888888; margin-top: 24px; }
        .footer { background: #f9f9f9; padding: 24px 40px; text-align: center; border-top: 1px solid #eeeeee; }
        .footer p { font-size: 13px; color: #888888; margin-bottom: 8px; }
        .footer a { color: #667eea; text-decoration: none; }
        .ignore { background: #f0f0f0; padding: 20px; border-radius: 8px; margin-top: 24px; font-size: 14px; color: #666666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="email-wrapper">
            <div class="header">
                <div class="logo">Vyzorix</div>
            </div>
            <div class="content">
                <h1>Verify your email address</h1>
                <p>Hi {{.Name}},</p>
                <p>Thanks for signing up! Please verify your email address by clicking the button below. This helps us keep your account secure.</p>
                <div class="button-wrapper">
                    <a href="{{.VerifyURL}}" class="button">Verify Email Address</a>
                </div>
                <p class="expiry">This link expires in {{.ExpiryHours}} hours</p>
                <div class="ignore">
                    <p>If you didn't create an account with Vyzorix, you can safely ignore this email. Someone may have entered your email address by mistake.</p>
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
