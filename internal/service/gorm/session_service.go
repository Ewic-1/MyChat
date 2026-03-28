package gorm

import (
	"encoding/json"
	"errors"
	"fmt"
	"mychat_server/internal/dto/request"
	"mychat_server/internal/dto/respond"
	"mychat_server/internal/model"
	"mychat_server/internal/service/myredis"
	"mychat_server/pkg/constants"
	"mychat_server/pkg/enum/contact/contact_status_enum"
	"mychat_server/pkg/enum/group_info/group_status_enum"
	"mychat_server/pkg/enum/user_info/user_status_enum"
	"mychat_server/pkg/utils/random"
	"mychat_server/pkg/utils/zlog"
	"time"

	"github.com/go-redis/redis/v8"
)

type SessionService struct{}

func (s *SessionService) CreateSession(req request.CreateSessionRequest) (string, string, int) {
	// 会话对象
	var session model.Session
	session.Uuid = fmt.Sprintf("S%s", random.GetNowAndLenRandomString(11))
	session.SendId = req.SendId
	session.ReceiveId = req.ReceiveId
	session.CreatedAt = time.Now()
	// 人/群两种情况
	if req.ReceiveId[0] == 'U' {
		msg, user, ret := userInfoDao.GetUserInfoByUuid(req.ReceiveId)
		if ret != 0 {
			zlog.Error(msg)
			return msg, session.Uuid, ret
		}
		if user.Status == user_status_enum.DISABLE {
			msg = "用户被禁用"
			zlog.Error(msg)
			return msg, session.Uuid, -2
		}
		session.Avatar = user.Avatar
		session.ReceiveName = user.Nickname
	} else {
		msg, group, ret := groupInfoDao.GetGroupInfoById(req.ReceiveId)
		if ret != 0 {
			zlog.Error(msg)
			return msg, session.Uuid, ret
		}
		if group.Status == group_status_enum.DISABLE {
			msg = "群聊被禁用"
			zlog.Error(msg)
			return msg, session.Uuid, -2
		}
		session.Avatar = group.Avatar
		session.ReceiveName = group.Name
	}
	// 落库
	msg, ret := sessionDao.SaveSession(session)
	if ret != 0 {
		zlog.Error(msg)
		return msg, session.Uuid, ret
	}
	// 删除缓存
	if err := myredis.DelKeysWithPattern("group_session_list_" + req.SendId); err != nil {
		zlog.Error(err.Error())
		return msg, session.Uuid, -1
	}
	if err := myredis.DelKeysWithPattern("session_list_" + req.ReceiveId); err != nil {
		zlog.Error(err.Error())
		return msg, session.Uuid, -1
	}
	// 返回（值为会话uuid）
	return "创建会话成功", session.Uuid, 0
}

// CheckOpenSessionAllowed 检查是否允许发起会话
func (s *SessionService) CheckOpenSessionAllowed(sendId, receiveId string) (string, bool, int) {
	msg, contact, ret := contactInfoDao.GetContactById(sendId, receiveId)
	if ret != 0 {
		zlog.Error(msg)
		return msg, false, ret
	}
	if contact.Status == contact_status_enum.BE_BLACK {
		return "已被对方拉黑，无法发起会话", false, -2
	} else if contact.Status == contact_status_enum.BLACK {
		return "已拉黑对方，先解除拉黑状态才能发起会话", false, -2
	}
	if receiveId[0] == 'U' {
		msg, user, ret := userInfoDao.GetUserInfoByUuid(receiveId)
		if ret != 0 {
			zlog.Error(msg)
			return msg, false, ret
		}
		if user.Status == user_status_enum.DISABLE {
			zlog.Info("对方已被禁用，无法发起会话")
			return "对方已被禁用，无法发起会话", false, -2
		}
	} else {
		msg, group, ret := groupInfoDao.GetGroupInfoById(receiveId)
		if ret != 0 {
			zlog.Error(msg)
			return msg, false, ret
		}
		if group.Status == group_status_enum.DISABLE {
			zlog.Info("对方已被禁用，无法发起会话")
			return "对方已被禁用，无法发起会话", false, -2
		}
	}
	return "可以发起会话", true, 0
}

