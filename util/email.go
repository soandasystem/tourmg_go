package util

import (
	"fmt"
	"net/smtp"
	"strings"

	"tourmanager/config"
)

// SendPaymentNotification envía un correo de notificación de pago al email indicado.
// Se espera que se llame desde una goroutine para no bloquear el flujo principal.
func SendPaymentNotification(cfg config.Config, toEmail, alumno, estado, monto, commerceOrder string) error {
	if cfg.SMTPHost == "" || cfg.SMTPFrom == "" {
		return fmt.Errorf("configuración SMTP incompleta: SMTP_HOST y SMTP_FROM son requeridos")
	}

	if toEmail == "" {
		return fmt.Errorf("email destinatario vacío, no se envía notificación")
	}

	// Determinar asunto y color según estado
	var subject, colorEstado, mensajeEstado string
	switch strings.ToLower(estado) {
	case "pendiente":
		subject = "Pago pendiente de confirmación"
		colorEstado = "#f59e0b"
		mensajeEstado = "Su pago se encuentra pendiente de confirmación. Le notificaremos cuando se complete."
	case "pagado":
		subject = "Pago confirmado exitosamente"
		colorEstado = "#10b981"
		mensajeEstado = "Su pago ha sido procesado y confirmado exitosamente. ¡Gracias por su pago!"
	case "rechazado":
		subject = "Pago rechazado"
		colorEstado = "#ef4444"
		mensajeEstado = "Lamentablemente su pago ha sido rechazado. Por favor intente nuevamente o contacte a soporte."
	case "anulado":
		subject = "Pago anulado"
		colorEstado = "#6b7280"
		mensajeEstado = "Su pago ha sido anulado. Si no realizó esta acción, por favor contacte a soporte."
	default:
		subject = "Notificación de estado de pago"
		colorEstado = "#6b7280"
		mensajeEstado = "El estado de su pago es: " + estado
	}

	// Construir el cuerpo del email en HTML
	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; margin: 0; padding: 20px;">
    <div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
        <div style="background-color: %s; padding: 20px; text-align: center;">
            <h1 style="color: #ffffff; margin: 0; font-size: 24px;">%s</h1>
        </div>
        <div style="padding: 30px;">
            <p style="font-size: 16px; color: #333;">Estimado/a <strong>%s</strong>,</p>
            <p style="font-size: 14px; color: #555;">%s</p>
            <table style="width: 100%%; border-collapse: collapse; margin: 20px 0;">
                <tr style="border-bottom: 1px solid #eee;">
                    <td style="padding: 10px; font-weight: bold; color: #333;">Estado:</td>
                    <td style="padding: 10px; color: %s; font-weight: bold;">%s</td>
                </tr>
                <tr style="border-bottom: 1px solid #eee;">
                    <td style="padding: 10px; font-weight: bold; color: #333;">Monto:</td>
                    <td style="padding: 10px; color: #333;">$%s</td>
                </tr>
                <tr>
                    <td style="padding: 10px; font-weight: bold; color: #333;">Orden:</td>
                    <td style="padding: 10px; color: #333;">%s</td>
                </tr>
            </table>
            <p style="font-size: 12px; color: #999; margin-top: 30px; text-align: center;">
                Este es un correo automático, por favor no responda a este mensaje.
            </p>
        </div>
    </div>
</body>
</html>`, colorEstado, subject, alumno, mensajeEstado, colorEstado, estado, monto, commerceOrder)

	// Construir headers del mensaje
	headers := make(map[string]string)
	headers["From"] = cfg.SMTPFrom
	headers["To"] = toEmail
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)

	// Configurar autenticación SMTP
	addr := fmt.Sprintf("%s:%s", cfg.SMTPHost, cfg.SMTPPort)

	var auth smtp.Auth
	if cfg.SMTPUser != "" && cfg.SMTPPassword != "" {
		auth = smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPHost)
	}

	// Enviar el email
	err := smtp.SendMail(addr, auth, cfg.SMTPFrom, []string{toEmail}, []byte(msg.String()))
	if err != nil {
		return fmt.Errorf("error enviando email a %s: %w", toEmail, err)
	}

	fmt.Printf("Email de notificación enviado a %s - Estado: %s\n", toEmail, estado)
	return nil
}
