package templates

// InvitationAcceptedEmail is the HTML template for invitation acceptance notifications.
const InvitationAcceptedEmail = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Invitation accepted - {{.OrganizationName}}</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #1a1a1a; background-color: #f4f4f5; }
        .container { max-width: 560px; margin: 40px auto; padding: 0 20px; }
        .email-wrapper { background: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.08); }
        .header { background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%); padding: 40px 40px 32px; text-align: center; }
        .logo { font-size: 28px; font-weight: 700; color: #ffffff; letter-spacing: -0.5px; }
        .content { padding: 40px; }
        h1 { font-size: 24px; font-weight: 600; margin-bottom: 16px; color: #1a1a1a; }
        p { font-size: 16px; color: #4a4a4a; margin-bottom: 24px; }
        .success-icon { text-align: center; margin: 24px 0; }
        .success-icon span { display: inline-block; background: #38ef7d; color: #ffffff; width: 64px; height: 64px; border-radius: 50%; line-height: 64px; font-size: 32px; }
        .details { background: #f9f9f9; border-radius: 8px; padding: 20px; margin: 24px 0; }
        .details p { margin-bottom: 8px; }
        .details strong { color: #1a1a1a; }
        .footer { background: #f9f9f9; padding: 24px 40px; text-align: center; border-top: 1px solid #eeeeee; }
        .footer p { font-size: 13px; color: #888888; margin-bottom: 8px; }
        .footer a { color: #11998e; text-decoration: none; }
    </style>
</head>
<body>
    <div class="container">
        <div class="email-wrapper">
            <div class="header">
                <div class="logo">Vyzorix</div>
            </div>
            <div class="content">
                <h1>Invitation Accepted!</h1>
                <p>Great news! <strong>{{.InviteeName}}</strong> has accepted your invitation to join <strong>{{.OrganizationName}}</strong>.</p>
                <div class="success-icon">
                    <span>✓</span>
                </div>
                <div class="details">
                    <p><strong>Role:</strong> {{.Role}}</p>
                    <p><strong>Accepted:</strong> {{.AcceptedAt}}</p>
                    {{if .InviteeNotes}}
                    <p><strong>Message:</strong> {{.InviteeNotes}}</p>
                    {{end}}
                </div>
                <p>You can now see them in your organization members list.</p>
            </div>
            <div class="footer">
                <p>Vyzorix Update Server</p>
                <p><a href="{{.BaseURL}}">View Organization</a></p>
            </div>
        </div>
    </div>
</body>
</html>`
