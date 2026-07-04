package templates

// DeviceOnlineEmail is the HTML template for device online notifications.
const DeviceOnlineEmail = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Device Online</title>
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
        p { font-size: 16px; color: #4a4a4a; margin-bottom: 16px; }
        .info-box { background: #f8f9fa; border-radius: 8px; padding: 20px; margin: 24px 0; }
        .info-row { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #eee; }
        .info-row:last-child { border-bottom: none; }
        .info-label { font-weight: 600; color: #333; }
        .info-value { color: #666; }
        .info-value.online { color: #28a745; font-weight: 600; }
        .success-text { font-size: 14px; color: #28a745; margin-top: 16px; padding: 12px; background: #d4edda; border-radius: 8px; }
        .footer { background: #f9f9f9; padding: 24px 40px; text-align: center; border-top: 1px solid #eeeeee; }
        .footer p { font-size: 13px; color: #888888; margin-bottom: 8px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="email-wrapper">
            <div class="header">
                <div class="icon">🟢</div>
                <div class="logo">Device Online</div>
            </div>
            <div class="content">
                <h1>Device Reconnected</h1>
                <p>Hi {{.OperatorName}},</p>
                <p>Your device has successfully reconnected to the server.</p>
                <div class="info-box">
                    <div class="info-row">
                        <span class="info-label">Device ID:</span>
                        <span class="info-value">{{.DeviceID}}</span>
                    </div>
                    <div class="info-row">
                        <span class="info-label">Device Name:</span>
                        <span class="info-value">{{.DeviceName}}</span>
                    </div>
                    <div class="info-row">
                        <span class="info-label">Status:</span>
                        <span class="info-value online">ONLINE</span>
                    </div>
                    <div class="info-row">
                        <span class="info-label">Connected At:</span>
                        <span class="info-value">{{.Timestamp}}</span>
                    </div>
                </div>
                <p class="success-text">All systems are operational. You can now send commands to your device.</p>
            </div>
            <div class="footer">
                <p>Vyzorix Update Server</p>
                <p>This is an automated notification. Do not reply.</p>
            </div>
        </div>
    </div>
</body>
</html>`
