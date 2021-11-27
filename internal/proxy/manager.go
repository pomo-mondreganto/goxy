package proxy

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"goxy/internal/common"
	"goxy/internal/export"
	"goxy/internal/filters"
	"goxy/internal/models"
	"goxy/internal/proxy/http"
	"goxy/internal/proxy/tcp"

	"github.com/sirupsen/logrus"
)

var (
	ErrNoSuchProxy = errors.New("no such proxy")
)

func NewManager(cfg *common.ProxyConfig, producer *export.ProducerClient) (*Manager, error) {
	rs, err := filters.NewRuleSet(cfg.Rules)
	if err != nil {
		return nil, fmt.Errorf("parsing rules: %w", err)
	}
	logrus.Debug("Parsed rules:")
	for name, rule := range rs.Rules {
		logrus.Debugf("%s -> %s", name, rule)
	}

	proxies := make([]Proxy, 0)
	for _, s := range cfg.Services {
		var p Proxy
		if s.Type == "tcp" {
			if p, err = tcp.NewProxy(s, rs, producer); err != nil {
				logrus.Fatalf("Error creating tcp proxy: %v", err)
			}
		} else if s.Type == "http" {
			if p, err = http.NewProxy(s, rs, producer); err != nil {
				logrus.Fatalf("Error creating http proxy: %v", err)
			}
		} else {
			return nil, fmt.Errorf("invalid proxy type: %s", s.Type)
		}
		proxies = append(proxies, p)
	}

	m := &Manager{proxies}
	return m, nil
}

type Manager struct {
	proxies []Proxy
}

func (m *Manager) StartAll() error {
	for i, p := range m.proxies {
		if err := p.Start(); err != nil {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			for j := 0; j < i; j += 1 {
				if serr := m.proxies[j].Shutdown(ctx); serr != nil {
					logrus.Errorf("Error shutting down proxy %v forcefully: %v", m.proxies[j], serr)
				}
			}
			cancel()
			return fmt.Errorf("starting proxy %v: %w", p, err)
		}
	}
	return nil
}

func (m *Manager) Shutdown(ctx context.Context) error {
	wg := sync.WaitGroup{}
	wg.Add(len(m.proxies))
	errCh := make(chan error)
	for _, p := range m.proxies {
		go func(p Proxy) {
			defer wg.Done()
			if err := p.Shutdown(ctx); err != nil {
				logrus.Errorf("Error shutting down proxy %v: %v", p, err)
				select {
				case errCh <- err:
				default:
				}
			}
		}(p)
	}
	wg.Wait()
	select {
	case err := <-errCh:
		return fmt.Errorf("error shutting down proxy: %w", err)
	default:
		return nil
	}
}

func (m Manager) DumpProxies() []models.ProxyDescription {
	result := make([]models.ProxyDescription, 0, len(m.proxies))
	for i, p := range m.proxies {
		proxyID := i + 1
		fs := p.GetFilters()
		descriptions := make([]models.FilterDescription, 0, len(fs))
		for j, f := range fs {
			desc := models.FilterDescription{
				ID:      j + 1,
				ProxyID: proxyID,
				Rule:    f.GetRule().String(),
				Verdict: f.GetVerdict().String(),
				Enabled: f.IsEnabled(),
				Alert:   f.GetAlert(),
			}
			descriptions = append(descriptions, desc)
		}
		desc := models.ProxyDescription{
			ID:                 proxyID,
			Service:            p.GetConfig(),
			Listening:          p.GetListening(),
			FilterDescriptions: descriptions,
		}
		result = append(result, desc)
	}
	return result
}

func (m *Manager) SetProxyListening(proxyID int, listening bool) error {
	if proxyID < 1 || proxyID > len(m.proxies) {
		return ErrNoSuchProxy
	}
	m.proxies[proxyID-1].SetListening(listening)
	return nil
}

func (m *Manager) SetFilterState(proxyID, filterID int, enabled, alert bool) error {
	if proxyID < 1 || proxyID > len(m.proxies) {
		return ErrNoSuchProxy
	}
	p := m.proxies[proxyID-1]
	if err := p.SetFilterState(filterID-1, enabled, alert); err != nil {
		return fmt.Errorf("setting filter enabled for proxy %v: %w", p, err)
	}
	return nil
}
