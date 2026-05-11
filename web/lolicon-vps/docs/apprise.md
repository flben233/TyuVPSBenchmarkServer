## Email

Syntax

Valid syntax is as follows:

    mailto://{user}:{password}@{domain}
    mailto://{user}:{password}@{domain}:{port}
    mailto://{domain}?user={user}&pass={password}
    mailto://{user}:{password}@{domain}/{to_email}
    mailto://{user}:{password}@{domain}/{to_email1}/{to_email2}/{to_emailN}

Adding an s to the schema (i.e. mailtos://) switches to a secure STARTTLS connection (port 587 by default):

    mailtos://{user}:{password}@{domain}
    mailtos://{user}:{password}@{domain}:{port}
    mailtos://{domain}?user={user}&pass={password}
    mailtos://{user}:{password}@{domain}/{to_email}
    mailtos://{user}:{password}@{domain}/{to_email1}/{to_email2}/{to_emailN}

Built-In Provider Support

Apprise automatically detects many email providers based on the From address derived from your URL.
When a provider is recognized, Apprise automatically configures:

    SMTP host
    Port
    Secure mode (SSL or STARTTLS)
    Login format (full email vs user id)

In most cases, you only need to provide your email and password.
Provider	Example URL	Notes
Google (Gmail)	mailto://user:app-password@gmail.com	If 2-Step Verification is enabled, generate an App Password: https://security.google.com/settings/security/apppasswords
Yahoo	mailto://user:app-password@yahoo.com	Requires an App Password: https://help.yahoo.com/kb/SLN15241.html
Fastmail	mailto://user:app-password@fastmail.com	App Password must permit SMTP. See supported domains here.
GMX	mailto://user:password@gmx.net	Also supports gmx.com, gmx.de, gmx.at, gmx.ch, gmx.fr.
Zoho	mailto://user:password@zoho.com	Provider defaults are applied automatically.
Yandex	mailto://user:password@yandex.com	Login may be user-id based depending on domain rules.
SendGrid (SMTP)	mailto://apikey:password@sendgrid.com?from=noreply@yourdomain.com	from= must use a validated sender identity.
QQ / Foxmail	mailto://user:password@qq.com	Provider defaults are applied automatically.
163.com	mailto://user:password@163.com	Provider defaults are applied automatically.
Microsoft (Outlook, Hotmail, Office 365)	Use azure:// instead	Microsoft disabled SMTP basic authentication. Use the azure:// plugin.

    This is not an exhaustive list. Additional domains are automatically detected when supported.

Automatic Secure Upgrade

When a supported provider is detected, Apprise automatically enforces secure connections using the correct TLS mode and port.

Even if you use mailto://, secure mode is applied when the provider template defines it.

If you explicitly specify smtp=, Apprise assumes you are overriding provider detection.
Email Address Formatting

Email addresses may be written as:

    user@example.com
    Optional Name<user@example.com>

This syntax works in:

    URL targets
    from=
    cc=
    bcc=
    reply=

If you need spaces inside a URL, encode them as %20.

Example:

from=Optional%20Name<noreply@example.com>

Recipient Behaviour
What you specify	What happens
No targets and no to=	Apprise sends the email to the sender address (the derived From email).
Targets in the URL path	Each target becomes a recipient.
to= in the query string	Treated as an additional recipient (same as adding a target).
cc= / bcc=	Applied to each generated email.
reply=	Sets the Reply-To header (can be multiple).
Using Custom SMTP Servers

If your provider is not automatically detected, configure SMTP manually.

Defaults:

    mailto://: defaults to port 25
    mailtos://: defaults to port 587 using STARTTLS

Most public providers require TLS. Prefer mailtos:// for external servers.
Authenticated SMTP Examples

Send using a custom SMTP host:

    mailtos://user:password@server.com?smtp=smtp.server.com&from=noreply@server.com

Include a From display name:

    mailtos://user:password@server.com?smtp=smtp.server.com&from=Optional%20Name<noreply@server.com>

Force SSL (usually port 465):

    mailtos://user:password@server.com:465?smtp=smtp.server.com&mode=ssl&from=noreply@server.com

## DingTalk

Syntax

Valid syntax is as follows:

    dingtalk://{ApiKey}/{ToPhoneNo}
    dingtalk://{ApiKey}/{ToPhoneNo1}/{ToPhoneNo2}/{ToPhoneNoN}
    dingtalk://{Secret}@{ApiKey}/{ToPhoneNo}
    dingtalk://{Secret}@{ApiKey}/{ToPhoneNo1}/{ToPhoneNo2}/{ToPhoneNoN}

Parameter Breakdown
Variable	Required	Description
ApiKey	Yes	The API Key associated with your DingTalk account. This is available to you via the DingTalk Dashboard.
ToPhoneNo	No	A phone number to send your notification to
Secret	No	The optional secret key to associate with the message signing

## QQ Push

Syntax

Valid syntax is as follows:

    https://qmsg.zendee.cn/send/{token}
    qq://{token}
    qq://?token={token}

Parameter Breakdown
Variable	Required	Description
token	Yes	Your personal QQ Push token from qmsg.zendee.cn