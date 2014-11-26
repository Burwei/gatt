package gatt

import (
	"errors"
	"sync"
	"time"
)

type notifier struct {
	l2c    *l2cap
	char   *Characteristic
	maxlen int
	donemu sync.RWMutex
	done   bool
	// This throttle prevents multiple subsequent notifications from
	// stepping on each others' toes. This toe-stepping appears to
	// happen at both the HCI and the link layer.
	throttle *time.Ticker
}

func newNotifier(l2c *l2cap, c *Characteristic, maxlen int) *notifier {
	return &notifier{
		l2c:      l2c,
		char:     c,
		maxlen:   maxlen,
		throttle: time.NewTicker(50 * time.Millisecond),
	}
}

func (n *notifier) Write(data []byte) (int, error) {
	if n.Done() {
		return 0, errors.New("central stopped notifications")
	}
	<-n.throttle.C
	if err := n.l2c.sendNotification(n.char, data); err != nil {
		return 0, err
	}
	return len(data), nil
}

func (n *notifier) Cap() int {
	return n.maxlen
}

func (n *notifier) Done() bool {
	n.donemu.RLock()
	done := n.done
	n.donemu.RUnlock()
	return done
}

func (n *notifier) stop() {
	n.donemu.Lock()
	n.done = true
	n.donemu.Unlock()
	n.throttle.Stop()
}
