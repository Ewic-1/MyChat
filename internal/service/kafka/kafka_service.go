package kafka

import (
	"context"
	"errors"
	myconfig "mychat_server/internal/config"
	"mychat_server/pkg/utils/zlog"
	"strings"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

var ctx = context.Background()

type kafkaService struct {
	ChatWriter *kafka.Writer
	ChatReader *kafka.Reader
	KafkaConn  *kafka.Conn
	initOnce   sync.Once
}

var KafkaService = new(kafkaService)

// KafkaInit 初始化kafka
func (k *kafkaService) KafkaInit() {
	k.initOnce.Do(func() {
		kafkaConfig := myconfig.GetConfig().KafkaConfig
		brokers := parseBrokers(kafkaConfig.HostPort)
		if len(brokers) == 0 {
			zlog.Error("kafka hostPort is empty")
			return
		}

		k.ChatWriter = &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Topic:                  kafkaConfig.ChatTopic,
			Balancer:               &kafka.Hash{},
			WriteTimeout:           kafkaConfig.Timeout * time.Second,
			RequiredAcks:           kafka.RequireOne,
			AllowAutoTopicCreation: false,
		}

		k.ChatReader = kafka.NewReader(kafka.ReaderConfig{
			Brokers:        brokers,
			Topic:          kafkaConfig.ChatTopic,
			CommitInterval: kafkaConfig.Timeout * time.Second,
			GroupID:        "chat",
			StartOffset:    kafka.LastOffset,
		})
	})
}

func (k *kafkaService) KafkaClose() {
	if k.ChatWriter != nil {
		if err := k.ChatWriter.Close(); err != nil {
			zlog.Error(err.Error())
		}
	}
	if k.ChatReader != nil {
		if err := k.ChatReader.Close(); err != nil {
			zlog.Error(err.Error())
		}
	}
	if k.KafkaConn != nil {
		if err := k.KafkaConn.Close(); err != nil {
			zlog.Error(err.Error())
		}
	}
}

func (k *kafkaService) PublishChatMessage(key string, payload []byte) error {
	if k.ChatWriter == nil {
		return errors.New("kafka writer not initialized")
	}
	return k.ChatWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: payload,
		Time:  time.Now(),
	})
}

func (k *kafkaService) ReadChatMessage(ctx context.Context) (kafka.Message, error) {
	if k.ChatReader == nil {
		return kafka.Message{}, errors.New("kafka reader not initialized")
	}
	return k.ChatReader.ReadMessage(ctx)
}

// CreateTopic 创建topic
func (k *kafkaService) CreateTopic() {
	// 如果已经有topic了，就不创建了
	kafkaConfig := myconfig.GetConfig().KafkaConfig

	chatTopic := kafkaConfig.ChatTopic
	brokers := parseBrokers(kafkaConfig.HostPort)
	if len(brokers) == 0 {
		zlog.Error("kafka hostPort is empty")
		return
	}

	// 连接至任意kafka节点
	var err error
	k.KafkaConn, err = kafka.Dial("tcp", brokers[0])
	if err != nil {
		zlog.Error(err.Error())
	}

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             chatTopic,
			NumPartitions:     kafkaConfig.Partition,
			ReplicationFactor: 1,
		},
	}

	// 创建topic
	if err = k.KafkaConn.CreateTopics(topicConfigs...); err != nil {
		zlog.Error(err.Error())
	}

}

func parseBrokers(hostPort string) []string {
	items := strings.Split(hostPort, ",")
	brokers := make([]string, 0, len(items))
	for _, item := range items {
		broker := strings.TrimSpace(item)
		if broker == "" {
			continue
		}
		brokers = append(brokers, broker)
	}
	return brokers
}
