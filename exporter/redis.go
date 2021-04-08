package exporter

import (
	"crypto/tls"
	"strings"

	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
)

func (e *Exporter) connectToRedis() (redis.Conn, error) {
	options := []redis.DialOption{
		redis.DialConnectTimeout(e.options.ConnectionTimeouts),
		redis.DialReadTimeout(e.options.ConnectionTimeouts),
		redis.DialWriteTimeout(e.options.ConnectionTimeouts),

		redis.DialTLSConfig(&tls.Config{
			InsecureSkipVerify: e.options.SkipTLSVerification,
			Certificates:       e.options.ClientCertificates,
			RootCAs:            e.options.CaCertificates,
		}),
	}

	if e.options.User != "" {
		options = append(options, redis.DialUsername(e.options.User))
	}

	uri := e.server.Addr
	if !strings.Contains(uri, "://") {
		uri = "redis://" + uri
	}

	if e.options.Password != "" {
		options = append(options, redis.DialPassword(e.options.Password))
	} else if e.server.Password != "" {
		options = append(options, redis.DialPassword(e.server.Password))
	} else if e.options.PasswordMap[uri] != "" {
		options = append(options, redis.DialPassword(e.options.PasswordMap[uri]))
	}

	log.Debugf("Trying DialURL(): %s", uri)
	c, err := redis.DialURL(uri, options...)
	if err != nil {
		log.Debugf("DialURL() failed, err: %s", err)
		if frags := strings.Split(e.server.Addr, "://"); len(frags) == 2 {
			log.Debugf("Trying: Dial(): %s %s", frags[0], frags[1])
			c, err = redis.Dial(frags[0], frags[1], options...)
		} else {
			log.Debugf("Trying: Dial(): tcp %s", e.server.Addr)
			c, err = redis.Dial("tcp", e.server.Addr, options...)
		}
	}
	return c, err
}

func doRedisCmd(c redis.Conn, cmd string, args ...interface{}) (interface{}, error) {
	log.Debugf("c.Do() - running command: %s %s", cmd, args)
	res, err := c.Do(cmd, args...)
	if err != nil {
		log.Debugf("c.Do() - err: %s", err)
	}
	log.Debugf("c.Do() - done")
	return res, err
}
