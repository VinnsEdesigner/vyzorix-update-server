package templates

// ThresholdBreachEmail is the HTML template for threshold breach alerts.
const ThresholdBreachEmail = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Threshold Breach Alert</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #1a1a1a; background-color: #f4f4f5; }
        .container { max-width: 560px; margin: 40px auto; padding: 0 20px; }
        .email-wrapper { background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
        .header { background: linear-gradient(135deg, #ff6b6b 0%, #ee5a24 100%); padding: 40px 40px 32px; text-align: center; }
        .logo { font-size: 28px; font-weight: 700; color: #ffffff; letter-spacing: -0.5px; }
        .icon { font-size: 48px; margin-bottom: 16px; }
        .content { padding: 40px; }
        h1 { font-size: 24px; font-weight: 600; margin-bottom: 16px; color: #1a1a1a; }
        p { font-size: 16px; color: #4a4a4a; margin-bottom: 16px; }
        .alert-box { background: #fff3cd; border: 1px solid #ffc107; border-radius: 8px; padding: 20px; margin: 24px 0; }
        .alert-row { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid rgba(0,0,0,0.1); }
        .alert-row:last-child { border-bottom: none; }
        .alert-label { font-weight: 600; color: #333; }
        .alert-value { color: #666; font-family: monospace; }
        .alert-value.critical { color: #dc3545; font-weight: bold; }
        .button-wrapper { text-align: center; margin: 32px 0; }
        .button { display: inline-block; background: linear-gradient(135deg, #ff6b6b 0%, #ee5a24 100%); color: #ffffff; text-decoration: none; font-size: 16px; font-weight: 600; padding: 16px 32px; border-radius: 8px; }
        .footer { background: #f9f9f9; padding: 24px 40px; text-align: center; border-top: 1px solid #eeeeee; }
        .footer p { font-size: 13px; color: #888888; margin-bottom: 8px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="email-wrapper">
            <div class="header">
                <div class="icon">⚠️</div>
                <div class="logo">Threshold Alert</div>
            </div>
            <div class="content">
                <h1>Device Threshold Breached</h1>
                <p>Hi {{.OperatorName}},</p>
                <p>A device threshold has been breached. Please review the alert details below.</p>
                <div class="alert-box">
                    <div class="alert-row">
                        <span class="alert-label">Device ID:</span>
                        <span class="alert-value">{{.DeviceID}}</span>
                    </div>
                    <div class="alert-row">
                        <span class="alert-label">Device Name:</span>
                        <span class="alert-value">{{.DeviceName}}</span>
                    </div>
                    <div class="alert-row">
                        <span class="alert-label">Alert Type:</span>
                        <span class="alert-value critical">{{.AlertType}}</span>
                    </div>
                    <div class="alert-row">
                        <span class="alert-label">Current Value:</span>
                        <span class="alert-value critical">{{.CurrentValue}}</span>
                    </div>
                    <div class="alert-row">
                        <span class="alert-label">Threshold:</span>
                        <span class="alert-value">{{.Threshold}}</span>
                    </div>
                    <div class="alert-row">
                        <span class="alert-label">Time:</span>
                        <span class="alert-value">{{.Timestamp}}</span>
                    </div>
                </div>
                <div class="button-wrapper">
                    <a href="{{.BaseURL}}/devices/{{.DeviceID}}" class="button">View Device</a>
                </div>
            </div>
            <div class="footer">
                <p>Vyzorix Update Server</p>
                <p>This is an automated alert. Do not reply.</p>
            </div>
        </div>
    </div>
</body>
</html>`
