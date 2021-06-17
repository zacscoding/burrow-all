package kafka

import (
	"context"
	"github.com/Shopify/sarama"
	"log"
	"time"
)

type ConsumerConfig struct {
	KafkaCfg        *Config
	Name            string
	GroupID         string
	Topic           string
	ConsumeInterval time.Duration
	ShouldFail      bool
}

type Consumer struct {
	cfg    *ConsumerConfig
	client sarama.ConsumerGroup

	proceed     int
	lastConsume time.Time
}

func NewConsumer(cfg *ConsumerConfig) (*Consumer, error) {
	c := sarama.NewConfig()
	v, err := sarama.ParseKafkaVersion(cfg.KafkaCfg.KafkaVersion)
	if err != nil {
		return nil, err
	}
	c.Version = v
	c.Consumer.Group.Session.Timeout = time.Second * 6
	c.Consumer.Group.Heartbeat.Interval = time.Second
	c.Consumer.Offsets.AutoCommit.Enable = false
	c.Consumer.Return.Errors = true
	c.Consumer.Offsets.Initial = sarama.OffsetNewest
	c.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
	consumerGroup, err := sarama.NewConsumerGroup(cfg.KafkaCfg.Brokers, cfg.GroupID, c)
	if err != nil {
		return nil, err
	}
	consumer := &Consumer{
		cfg:         cfg,
		client:      consumerGroup,
		proceed:     0,
		lastConsume: time.Now(),
	}
	go consumer.mainLoop()
	return consumer, nil
}

func (c *Consumer) Stop() {
	c.client.Close()
}

func (c *Consumer) SetShouldFail(shouldFail bool) {
	c.cfg.ShouldFail = shouldFail
}

func (c *Consumer) SetConsumeInterval(interval time.Duration) {
	c.cfg.ConsumeInterval = interval
}

func (c *Consumer) mainLoop() {
	ctx := context.Background()
	for {
		if err := c.client.Consume(ctx, []string{c.cfg.Topic}, c); err != nil {
			if sarama.ErrClosedClient == err {
				log.Printf("[Consumer-%s] terminate", c.cfg.Name)
				return
			}
			log.Printf("Consumer-%s failed to consume. err: %v", c.cfg.Name, err)
		}
	}
}

func (c *Consumer) Setup(session sarama.ConsumerGroupSession) error {
	log.Printf("[Consumer-%s] Setup is called:%v", c.cfg.Name, session)
	return nil
}

func (c *Consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Printf("[Consumer-%s]Cleanup is called:%v", c.cfg.Name, session)
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		if c.cfg.ConsumeInterval != 0 {
			timer := time.NewTimer(c.lastConsume.Add(c.cfg.ConsumeInterval).Sub(time.Now()))
			<-timer.C
			c.lastConsume = time.Now()
		}
		c.proceed++
		if c.proceed%50 == 0 {
			log.Printf("[Consumer-%s] consume message proceed: %d", c.cfg.Name, c.proceed)
		}
		if c.cfg.ShouldFail {
			continue
		} else {
			session.MarkMessage(msg, "")
			session.Commit()
		}
	}
	return nil
}
