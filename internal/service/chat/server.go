package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"mychat_server/internal/config"
	"mychat_server/internal/dto/request"
	"mychat_server/internal/model"
	mykafka "mychat_server/internal/service/kafka"
	"mychat_server/pkg/enum/contact/contact_status_enum"
	"mychat_server/pkg/utils/random"
	"mychat_server/pkg/utils/zlog"
)

const (
	messageModeChannel = "channel"
	messageModeKafka   = "kafka"
	messageModeHybrid  = "hybrid"

	defaultChannelQueueSize     = 1024
	defaultHighWatermarkRatio   = 0.8
	defaultHighWatermarkSeconds = 3
	defaultOffloadPercent       = 50
)

func InitRuntime() {
	ChatServer.InitRuntime()
}

func StopRuntime() {
	ChatServer.StopRuntime()
}

func (s *chatServer) InitRuntime() {
	s.runtimeOnce.Do(func() {
		kafkaConfig := config.GetConfig().KafkaConfig
		mode := normalizeMessageMode(kafkaConfig.MessageMode)

		s.runtimeCtx, s.runtimeCancel = context.WithCancel(context.Background())

		s.kafkaEnabled = mode == messageModeKafka || mode == messageModeHybrid
		s.hybridEnabled = mode == messageModeHybrid

		if mode != messageModeKafka {
			queueSize := normalizeChannelQueueSize(kafkaConfig.ChannelQueueSize)
			s.dispatchQueue = make(chan wsMessage, queueSize)
			s.runtimeWg.Add(1)
			go s.consumeChannelMessages()
			zlog.Info(fmt.Sprintf("chat channel dispatcher enabled: queue_size=%d", queueSize))
		}

		if s.kafkaEnabled {
			mykafka.KafkaService.KafkaInit()
			s.runtimeWg.Add(1)
			go s.consumeKafkaMessages()
		}

		if s.hybridEnabled {
			s.highWatermarkRatio = normalizeHighWatermarkRatio(kafkaConfig.HighWatermarkRatio)
			s.highWatermarkSeconds = normalizeHighWatermarkSeconds(kafkaConfig.HighWatermarkSeconds)
			s.offloadPercent = normalizeOffloadPercent(kafkaConfig.OffloadPercent)

			s.runtimeWg.Add(1)
			go s.monitorChannelPressure()

			zlog.Info(fmt.Sprintf("chat hybrid offload enabled: high_water_ratio=%.2f, high_water_seconds=%d, offload_percent=%d",
				s.highWatermarkRatio,
				s.highWatermarkSeconds,
				s.offloadPercent,
			))
		}

		zlog.Info("chat runtime started with " + mode + " mode")
	})
}

func (s *chatServer) StopRuntime() {
	if s.runtimeCancel != nil {
		s.runtimeCancel()
	}
	if s.kafkaEnabled {
		mykafka.KafkaService.KafkaClose()
	}
	s.runtimeWg.Wait()
}

func (s *chatServer) consumeChannelMessages() {
	defer s.runtimeWg.Done()
	for {
		select {
		case <-s.runtimeCtx.Done():
			return
		case msg := <-s.dispatchQueue:
			s.dispatchMessage(msg)
		}
	}
}

func (s *chatServer) consumeKafkaMessages() {
	defer s.runtimeWg.Done()
	for {
		msg, err := mykafka.KafkaService.ReadChatMessage(s.runtimeCtx)
		if err != nil {
			if s.runtimeCtx.Err() != nil || errors.Is(err, context.Canceled) {
				return
			}
			zlog.Error("read kafka chat message failed: " + err.Error())
			time.Sleep(500 * time.Millisecond)
			continue
		}

		var out wsMessage
		if err := json.Unmarshal(msg.Value, &out); err != nil {
			zlog.Error("unmarshal kafka chat message failed: " + err.Error())
			continue
		}
		s.dispatchMessage(out)
	}
}

func (s *chatServer) handleIncomingMessage(clientId string, req request.ChatMessageRequest) {
	if strings.TrimSpace(req.ReceiveId) == "" {
		zlog.Warn("ignore message with empty receive_id")
		return
	}
	if strings.TrimSpace(req.SendId) == "" {
		req.SendId = clientId
	}
	if req.SendId != clientId {
		zlog.Warn("ws send_id mismatch, reset to authenticated client")
		req.SendId = clientId
	}

	message := buildModelMessage(req)
	msg, ret := s.messageDao.SaveMessage(message)
	if ret != 0 {
		zlog.Error("save message failed: " + msg)
		return
	}

	out := buildWsMessage(message)
	s.routeMessage(out)
}

func (s *chatServer) routeMessage(message wsMessage) {
	if s.shouldOffloadToKafka() {
		if err := s.publishByKafka(message); err == nil {
			return
		}
	}

	if s.dispatchQueue == nil {
		if s.kafkaEnabled {
			if err := s.publishByKafka(message); err == nil {
				return
			}
		}
		s.dispatchMessage(message)
		return
	}

	select {
	case s.dispatchQueue <- message:
		return
	default:
		if s.kafkaEnabled {
			if err := s.publishByKafka(message); err == nil {
				zlog.Warn("channel queue full, fallback to kafka")
				return
			}
		}
		s.dispatchMessage(message)
	}
}

func (s *chatServer) publishByKafka(message wsMessage) error {
	payload, err := json.Marshal(message)
	if err != nil {
		zlog.Error("marshal kafka message failed: " + err.Error())
		return err
	}

	if err := mykafka.KafkaService.PublishChatMessage(message.ReceiveId, payload); err != nil {
		zlog.Error("publish kafka message failed: " + err.Error())
		return err
	}
	return nil
}

