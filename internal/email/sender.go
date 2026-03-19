package email

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strings"
	"time"
)

// Sender sends notification emails through an SMTP server.
type Sender struct {
	To       string
	From     string
	SMTPAddr string
	Password string

	// DialTimeout limits TCP connect time to SMTP server.
	DialTimeout time.Duration
	// CommandTimeout limits total SMTP command time after connect.
	CommandTimeout time.Duration
}

// New creates a sender with raw SMTP settings.
func New(to, from, smtpAddr, password string) *Sender {
	return &Sender{
		To:       to,
		From:     from,
		SMTPAddr: smtpAddr,
		Password: password,
		DialTimeout:    8 * time.Second,
		CommandTimeout: 12 * time.Second,
	}
}

// SendMetricAlert sends one alert email for a metric (cpu, memory, etc).
func (s *Sender) SendMetricAlert(metricName string, usage, threshold float64, sustainedFor, cooldown time.Duration) error {
	host, _, err := net.SplitHostPort(s.SMTPAddr)
	if err != nil {
		return fmt.Errorf("invalid smtp address %q: %w", s.SMTPAddr, err)
	}

	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "unknown-host"
	}

	now := time.Now()
	localTime := now.Format("2006-01-02 15:04:05 MST")
	utcTime := now.UTC().Format("2006-01-02 15:04:05 UTC")
	prettyMetric := strings.ToUpper(metricName)

	auth := smtp.PlainAuth("", s.From, s.Password, host)
	subject := fmt.Sprintf("%s Alert: %.2f%% on %s", prettyMetric, usage, hostname)
	body := fmt.Sprintf(
		"Hello,\n\n"+
			"cpu-alert detected sustained high %s usage and triggered this notification.\n\n"+
			"Alert details:\n"+
			"- Host: %s\n"+
			"- Alert time: %s (%s)\n"+
			"- Metric: %s\n"+
			"- Current usage: %.2f%%\n"+
			"- Configured threshold: %.2f%%\n"+
			"- Time above threshold: %s\n"+
			"- Next possible alert after: %s (cooldown)\n\n"+
			"What this means:\n"+
			"%s usage stayed above your configured threshold long enough to be considered sustained, not a short spike.\n\n"+
			"Suggested checks:\n"+
			"1. Run: top or htop to inspect resource-heavy processes\n"+
			"2. Check service logs: journalctl -xe\n"+
			"3. Verify recent deployments or batch jobs\n\n"+
			"This message was sent by cpu-alert.\n",
		metricName,
		hostname,
		localTime,
		utcTime,
		prettyMetric,
		usage,
		threshold,
		sustainedFor.Truncate(time.Second),
		cooldown.Truncate(time.Second),
		prettyMetric,
	)

	msg := []byte("To: " + s.To + "\r\n" +
		"From: " + s.From + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body +
		"\r\n")

	if err := s.sendMailWithTimeout(host, auth, []string{s.To}, msg); err != nil {
		return fmt.Errorf("send mail: %w", err)
	}

	return nil
}

// SendMetricAlertWithRetry retries email sending up to maxAttempts.
func (s *Sender) SendMetricAlertWithRetry(metricName string, usage, threshold float64, sustainedFor, cooldown time.Duration, maxAttempts int) error {
	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := s.SendMetricAlert(metricName, usage, threshold, sustainedFor, cooldown)
		if err == nil {
			return nil
		}

		lastErr = err
		if attempt < maxAttempts {
			time.Sleep(2 * time.Second)
		}
	}

	return fmt.Errorf("email failed after %d attempts: %w", maxAttempts, lastErr)
}

// sendMailWithTimeout performs SMTP operations with hard network time limits.
func (s *Sender) sendMailWithTimeout(host string, auth smtp.Auth, recipients []string, msg []byte) error {
	dialer := net.Dialer{Timeout: s.DialTimeout}
	conn, err := dialer.Dial("tcp", s.SMTPAddr)
	if err != nil {
		return err
	}

	deadline := time.Now().Add(s.CommandTimeout)
	if err := conn.SetDeadline(deadline); err != nil {
		_ = conn.Close()
		return err
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		_ = conn.Close()
		return err
	}
	defer func() {
		_ = client.Close()
	}()

	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
		if err := client.StartTLS(tlsConfig); err != nil {
			return err
		}
	}

	if ok, _ := client.Extension("AUTH"); ok {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}

	if err := client.Mail(s.From); err != nil {
		return err
	}
	for _, rcpt := range recipients {
		if err := client.Rcpt(rcpt); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		_ = w.Close()
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	if err := client.Quit(); err != nil {
		return err
	}

	return nil
}
