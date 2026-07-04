// Package templates provides all email template constants.
package templates

// NewLoginEmail is the HTML template for new login notification emails.
// MEDIUM-10: Added for login notification feature.
const NewLoginEmail = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>New Login to Your Account</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #4F46E5; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f9fafb; padding: 20px; border-radius: 0 0 8px 8px; }
        .alert { background: #FEF3C7; border-left: 4px solid #F59E0B; padding: 15px; margin: 15px 0; }
        .details { background: white; padding: 15px; border-radius: 8px; margin: 15px 0; }
        .details table { width: 100%; border-collapse: collapse; }
        .details td { padding: 8px 0; border-bottom: 1px solid #e5e7eb; }
        .details td:first-child { font-weight: bold; width: 120px; }
        .button { display: inline-block; background: #4F46E5; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin: 15px 0; }
        .footer { text-align: center; margin-top: 20px; font-size: 12px; color: #6b7280; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🔐 New Login Detected</h1>
    </div>
    <div class="content">
        <p>Hello {{.OperatorName}},</p>
        
        <div class="alert">
            <strong>⚠️ Security Notice</strong><br>
            A new login was detected on your account. If this wasn't you, please secure your account immediately.
        </div>
        
        <div class="details">
            <strong>Login Details:</strong>
            <table>
                <tr><td>Time:</td><td>{{.Timestamp}}</td></tr>
                <tr><td>IP Address:</td><td>{{.IPAddress}}</td></tr>
                <tr><td>Location:</td><td>{{.Location}}</td></tr>
                <tr><td>Device:</td><td>{{.Device}}</td></tr>
            </table>
        </div>
        
        <p>If you don't recognize this login:</p>
        <ul>
            <li>Click the button below to review your account security settings</li>
            <li>Consider changing your password</li>
            <li>Enable two-factor authentication if not already enabled</li>
        </ul>
        
        <p style="text-align: center;">
            <a href="{{.BaseURL}}/auth/security" class="button">Review Security Settings</a>
        </p>
        
        <div class="footer">
            <p>This is an automated security notification from Vyzorix.</p>
            <p>You received this email because a login to your account was detected.</p>
        </div>
    </div>
</body>
</html>
`
