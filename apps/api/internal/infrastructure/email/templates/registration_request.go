package templates

// RegistrationRequestEmail is the HTML template for registration request notifications.
const RegistrationRequestEmail = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Registration Request</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #1a1a1a; background-color: #f4f4f5; }
        .container { max-width: 560px; margin: 40px auto; padding: 0 20px; }
        .email-wrapper { background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
        .header { background: linear-gradient(135deg, #6f42c1 0%, #5a3d8a 100%); padding: 40px 40px 32px; text-align: center; }
        .logo { font-size: 28px; font-weight: 700; color: #ffffff; letter-spacing: -0.5px; }
        .icon { font-size: 48px; margin-bottom: 16px; }
        .content { padding: 40px; }
        h1 { font-size: 24px; font-weight: 600; margin-bottom: 16px; color: #1a1a1a; }
        p { font-size: 16px; color: #4a4a4a; margin-bottom: 16px; }
        .request-box { background: #f8f0ff; border: 1px solid #d4b8e8; border-radius: 8px; padding: 20px; margin: 24px 0; }
        .request-row { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #d4b8e8; }
        .request-row:last-child { border-bottom: none; }
        .request-label { font-weight: 600; color: #333; }
        .request-value { color: #6f42c1; }
        .button-wrapper { text-align: center; margin: 32px 0; }
        .button { display: inline-block; background: linear-gradient(135deg, #28a745 0%, #20c997 100%); color: #ffffff; text-decoration: none; font-size: 16px; font-weight: 600; padding: 16px 32px; border-radius: 8px; margin-right: 12px; }
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
                <div class="logo">Registration Request</div>
            </div>
            <div class="content">
                <h1>New Device Registration</h1>
                <p>Hi {{.OperatorName}},</p>
                <p>A new device is requesting to be registered to your account.</p>
                <div class="request-box">
                    <div class="request-row">
                        <span class="request-label">Requester Name:</span>
                        <span class="request-value">{{.RequesterName}}</span>
                    </div>
                    <div class="request-row">
                        <span class="request-label">Device ID:</span>
                        <span class="request-value">{{.DeviceID}}</span>
                    </div>
                    <div class="request-row">
                        <span class="request-label">Device Name:</span>
                        <span class="request-value">{{.DeviceName}}</span>
                    </div>
                    <div class="request-row">
                        <span class="request-label">Requested At:</span>
                        <span class="request-value">{{.Timestamp}}</span>
                    </div>
                </div>
                <div class="button-wrapper">
                    <a href="{{.BaseURL}}/registrations/pending" class="button">Review Request</a>
                    <a href="{{.BaseURL}}/registrations" class="button button-secondary">View All</a>
                </div>
            </div>
            <div class="footer">
                <p>Vyzorix Update Server</p>
                <p>This is an automated notification. Do not reply.</p>
            </div>
        </div>
    </div>
</body>
</html>`
