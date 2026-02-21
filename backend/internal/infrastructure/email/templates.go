package email

import "fmt"

func VerificationEmail(appURL, token string) (subject string, body string) {
	subject = "DNA Fami — Verifique seu email"
	link := fmt.Sprintf("%s/verify-email?token=%s", appURL, token)
	body = fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body style="font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
  <h2 style="color: #2563EB;">DNA Fami</h2>
  <p>Obrigado por se cadastrar! Clique no botão abaixo para verificar seu email:</p>
  <a href="%s" style="display: inline-block; background: #2563EB; color: white; padding: 12px 24px; border-radius: 8px; text-decoration: none; font-weight: bold;">
    Verificar Email
  </a>
  <p style="color: #6B7280; font-size: 14px; margin-top: 20px;">
    Ou copie e cole este link no navegador:<br>
    <a href="%s">%s</a>
  </p>
  <p style="color: #9CA3AF; font-size: 12px;">Este link expira em 24 horas.</p>
</body>
</html>`, link, link, link)
	return
}

func InviteEmail(appURL, token, tenantName, inviterName string) (subject string, body string) {
	subject = fmt.Sprintf("DNA Fami — Convite para %s", tenantName)
	link := fmt.Sprintf("%s/accept-invite?token=%s", appURL, token)
	body = fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body style="font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
  <h2 style="color: #2563EB;">DNA Fami</h2>
  <p>Você foi convidado por <strong>%s</strong> para participar do dashboard <strong>%s</strong>.</p>
  <a href="%s" style="display: inline-block; background: #2563EB; color: white; padding: 12px 24px; border-radius: 8px; text-decoration: none; font-weight: bold;">
    Aceitar Convite
  </a>
  <p style="color: #6B7280; font-size: 14px; margin-top: 20px;">
    Ou copie e cole este link no navegador:<br>
    <a href="%s">%s</a>
  </p>
  <p style="color: #9CA3AF; font-size: 12px;">Este convite expira em 7 dias.</p>
</body>
</html>`, inviterName, tenantName, link, link, link)
	return
}
