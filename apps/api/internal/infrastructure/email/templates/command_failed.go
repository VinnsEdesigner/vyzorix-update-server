package templates

// CommandFailedEmail is the HTML template for command failed notifications.
const CommandFailedEmail = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Command Failed</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #1a1a1a; background-color: #f4f4f5; }
        .container { max-width: 560px; margin: 40px auto; padding: 0 20px; }
        .email-wrapper { background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
        .header { background: linear-gradient(135deg, #dc3545 0%, #bd2139 100%); padding: 40px 40px 32px; text-align: center; }
        .logo { font-size: 28px; font-weight: 700; color: #ffffff; letter-spacing: -0.5px; }
        .icon { font-size: 48px; margin-bottom: 16px; }
        .content { padding: 40px; }
        h1 { font-size: 24px; font-weight: 600; margin-bottom: 16px; color: #1a1a1a; }
        p { font-size: 16px; color: #4a4a4a; margin-bottom: 16px; }
        .error-box { background: #f8d7da; border: 1px solid #f5c6cb; border-radius: 8px; padding: 20px; margin: 24px 0; }
        .error-row { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid rgba(0,0,0,0.1); }
        .error-row:last-child { border-bottom: none; }
        .error-label { font-weight: 600; color: #333; }
        .error-value { color: #666; }
        .failure-reason { background: #721c24; color: #f8d7da; padding: 12px; border-radius: 4px; margin-top: 12px; font-family: monospace; }
        .button-wrapper { text-align: center; margin: 32px 0; }
        .button { display: inline-block; background: linear-gradient(135deg, #dc3545 0%, #bd2139 100%); color: #ffffff; text-decoration: none; font-size: 16px; font-weight: 600; padding: 16px 32px; border-radius: 8px; }
        .footer { background: #f9f9f9; padding: 24px 40px; text-align: center; border-top: 1px solid #eeeeee; }
        .footer p { font-size: 13px; color: #888888; margin-bottom: 8px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="email-wrapper">
            <div class="header">
                <div class="icon">❌</div>
                <div class="logo">Command Failed</div>
            </div>
            <div class="content">
                <h1>Command Execution Failed</h1>
                <p>Hi {{.OperatorName}},</p>
                <p>A command you sent to your device has failed to execute.</p>
                <div class="error-box">
                    <div class="error-row">
                        <span class="error-label">Device ID:</span>
                        <span class="error-value">{{.DeviceID}}</span>
                    </div>
                    <div class="error-row">
                        <span class="error-label">Command:</span>
                        <span class="error-value">{{.CommandName}}</span>
                    </div>
                    <div class="error-row">
                        <span class="error-label">Time:</span>
                        <span class="error-value">{{.Timestamp}}</span>
                    </div>
                    <div class="failure-reason">Reason: {{.FailureReason}}</div>
                </div>
                <div class="button-wrapper">
                    <a href="{{.BaseURL}}/devices/{{.DeviceID}}/commands" class="button">View Commands</a>
                </div>
                <p>Please check your device status and retry the command if necessary.</p>
            </div>
            <div class="footer">
                <p>Vyzorix Update Server</p>
                <p>This is an automated notification. Do not reply.</p>
            </div>
        </div>
    </div>
</body>
</html>`
