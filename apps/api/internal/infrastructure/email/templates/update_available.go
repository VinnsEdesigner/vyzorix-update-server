package templates

// UpdateAvailableEmail is the HTML template for update available notifications.
const UpdateAvailableEmail = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Update Available</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #1a1a1a; background-color: #f4f4f5; }
        .container { max-width: 560px; margin: 40px auto; padding: 0 20px; }
        .email-wrapper { background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
        .header { background: linear-gradient(135deg, #007bff 0%, #0056b3 100%); padding: 40px 40px 32px; text-align: center; }
        .logo { font-size: 28px; font-weight: 700; color: #ffffff; letter-spacing: -0.5px; }
        .icon { font-size: 48px; margin-bottom: 16px; }
        .content { padding: 40px; }
        h1 { font-size: 24px; font-weight: 600; margin-bottom: 16px; color: #1a1a1a; }
        p { font-size: 16px; color: #4a4a4a; margin-bottom: 16px; }
        .update-box { background: #e7f3ff; border: 1px solid #b3d7ff; border-radius: 8px; padding: 20px; margin: 24px 0; }
        .update-row { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #b3d7ff; }
        .update-row:last-child { border-bottom: none; }
        .update-label { font-weight: 600; color: #333; }
        .update-value { color: #0066cc; font-weight: 500; }
        .button-wrapper { text-align: center; margin: 32px 0; }
        .button { display: inline-block; background: linear-gradient(135deg, #007bff 0%, #0056b3 100%); color: #ffffff; text-decoration: none; font-size: 16px; font-weight: 600; padding: 16px 32px; border-radius: 8px; margin-right: 12px; }
        .button-secondary { background: linear-gradient(135deg, #6c757d 0%, #495057 100%); }
        .footer { background: #f9f9f9; padding: 24px 40px; text-align: center; border-top: 1px solid #eeeeee; }
        .footer p { font-size: 13px; color: #888888; margin-bottom: 8px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="email-wrapper">
            <div class="header">
                <div class="icon"></div>
                <div class="logo">Update Available</div>
            </div>
            <div class="content">
                <h1>New Update Ready</h1>
                <p>Hi {{.OperatorName}},</p>
                <p>A software update is available for your device.</p>
                <div class="update-box">
                    <div class="update-row">
                        <span class="update-label">Device ID:</span>
                        <span class="update-value">{{.DeviceID}}</span>
                    </div>
                    <div class="update-row">
                        <span class="update-label">New Version:</span>
                        <span class="update-value">{{.UpdateVersion}}</span>
                    </div>
                </div>
                <div class="button-wrapper">
                    <a href="{{.BaseURL}}/devices/{{.DeviceID}}/updates" class="button">View Update</a>
                    <a href="{{.BaseURL}}/updates" class="button button-secondary">All Updates</a>
                </div>
                <p>You can deploy this update from your dashboard at any time.</p>
            </div>
            <div class="footer">
                <p>Vyzorix Update Server</p>
                <p>This is an automated notification. Do not reply.</p>
            </div>
        </div>
    </div>
</body>
</html>`
