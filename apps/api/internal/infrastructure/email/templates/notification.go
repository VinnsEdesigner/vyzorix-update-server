package templates

// NotificationContactEmail is the generic contact point email template for
// alert and dispatcher-driven notifications. Rendered when the alert channel
// can't resolve a per-event template.
const NotificationContactEmail = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Subject}}</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #1a1a1a; background-color: #f4f4f5; }
        .container { max-width: 560px; margin: 40px auto; padding: 0 20px; }
        .email-wrapper { background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
        .header { background: linear-gradient(135deg, #0d2b45 0%, #1e4a73 100%); padding: 40px 40px 32px; text-align: center; }
        .logo { font-size: 28px; font-weight: 700; color: #ffffff; letter-spacing: -0.5px; }
        .content { padding: 40px; }
        h1 { font-size: 24px; font-weight: 600; margin-bottom: 16px; color: #1a1a1a; }
        p { font-size: 16px; color: #4a4a4a; margin-bottom: 16px; }
        .fields { background: #f8f9fa; border: 1px solid #e2e8f0; border-radius: 8px; padding: 20px; margin: 24px 0; }
        .field-row { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid rgba(0,0,0,0.08); }
        .field-row:last-child { border-bottom: none; }
        .field-label { font-weight: 600; color: #333; }
        .field-value { color: #666; font-family: monospace; }
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
                <h1>{{.Subject}}</h1>
                <p>{{.Body}}</p>
                {{if .Data}}
                <div class="fields">
                    {{range $key, $value := .Data}}
                    <div class="field-row">
                        <span class="field-label">{{$key}}</span>
                        <span class="field-value">{{$value}}</span>
                    </div>
                    {{end}}
                </div>
                {{end}}
            </div>
            <div class="footer">
                <p class="footer-text">This is a VyzoriX notification contact point message.</p>
            </div>
        </div>
    </div>
</body>
</html>`
