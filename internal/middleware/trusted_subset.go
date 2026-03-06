package middleware

import (
	"net"
	"net/http"

	"github.com/mrhyman/shortner/internal/logger"
)

// WithTrustedSubnet проверяет, что IP-адрес клиента входит в доверенную подсеть.
// Если trustedSubnetCIDR пустая строка, доступ запрещён для всех.
func WithTrustedSubnet(trustedSubnetCIDR string) func(http.HandlerFunc) http.HandlerFunc {
	var trustedSubnet *net.IPNet

	log := logger.Get()

	if trustedSubnetCIDR != "" {
		_, subnet, err := net.ParseCIDR(trustedSubnetCIDR)
		if err != nil {
			log.With("cidr", trustedSubnetCIDR, "err", err.Error()).Error("Invalid trusted subnet CIDR")
		} else {
			trustedSubnet = subnet
		}
	}

	log.Info(trustedSubnetCIDR)

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Если доверенная подсеть не задана, запрещаем доступ
			if trustedSubnet == nil {
				log.Warn("Access denied: trusted subnet not configured")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Получаем IP из заголовка X-Real-IP
			realIP := r.Header.Get("X-Real-IP")
			if realIP == "" {
				log.Warn("Access denied: X-Real-IP header not provided")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Парсим IP-адрес
			ip := net.ParseIP(realIP)
			if ip == nil {
				log.With("ip", realIP).Warn("Access denied: invalid IP address")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Проверяем принадлежность к подсети
			if !trustedSubnet.Contains(ip) {
				log.With("ip", realIP, "subnet", trustedSubnet.String()).Warn("Access denied: IP not in trusted subnet")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next(w, r)
		}
	}
}
