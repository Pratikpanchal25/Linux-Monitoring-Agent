package daemon

import (
	"context"
	"log"
	"time"

	"watchd/internal/config"
	"watchd/internal/email"
	"watchd/internal/metric"
)

const (
	startupRetryDelay = 5 * time.Second
	alertFailureBackoff = 60 * time.Second
	alertQueueSize = 16
)

type metricState struct {
	collector        metric.Collector
	threshold        float64
	aboveThreshold   bool
	aboveSince       time.Time
	lastAlertTime    time.Time
	nextAlertAttempt time.Time
	pendingAlert     bool
}

type alertEvent struct {
	metricName  string
	usage       float64
	threshold   float64
	sustainedFor time.Duration
	cooldown    time.Duration
}

type alertResult struct {
	metricName string
	err        error
	when       time.Time
}

// Monitor runs the infinite metric check loop.
type Monitor struct {
	cfg    config.Config
	sender *email.Sender
	states map[string]*metricState
}

// New creates a monitor from validated config.
func New(cfg config.Config) *Monitor {
	sender := email.New(cfg.Email.To, cfg.Email.From, cfg.Email.SMTP, cfg.Email.Password)

	states := map[string]*metricState{
		metric.NameCPU: {
			collector: metric.NewCPUCollector(),
			threshold: cfg.Thresholds.CPU,
		},
		metric.NameMemory: {
			collector: metric.NewMemoryCollector(),
			threshold: cfg.Thresholds.Memory,
		},
	}

	return &Monitor{cfg: cfg, sender: sender, states: states}
}

// Run checks metrics forever until context is canceled.
func (m *Monitor) Run(ctx context.Context) error {
	for name, state := range m.states {
		for {
			err := state.collector.Init()
			if err == nil {
				break
			}

			log.Printf("initial %s metric setup failed, retrying in %s: %v", name, startupRetryDelay, err)
			select {
			case <-ctx.Done():
				log.Println("shutdown signal received during startup")
				return nil
			case <-time.After(startupRetryDelay):
			}
		}
	}

	ticker := time.NewTicker(m.cfg.IntervalDuration())
	defer ticker.Stop()

	alertQueue := make(chan alertEvent, alertQueueSize)
	alertResults := make(chan alertResult, alertQueueSize)
	go m.alertWorker(ctx, alertQueue, alertResults)

	for {
		select {
		case <-ctx.Done():
			log.Println("shutdown signal received, stopping loop")
			return nil
		case result := <-alertResults:
			state, ok := m.states[result.metricName]
			if !ok {
				continue
			}

			state.pendingAlert = false
			if result.err != nil {
				log.Printf("%s alert send failed: %v", result.metricName, result.err)
				state.nextAlertAttempt = result.when.Add(alertFailureBackoff)
				continue
			}

			state.lastAlertTime = result.when
			state.nextAlertAttempt = time.Time{}
			state.aboveSince = result.when
			log.Printf("%s alert email sent to %s", result.metricName, m.cfg.Email.To)
		case now := <-ticker.C:
			for name, state := range m.states {
				usage, err := state.collector.Sample()
				if err != nil {
					log.Printf("%s sample failed: %v", name, err)
					continue
				}

				log.Printf("%s usage %.2f%% (threshold %.2f%%)", name, usage, state.threshold)

				if usage >= state.threshold {
					if !state.aboveThreshold {
						state.aboveThreshold = true
						state.aboveSince = now
						log.Printf("%s crossed threshold, sustained timer started", name)
					}

					if now.Sub(state.aboveSince) < m.cfg.DurationDuration() {
						continue
					}
					if state.pendingAlert {
						continue
					}
					if !state.nextAlertAttempt.IsZero() && now.Before(state.nextAlertAttempt) {
						remaining := state.nextAlertAttempt.Sub(now)
						log.Printf("%s alert retry backoff active, next attempt in %s", name, remaining.Truncate(time.Second))
						continue
					}
					if !state.lastAlertTime.IsZero() && now.Sub(state.lastAlertTime) < m.cfg.CooldownDuration() {
						remaining := m.cfg.CooldownDuration() - now.Sub(state.lastAlertTime)
						log.Printf("%s cooldown active, next alert in %s", name, remaining.Truncate(time.Second))
						continue
					}

					event := alertEvent{
						metricName: name,
						usage: usage,
						threshold: state.threshold,
						sustainedFor: now.Sub(state.aboveSince),
						cooldown: m.cfg.CooldownDuration(),
					}

					select {
					case alertQueue <- event:
						state.pendingAlert = true
					default:
						// Non-blocking design: never stall the metric loop on alert queue pressure.
						log.Printf("%s alert queue is full, deferring alert", name)
						state.nextAlertAttempt = now.Add(alertFailureBackoff)
					}
					continue
				}

				if state.aboveThreshold {
					log.Printf("%s dropped below threshold, sustained timer reset", name)
				}
				state.aboveThreshold = false
			}
		}
	}
}

func (m *Monitor) alertWorker(ctx context.Context, events <-chan alertEvent, results chan<- alertResult) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-events:
			err := m.sender.SendMetricAlertWithRetry(
				event.metricName,
				event.usage,
				event.threshold,
				event.sustainedFor,
				event.cooldown,
				3,
			)

			result := alertResult{metricName: event.metricName, err: err, when: time.Now()}
			select {
			case results <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}