func (s *SessionService) OpenSession(req request.OpenSessionRequest) (string, string, int) {
	// 查找缓存中是否存在要打开的会话
	rep, err := myredis.GetKeyWithPrefixNilIsErr("session_" + req.SendId + "_" + req.ReceiveId)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// 如果没查到，则到数据库查找
			msg, session, ret := sessionDao.GetSessionById(req.SendId, req.ReceiveId)
			if ret == -2 {
				zlog.Info(msg)
				// 还是没查到就新建
				createReq := request.CreateSessionRequest{
					SendId:    req.SendId,
					ReceiveId: req.ReceiveId,
				}
				return s.CreateSession(createReq)
			}
			return "会话创建成功", session.Uuid, 0
		} else {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, "", -1
		}
	}
	// 查找到的话反序列化
	var session model.Session
	if err := json.Unmarshal([]byte(rep), &session); err != nil {
		zlog.Error(err.Error())
		return constants.SYSTEM_ERROR, "", -1
	}
	// 返回创建的会话id
	return "会话创建成功", session.Uuid, 0
}

func (s *SessionService) GetUserSessionList(ownerId string) (string, []respond.UserSessionListRespond, int) {
	rsp, err := myredis.GetKeyNilIsErr("session_list_" + ownerId)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			msg, sessionList, ret := sessionDao.GetSessionBySendId(ownerId)
			if ret == -2 {
				zlog.Info(msg)
				return msg, nil, 0
			} else if ret == -1 {
				zlog.Error(msg)
				return constants.SYSTEM_ERROR, nil, -1
			}
			var sessionListRsp []respond.UserSessionListRespond
			for i := 0; i < len(sessionList); i++ {
				if sessionList[i].ReceiveId[0] == 'U' {
					sessionListRsp = append(sessionListRsp, respond.UserSessionListRespond{
						SessionId: sessionList[i].Uuid,
						Avatar:    sessionList[i].Avatar,
						UserId:    sessionList[i].ReceiveId,
						Username:  sessionList[i].ReceiveName,
					})
				}
			}
			rsp, err := json.Marshal(sessionListRsp)
			if err != nil {
				zlog.Error(err.Error())
			}
			if err := myredis.SetKeyEx("session_list_"+ownerId, string(rsp), time.Minute*constants.REDIS_TIMEOUT); err != nil {
				zlog.Error(err.Error())
			}
			return "获取成功", sessionListRsp, 0
		} else {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, nil, -1
		}
	}
	var r []respond.UserSessionListRespond
	if err := json.Unmarshal([]byte(rsp), &r); err != nil {
		zlog.Error(err.Error())
		return constants.SYSTEM_ERROR, nil, -1
	}
	return "获取成功", r, 0
}

func (s *SessionService) GetGroupSessionList(ownerId string) (string, []respond.GroupSessionListRespond, int) {
	rsp, err := myredis.GetKeyNilIsErr("group_session_list_" + ownerId)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			msg, sessionList, ret := sessionDao.GetSessionBySendId(ownerId)
			if ret == -2 {
				zlog.Info(msg)
				return msg, nil, 0
			} else if ret == -1 {
				zlog.Error(msg)
				return constants.SYSTEM_ERROR, nil, -1
			}
			var sessionListRsp []respond.GroupSessionListRespond
			for i := 0; i < len(sessionList); i++ {
				if sessionList[i].ReceiveId[0] == 'G' {
					sessionListRsp = append(sessionListRsp, respond.GroupSessionListRespond{
						SessionId: sessionList[i].Uuid,
						Avatar:    sessionList[i].Avatar,
						GroupId:   sessionList[i].ReceiveId,
						GroupName: sessionList[i].ReceiveName,
					})
				}
			}
			rsp, err := json.Marshal(sessionListRsp)
			if err != nil {
				zlog.Error(err.Error())
			}
			if err := myredis.SetKeyEx("group_session_list_"+ownerId, string(rsp), time.Minute*constants.REDIS_TIMEOUT); err != nil {
				zlog.Error(err.Error())
			}
			return "获取成功", sessionListRsp, 0
		} else {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, nil, -1
		}
	}
	var r []respond.GroupSessionListRespond
	if err := json.Unmarshal([]byte(rsp), &r); err != nil {
		zlog.Error(err.Error())
	}
	return "获取成功", r, 0
}

func (s *SessionService) DeleteSession(ownerId, sessionId string) (string, int) {
	msg, session, ret := sessionDao.GetSessionBySessionId(sessionId)
	if ret != 0 {
		zlog.Error(msg)
		return msg, -1
	}
	session.DeletedAt.Valid = true
	session.DeletedAt.Time = time.Now()
	msg, ret = sessionDao.SaveSession(session)
	if ret != 0 {
		zlog.Error(msg)
		return msg, -1
	}
	//if err := myredis.DelKeysWithSuffix(sessionId); err != nil {
	//	zlog.Error(err.Error())
	//}
	if err := myredis.DelKeysWithPattern("group_session_list_" + ownerId); err != nil {
		zlog.Error(err.Error())
	}
	if err := myredis.DelKeysWithPattern("session_list_" + ownerId); err != nil {
		zlog.Error(err.Error())
	}
	return "删除成功", 0
}
