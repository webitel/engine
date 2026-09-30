package rabbit

type Table map[string]any

type QueueConfig struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       Table
}

func NewQueueConfig(name string, opts ...QueueConfigOption) *QueueConfig {
	cfg := &QueueConfig{
		Name:       name,
		Durable:    false,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     true,
		Args:       make(Table),
	}

	for _, o := range opts {
		o(cfg)
	}

	return cfg
}

type QueueConfigOption func(*QueueConfig)

func WithQueueName(name string) QueueConfigOption {
	return func(qc *QueueConfig) {
		qc.Name = name
	}
}

func WithQueueDurable(d bool) QueueConfigOption {
	return func(qc *QueueConfig) {
		qc.Durable = d
	}
}

func WithQueueAutoDelete(ad bool) QueueConfigOption {
	return func(qc *QueueConfig) {
		qc.AutoDelete = ad
	}
}

func WithQueueExclusive(e bool) QueueConfigOption {
	return func(qc *QueueConfig) {
		qc.Exclusive = e
	}
}

func WithQueueNoWait(w bool) QueueConfigOption {
	return func(qc *QueueConfig) {
		qc.NoWait = w
	}
}

func WithQueueArg(key string, value any) QueueConfigOption {
	return func(qc *QueueConfig) {
		if qc.Args == nil {
			return
		}

		qc.Args[key] = value
	}
}

type QueueBindConfig struct {
	Name     string
	Key      string
	Exchange string
	NoWait   bool
	Args     Table
}

func NewQueueBindConfig(name, key, exchange string, opts ...QueueBindConfigOption) *QueueBindConfig {
	cfg := &QueueBindConfig{
		Name:     name,
		Key:      key,
		Exchange: exchange,
		NoWait:   true,
		Args:     make(Table),
	}

	for _, o := range opts {
		o(cfg)
	}

	return cfg
}

type QueueBindConfigOption func(*QueueBindConfig)

func WithQueueBindName(name string) QueueBindConfigOption {
	return func(qbc *QueueBindConfig) {
		qbc.Name = name
	}
}

func WithQueueBindKey(key string) QueueBindConfigOption {
	return func(qbc *QueueBindConfig) {
		qbc.Key = key
	}
}

func WithQueueBindExchange(exchange string) QueueBindConfigOption {
	return func(qbc *QueueBindConfig) {
		qbc.Exchange = exchange
	}
}

func WithQueueBindNoWait(noWait bool) QueueBindConfigOption {
	return func(qbc *QueueBindConfig) {
		qbc.NoWait = noWait
	}
}
