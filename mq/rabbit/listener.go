package rabbit

import (
	"fmt"
	"github.com/webitel/wlog"
)

func (a *AMQP) Listen() {
	defer func() {
		wlog.Info("close amqp listener")
		close(a.stopped)
	}()

	for {
		select {
		case err, ok := <-a.errorChan:
			if !ok {
				break
			}
			wlog.Error(fmt.Sprintf("amqp connection receive error: %s", err.Error()))
			a.initConnection()
		case <-a.stop:
			for _, q := range a.domainQueues {
				q.Stop()
			}
			wlog.Debug("listener call received stop signal")
			return

		case q := <-a.registerDomainQueue:
			q.Start()
		case q := <-a.unRegisterDomainQueue:
			q.Stop()
		}
	}

}
