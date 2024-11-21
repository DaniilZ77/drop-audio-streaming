package minio

import "time"

// Option -.
type Option func(*Minio)

// ConnAttempts -.
func ConnAttempts(attempts int) Option {
	return func(c *Minio) {
		c.connAttempts = attempts
	}
}

// ConnTimeout -.
func ConnTimeout(timeout time.Duration) Option {
	return func(c *Minio) {
		c.connTimeout = timeout
	}
}
