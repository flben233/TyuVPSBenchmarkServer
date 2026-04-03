package util

import (
	"context"
	qdb "github.com/questdb/go-questdb-client/v4"
)

type QDBPool struct {
	pool   chan qdb.LineSender
	url    string
	user   string
	passwd string
	size   int
}

func NewQDBPool(ctx context.Context, url, user, pass string, size int) (*QDBPool, error) {
	pool := make(chan qdb.LineSender, size)
	for i := 0; i < size; i++ {
		qdbSender, err := qdb.NewLineSender(ctx, qdb.WithHttp(), qdb.WithAddress(url), qdb.WithBasicAuth(user, pass))
		if err != nil {
			for sender := range pool {
				sender.Close(ctx)
			}
			return nil, err
		}
		pool <- qdbSender
	}
	return &QDBPool{
		pool:   pool,
		url:    url,
		user:   user,
		passwd: pass,
		size:   size,
	}, nil
}

func (p *QDBPool) Get() qdb.LineSender {
	return <-p.pool
}

func (p *QDBPool) Put(sender qdb.LineSender) {
	p.pool <- sender
}
