package templates

// InvitationEmail is the HTML template for organization invitations.
const InvitationEmail = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>You've been invited to join {{.OrganizationName}}</title>
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
        .role-badge { display: inline-block; background: #e8eaff; color: #667eea; padding: 6px 16px; border-radius: 20px; font-size: 14px; font-weight: 600; margin-bottom: 24px; }
        .button-wrapper { text-align: center; margin: 32px 0; }
        .button { display: inline-block; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: #ffffff; text-decoration: none; font-size: 16px; font-weight: 600; padding: 16px 32px; border-radius: 8px; transition: transform 0.2s; }
        .button:hover { transform: translateY(-2px); }
        .expiry { text-align: center; font-size: 14px; color: #888888; margin-top: 24px; }
        .notes { background: #f9f9f9; border-left: 4px solid #667eea; padding: 16px; margin: 24px 0; font-style: italic; }
        .footer { background: #f9f9f9; padding: 24px 40px; text-align: center; border-top: 1px solid #eeeeee; }
        .footer p { font-size: 13px; color: #888888; margin-bottom: 8px; }
        .footer a { color: #667eea; text-decoration: none; }
    </style>
</head>
<body>
    <div class="container">
        <div class="email-wrapper">
            <div class="header">
                <div class="logo">Vyzorix</div>
            </div>
            <div class="content">
                <h1>You've been invited!</h1>
                <p>Hi {{.InviteeName}},</p>
                <p><strong>{{.InviterName}}</strong> has invited you to join <strong>{{.OrganizationName}}</strong> as a {{.Role}}.</p>
                <div class="role-badge">{{.Role}}</div>
                {{if .InviterNotes}}
                <div class="notes">
                    <strong>Message from {{.InviterName}}:</strong><br>
                    {{.InviterNotes}}
                </div>
                {{end}}
                <div class="button-wrapper">
                    <a href="{{.AcceptURL}}" class="button">Accept Invitation</a>
                </div>
                <p>If you didn't expect this invitation, you can safely ignore this email.</p>
                <p class="expiry">This invitation expires in {{.ExpiryDays}} days</p>
            </div>
            <div class="footer">
                <p>Vyzorix Update Server</p>
                <p><a href="{{.BaseURL}}">Visit your dashboard</a></p>
            </div>
        </div>
    </div>
</body>
</html>`
