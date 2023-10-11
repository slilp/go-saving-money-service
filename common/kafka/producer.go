package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"log"
	"os"

	"github.com/IBM/sarama"
)

type Relayer struct {
	producer  sarama.SyncProducer
	consumer  sarama.ConsumerGroup
	client    sarama.Client
	topics    map[string]string
	listeners map[string][]SubscribeFunc
}

func NewRelayer() KafkaPubSub {

	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Partitioner = sarama.NewRandomPartitioner
	saramaCfg.Producer.RequiredAcks = sarama.WaitForAll
	saramaCfg.Producer.Return.Successes = true

	// if config.CertPath != "" {
	// 	saramaCfg.Net.TLS.Config = tlsConfig(config.CertPath)
	// 	saramaCfg.Net.TLS.Enable = true
	// }

	client, err := sarama.NewClient([]string{"Brokers"}, saramaCfg)
	if err != nil {
		panic(err)
	}

	consumer, err := sarama.NewConsumerGroupFromClient("Consumergroup", client)
	if err != nil {
		panic(err)
	}

	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		panic(err)
	}

	// return for closing client.Close()

	relayer := &Relayer{
		producer,
		consumer,
		client,
		map[string]string{"2": "config.TopicMapping"},
		map[string][]SubscribeFunc{},
	}
	return relayer
}

func (r *Relayer) Publish(ctx context.Context, eventName string, payload any) error {
	bytes, err := json.Marshal(payload)
	if err != nil {
		// log.Error("emitter json.Marshal(event) failed", zap.Error(err), zap.Any("event", payload))
		return err
	}
	topic, found := r.TopicForEvent(eventName)
	if !found {
		return nil
	}

	// log.Debug("kafka sending message")
	_, _, err = r.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(bytes),
	})
	if err != nil {
		// log.Error("failed to send kafka event", zap.Error(err), zap.Any("event", payload))
	}
	// log.Debug("successfully sent event", zap.Any("event", payload))

	return nil
}

func (r *Relayer) Subscribe(eventName string, handlerFunc func(payload []byte)) error {
	topic, found := r.TopicForEvent(eventName)
	if !found {
		return nil
	}

	topicListeners, found := r.listeners[topic]
	if !found {
		topicListeners = []SubscribeFunc{}
		r.listeners[topic] = topicListeners
	}

	r.listeners[topic] = append(topicListeners, handlerFunc)

	return nil
}

func (r *Relayer) Stop() {
	r.client.Close()
}

func (r *Relayer) Start() {

	topics := make([]string, 0, len(r.listeners))
	for topic := range r.listeners {
		topics = append(topics, topic)
	}
	if len(topics) == 0 {
		// r.log.Info("no subscriptions toany kafka topic")
	}

	handlerRelayer := func(message *sarama.ConsumerMessage) {
		r.handleInboud(message)
	}

	consumer := &Consumer{
		ready:   make(chan bool),
		Consume: handlerRelayer,
	}

	ctx, _ := context.WithCancel(context.Background())

	go func() {
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := r.consumer.Consume(ctx, topics, consumer); err != nil {
				log.Panicf("Error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready // Await till the consumer has been set up
	log.Println("Sarama consumer up and running!...")

}

func (r *Relayer) TopicForEvent(eventName string) (string, bool) {
	topic, found := r.topics[eventName]
	return topic, found
}

func (r *Relayer) EventForTopic(topic string) (string, bool) {
	for eventName, topicName := range r.topics {
		if topic == topicName {
			return eventName, true
		}
	}
	return "", false
}

func (r *Relayer) handleInboud(message *sarama.ConsumerMessage) {
	// monitoring.Logger().Info("handleInboud: " + message.Topic)
	_, found := r.EventForTopic(message.Topic)
	if !found {
		// monitoring.Logger().Warn("received message on topic with no internal listeners")
		return
	}

	// monitoring.Logger().Info("listeners", zap.Any("map", len(r.listeners)))
	listeners, found := r.listeners[message.Topic]
	// monitoring.Logger().Info("found listener count", zap.Int("len", len(listeners)), zap.Bool("found", found))

	for _, listener := range listeners {
		listener(message.Value)
	}
}

func tlsConfig(pemPath string) *tls.Config {

	conf := tls.Config{}
	conf.RootCAs = x509.NewCertPool()

	// Load CA cert
	caCert, err := os.ReadFile(pemPath)
	if err != nil {
		panic(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	conf.RootCAs = caCertPool

	return &conf
}
