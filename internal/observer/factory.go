package observer

import "github.com/mrhyman/shortner/internal/config"

func SetupObservers(cfg config.AppConfig) (*Publisher, func() error, error) {
	p := NewPublisher()
	var closers []func() error

	if cfg.AuditFile != "" {
		fo, err := NewFileObserver(cfg.AuditFile)
		if err != nil {
			return nil, nil, err
		}
		p.Subscribe(fo)
		closers = append(closers, fo.Close)
	}

	if cfg.AuditURL != "" {
		ro := NewRemoteObserver(cfg.AuditURL)
		p.Subscribe(ro)
	}

	// Функция для graceful shutdown
	cleanup := func() error {
		for _, close := range closers {
			if err := close(); err != nil {
				return err
			}
		}
		return nil
	}

	return p, cleanup, nil
}
