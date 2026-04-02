package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"mychat_server/internal/config"
	"mychat_server/internal/dto/request"
	"mychat_server/internal/model"
	mykafka "mychat_server/internal/service/kafka"
	"mychat_server/pkg/enum/contact/contact_status_enum"
	"mychat_server/pkg/utils/random"
	"mychat_server/pkg/utils/zlog"
)

func InitRuntime() {
	ChatServer.InitRuntime()
}

func StopRuntime() {
	ChatServer.StopRuntime()
}

func (s *chatServer) InitRuntime() {
	s.runtimeOnce.Do(func() {
		mode := strings.ToLower(strings.TrimSpace(config.GetConfig().KafkaConfig.MessageMode))
		s.kafkaEnabled = mode == "kafka"
		if !s.kafkaEnabled {
			zlog.Info("chat runtime started with channel mode")
			return
		}

		mykafka.KafkaService.KafkaInit()
		go s.consumeKafkaMessages()
		zlog.Info("chat runtime started with kafka mode")
	})
}

func (s *chatServer) StopRuntime() {
	if s.kafkaEnabled {
		mykafka.KafkaService.KafkaClose()
	}
}

func (s *chatServer) consumeKafkaMessages() {
	for {
		msg, err := mykafka.KafkaService.ReadChatMessage(context.Background())
		if err != nil {
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
	if s.kafkaEnabled {
		payload, err := json.Marshal(out)
		if err != nil {
			zlog.Error("marshal kafka message failed: " + err.Error())
			s.dispatchMessage(out)
			return
		}
		if err := mykafka.KafkaService.PublishChatMessage(out.ReceiveId, payload); err != nil {
			zlog.Error("publish kafka message failed: " + err.Error())
			s.dispatchMessage(out)
			return
		}
		return
	}

	s.dispatchMessage(out)
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
