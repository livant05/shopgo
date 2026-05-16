package mailer

import (
	"fmt"
	"net/smtp"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Pass     string
	From     string
	StoreURL string
}

type Mailer struct{ cfg Config }

func New(cfg Config) *Mailer { return &Mailer{cfg} }

func (m *Mailer) Enabled() bool { return m.cfg.Host != "" }

func (m *Mailer) SendQuoteStatus(to, customerName, storeName, quoteID string, quoteNumber int, total float64, status, note string) error {
	if !m.Enabled() || to == "" {
		return nil
	}
	numStr := fmt.Sprintf("%05d", quoteNumber)
	link := fmt.Sprintf("%s/quote/%s", m.cfg.StoreURL, quoteID)

	var subject, body string
	if status == "accepted" {
		subject = fmt.Sprintf("✅ Cotización N.° %s aprobada — %s", numStr, storeName)
		body = acceptedBody(customerName, numStr, total, link, note, storeName)
	} else {
		subject = fmt.Sprintf("Tu cotización N.° %s — %s", numStr, storeName)
		body = rejectedBody(customerName, numStr, note, storeName)
	}
	return m.send(to, subject, body)
}

func (m *Mailer) send(to, subject, htmlBody string) error {
	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		m.cfg.From, to, subject, htmlBody,
	)
	addr := m.cfg.Host + ":" + m.cfg.Port
	var auth smtp.Auth
	if m.cfg.User != "" {
		auth = smtp.PlainAuth("", m.cfg.User, m.cfg.Pass, m.cfg.Host)
	}
	return smtp.SendMail(addr, auth, m.cfg.From, []string{to}, []byte(msg))
}

func acceptedBody(name, number string, total float64, link, note, storeName string) string {
	noteHTML := ""
	if note != "" {
		noteHTML = fmt.Sprintf(
			`<p style="background:#f0fdf4;border-left:3px solid #86efac;padding:10px 14px;border-radius:0 6px 6px 0;color:#166534">📝 %s</p>`,
			note,
		)
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html><body style="font-family:Arial,sans-serif;max-width:560px;margin:0 auto;padding:32px 16px;background:#f8fafc">
<div style="background:#fff;border-radius:12px;padding:32px;box-shadow:0 2px 8px rgba(0,0,0,.08)">
  <h1 style="color:#16a34a;font-size:1.4rem;margin:0 0 16px">✅ Cotización aprobada</h1>
  <p style="color:#0f172a">Hola <strong>%s</strong>,</p>
  <p style="color:#475569">Tu cotización N.° <strong>%s</strong> por un total de <strong>B/. %.2f</strong> ha sido aprobada.</p>
  %s
  <a href="%s" style="display:inline-block;background:#16a34a;color:#fff;text-decoration:none;border-radius:8px;padding:12px 24px;font-weight:700;margin:16px 0">Ver cotización y proceder al pago →</a>
  <hr style="border:none;border-top:1px solid #e2e8f0;margin:24px 0">
  <p style="color:#94a3b8;font-size:.8rem">%s — responde a este correo si tienes preguntas.</p>
</div>
</body></html>`, name, number, total, noteHTML, link, storeName)
}

func rejectedBody(name, number, note, storeName string) string {
	noteHTML := ""
	if note != "" {
		noteHTML = fmt.Sprintf(
			`<p style="background:#fef2f2;border-left:3px solid #fca5a5;padding:10px 14px;border-radius:0 6px 6px 0;color:#7f1d1d">Motivo: %s</p>`,
			note,
		)
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html><body style="font-family:Arial,sans-serif;max-width:560px;margin:0 auto;padding:32px 16px;background:#f8fafc">
<div style="background:#fff;border-radius:12px;padding:32px;box-shadow:0 2px 8px rgba(0,0,0,.08)">
  <h1 style="color:#dc2626;font-size:1.4rem;margin:0 0 16px">Cotización no aprobada</h1>
  <p style="color:#0f172a">Hola <strong>%s</strong>,</p>
  <p style="color:#475569">Tu cotización N.° <strong>%s</strong> no pudo ser aprobada en esta oportunidad.</p>
  %s
  <p style="color:#475569">Si tienes preguntas o deseas ajustar la cotización, contáctanos.</p>
  <hr style="border:none;border-top:1px solid #e2e8f0;margin:24px 0">
  <p style="color:#94a3b8;font-size:.8rem">%s — responde a este correo si tienes preguntas.</p>
</div>
</body></html>`, name, number, noteHTML, storeName)
}
