package templates

// LeakAlertEmail is sent when a leaked service account token is detected.
const LeakAlertEmail = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Service Account Token Leak Detected</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #1a1a1a; background-color: #f4f4f5; }
        .container { max-width: 560px; margin: 40px auto; padding: 0 20px; }
        .email-wrapper { background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
        .header { background: linear-gradient(135deg, #c92a2a 0%, #a61e1e 100%); padding: 40px 40px 32px; text-align: center; }
        .logo { font-size: 28px; font-weight: 700; color: #ffffff; letter-spacing: -0.5px; }
        .content { padding: 40px; }
        h1 { font-size: 24px; font-weight: 600; margin-bottom: 16px; color: #c92a2a; }
        p { font-size: 16px; color: #4a4a4a; margin-bottom: 16px; }
        .alert-box { background: #ffe3e3; border: 1px solid #ffa8a8; border-radius: 8px; padding: 20px; margin: 24px 0; }
        .alert-row { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid rgba(0,0,0,0.1); }
        .alert-row:last-child { border-bottom: none; }
        .alert-label { font-weight: 600; color: #333; }
        .alert-value { color: #666; font-family: monospace; }
        .action { margin: 24px 0; }
        .action a { background: #c92a2a; color: #ffffff; padding: 12px 24px; text-decoration: none; border-radius: 6px; display: inline-block; }
        .footer { padding: 24px 40px; text-align: center; border-top: 1px solid #e2e8f0; }
        .footer-text { font-size: 14px; color: #9ca3af; }
    </style>
</head>
<body>
    <div class="container">
        <div class="email-wrapper">
            <div class="header">
                <div class="logo">{{if .BaseURL}}{{.BaseURL}}{{else}}VyzoriX{{end}}</div>
            </div>
            <div class="content">
                <h1>Service Account Token Leak Detected</h1>
                <p>
                    A service account token belonging to this organization has been
                    detected in an outbound payload (likely a Git push, export, or
                    external share). This token grants scoped access to the service
                    account it belongs to and should be treated as compromised.
                </p>
                <div class="alert-box">
                    <div class="alert-row">
                        <span class="alert-label">Service Account</span>
                        <span class="alert-value">{{.ServiceAccountName}}</span>
                    </div>
                    <div class="alert-row">
                        <span class="alert-label">Token Prefix</span>
                        <span class="alert-value">{{.TokenPrefix}}…</span>
                    </div>
                    <div class="alert-row">
                        <span class="alert-label">Detected At</span>
                        <span class="alert-value">{{.DetectedAt}}</span>
                    </div>
                    <div class="alert-row">
                        <span class="alert-label">Leaks Found</span>
                        <span class="alert-value">{{.LeakCount}}</span>
                    </div>
                </div>
                <p>
                    Rotate the affected token immediately. The old key is already
                    unusable once rotated; the new key is issued to you directly.
                </p>
                <div class="action">
                    <a href="{{.BaseURL}}/service-accounts">Review Service Accounts</a>
                </div>
            </div>
            <div class="footer">
                <p class="footer-text">This is an automated security alert from VyzoriX.</p>
            </div>
        </div>
    </div>
</body>
</html>`
