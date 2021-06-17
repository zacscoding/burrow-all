package kafka

import (
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"log"
	"time"
)

type ProducerConfig struct {
	KafkaCfg *Config
	Name     string
	Topic    string
	Interval time.Duration
}

type Producer struct {
	cfg     *ProducerConfig
	client  sarama.SyncProducer
	closeCh chan bool
}

func NewProducer(cfg *ProducerConfig) (*Producer, error) {
	c := sarama.NewConfig()
	v, err := sarama.ParseKafkaVersion(cfg.KafkaCfg.KafkaVersion)
	if err != nil {
		return nil, err
	}
	c.Version = v
	c.Producer.Return.Successes = true

	syncProducer, err := sarama.NewSyncProducer(cfg.KafkaCfg.Brokers, c)
	if err != nil {
		return nil, err
	}

	p := &Producer{
		cfg:     cfg,
		client:  syncProducer,
		closeCh: make(chan bool),
	}
	go p.mainLoop()
	return p, nil
}

func (p *Producer) Stop() {
	select {
	case <-p.closeCh:
		return
	default:
		close(p.closeCh)
	}
	p.client.Close()
}

func (p *Producer) mainLoop() {
	log.Printf("[%s] start producer", p.cfg.Name)
	ticker := time.NewTicker(p.cfg.Interval)
	defer func() {
		ticker.Stop()
		p.client.Close()
	}()

	count := 0
	for {
		select {
		case <-ticker.C:
			m, _ := json.Marshal(&Message{
				Value: fmt.Sprintf("Message-%d", count),
			})
			_, _, err := p.client.SendMessage(&sarama.ProducerMessage{
				Topic: p.cfg.Topic,
				Value: sarama.StringEncoder(m),
			})
			if err != nil {
				log.Println("failed to produce message", err)
			} else {
				count++
			}
		case <-p.closeCh:
			log.Printf("[%s] terminate producer", p.cfg.Name)
			return
		}
	}
}
