package msg

import (
	"time"
)

type Service struct {
	UUID        string
	Name        string
	Address4    string
	Address6    string
	Environment string
	Region      string
	Port        uint16
	TTL         uint32 // seconds
	Expires     time.Time
	host        string // uuid.<skydns.local.>
}

// RemainingTTL returns the amount of time remaining before expiration.
func (s *Service) RemainingTTL() uint32 {
	d := s.Expires.Sub(time.Now())
	ttl := uint32(d.Seconds())
	if ttl < 1 {
		return 0
	}
	return ttl
}

// UpdateTTL updates the TTL property to the RemainingTTL.
func (s *Service) UpdateTTL() {
	s.TTL = s.RemainingTTL()
}