func (s *chatServer) monitorChannelPressure() {
	defer s.runtimeWg.Done()

	if s.dispatchQueue == nil {
		return
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	highWatermarkSize := calculateHighWatermarkSize(cap(s.dispatchQueue), s.highWatermarkRatio)
	consecutiveHighSeconds := 0

	for {
		select {
		case <-s.runtimeCtx.Done():
			return
		case <-ticker.C:
			queueLen := len(s.dispatchQueue)
			if queueLen >= highWatermarkSize {
				consecutiveHighSeconds++
				if consecutiveHighSeconds >= s.highWatermarkSeconds && !s.isOffloadActive() {
					atomic.StoreInt32(&s.offloadActive, 1)
					zlog.Warn(fmt.Sprintf("channel queue high pressure: len=%d cap=%d threshold=%d seconds=%d, enable kafka offload",
						queueLen,
						cap(s.dispatchQueue),
						highWatermarkSize,
						consecutiveHighSeconds,
					))
				}
				continue
			}

			consecutiveHighSeconds = 0
			if s.isOffloadActive() {
				atomic.StoreInt32(&s.offloadActive, 0)
				zlog.Info("channel queue pressure recovered, disable kafka offload")
			}
		}
	}
}

func (s *chatServer) shouldOffloadToKafka() bool {
	if !s.hybridEnabled || !s.kafkaEnabled || !s.isOffloadActive() {
		return false
	}

	if s.offloadPercent <= 0 {
		return false
	}

	seq := atomic.AddUint64(&s.offloadSeq, 1)
	return int((seq-1)%100) < s.offloadPercent
}

func (s *chatServer) isOffloadActive() bool {
	return atomic.LoadInt32(&s.offloadActive) == 1
}

func normalizeMessageMode(rawMode string) string {
	mode := strings.ToLower(strings.TrimSpace(rawMode))
	switch mode {
	case messageModeKafka:
		return messageModeKafka
	case messageModeHybrid, "channel_kafka", "channel+kafka", "mix":
		return messageModeHybrid
	default:
		return messageModeChannel
	}
}

func normalizeChannelQueueSize(size int) int {
	if size <= 0 {
		return defaultChannelQueueSize
	}
	return size
}

func normalizeHighWatermarkRatio(ratio float64) float64 {
	if ratio <= 0 || ratio >= 1 {
		return defaultHighWatermarkRatio
	}
	return ratio
}

func normalizeHighWatermarkSeconds(seconds int) int {
	if seconds <= 0 {
		return defaultHighWatermarkSeconds
	}
	return seconds
}

func normalizeOffloadPercent(percent int) int {
	if percent <= 0 {
		return defaultOffloadPercent
	}
	if percent > 100 {
		return 100
	}
	return percent
}

func calculateHighWatermarkSize(queueSize int, ratio float64) int {
	highWatermarkSize := int(float64(queueSize) * ratio)
	if highWatermarkSize <= 0 {
		return 1
	}
	if highWatermarkSize > queueSize {
		return queueSize
	}
	return highWatermarkSize
}

func buildModelMessage(req request.ChatMessageRequest) model.Message {
	now := time.Now()
	return model.Message{
		Uuid:       fmt.Sprintf("M%s", random.GetNowAndLenRandomString(11)),
		SessionId:  req.SessionId,
		Type:       req.Type,
		Content:    req.Content,
		URL:        req.Url,
		SendId:     req.SendId,
		SendName:   req.SendName,
		SendAvatar: req.SendAvatar,
		ReceiveId:  req.ReceiveId,
		FileType:   req.FileType,
		FileName:   req.FileName,
		FileSize:   req.FileSize,
		AVData:     req.AVdata,
		Status:     0,
		CreatedAt:  now,
	}
}

func buildWsMessage(message model.Message) wsMessage {
	return wsMessage{
		SessionId:  message.SessionId,
		Type:       message.Type,
		Content:    message.Content,
		Url:        message.URL,
		SendId:     message.SendId,
		SendName:   message.SendName,
		SendAvatar: message.SendAvatar,
		ReceiveId:  message.ReceiveId,
		FileSize:   message.FileSize,
		FileType:   message.FileType,
		FileName:   message.FileName,
		CreatedAt:  message.CreatedAt.Format("2006-01-02 15:04:05"),
		AVdata:     message.AVData,
	}
}

func (s *chatServer) dispatchMessage(message wsMessage) {
	targets := s.resolveTargets(message)
	for clientId := range targets {
		s.sendToClient(clientId, message)
	}
}

func (s *chatServer) resolveTargets(message wsMessage) map[string]struct{} {
	targets := make(map[string]struct{})

	if strings.HasPrefix(message.ReceiveId, "G") {
		msg, members, ret := s.contactDao.GetContactByContactId(message.ReceiveId)
		if ret != 0 {
			zlog.Error("query group members failed: " + msg)
		} else {
			for _, member := range members {
				if canReceiveGroupMessage(member.Status) {
					targets[member.UserId] = struct{}{}
				}
			}
		}
	} else {
		targets[message.ReceiveId] = struct{}{}
	}

	targets[message.SendId] = struct{}{}
	return targets
}

func canReceiveGroupMessage(status int8) bool {
	return status == contact_status_enum.NORMAL || status == contact_status_enum.SILENCE
}

func (s *chatServer) sendToClient(clientId string, message wsMessage) {
	cli := s.getClient(clientId)
	if cli == nil {
		return
	}

	if err := cli.WriteJSON(message); err != nil {
		zlog.Error("send ws message failed: " + err.Error())
		staleClient := s.removeClient(clientId)
		if staleClient != nil {
			_ = staleClient.Close()
		}
	}
}
