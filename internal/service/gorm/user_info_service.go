package gorm

import (
	"fmt"
	"mychat_server/internal/dao"
	"mychat_server/internal/dto/request"
	"mychat_server/internal/dto/respond"
	"mychat_server/internal/model"
	"mychat_server/internal/service/redis"
	"mychat_server/internal/service/sms"
	"mychat_server/pkg/constants"
	"mychat_server/pkg/enum/user_info/user_status_enum"
	"mychat_server/pkg/utils/jwtutil"
	"mychat_server/pkg/utils/passwordutil"
	"mychat_server/pkg/utils/random"
	"mychat_server/pkg/utils/zlog"
	"regexp"
	"time"
)

type UserInfoService struct{}

var userInfoDao = new(dao.UserInfoDao)

func (u *UserInfoService) checkTelephoneValid(telephone string) bool {
	pattern := `^1([38][0-9]|14[579]|5[^4]|16[6]|7[1-35-8]|9[189])\d{8}$`
	match, err := regexp.MatchString(pattern, telephone)
	if err != nil {
		zlog.Error(err.Error())
	}
	return match
}

func (u *UserInfoService) checkEmailValid(email string) bool {
	pattern := `^[^\s@]+@[^\s@]+\.[^\s@]+$`
	match, err := regexp.MatchString(pattern, email)
	if err != nil {
		zlog.Error(err.Error())
	}
	return match
}

func (u *UserInfoService) checkUserIsAdminOrNot(user model.UserInfo) int8 {
	return user.IsAdmin
}

func (u *UserInfoService) Login(loginReq request.LoginRequest) (message string, data *respond.LoginRespond, ret int) {
	msg, user, code := userInfoDao.GetUserInfoByTelephone(loginReq.Telephone)
	if code != 0 {
		return msg, nil, code
	}

	if !passwordutil.CheckPassword(user.Password, loginReq.Password) {
		zlog.Error("密码错误")
		return "密码错误", nil, -2
	}

	// 登录成功后签发 access + refresh，前端后续通过 refresh 自动续期。
	tokenPair, err := jwtutil.GenerateTokenPair(user.Uuid)
	if err != nil {
		zlog.Error(err.Error())
		return "获取token失败", nil, -1
	}

	data = &respond.LoginRespond{
		Uuid:             user.Uuid,
		Telephone:        user.Telephone,
		Nickname:         user.Nickname,
		Email:            user.Email,
		Avatar:           user.Avatar,
		Gender:           user.Gender,
		Birthday:         user.Birthday,
		Signature:        user.Signature,
		IsAdmin:          user.IsAdmin,
		Status:           user.Status,
		Token:            tokenPair.AccessToken,
		RefreshToken:     tokenPair.RefreshToken,
		AccessExpiresAt:  tokenPair.AccessExpiresAt.Unix(),
		RefreshExpiresAt: tokenPair.RefreshExpiresAt.Unix(),
	}

	return "login success", data, 0
}

func (u *UserInfoService) SendSmsCode(telephone string) (string, int) {
	return sms.VerifyCode(telephone)
}

func (u *UserInfoService) Register(req request.RegisterRequest) (string, *respond.RegisterRespond, int) {
	var msg string
	// 查看手机号是否存在
	if msg, exist, ret := userInfoDao.ExistsByTelephone(req.Telephone); ret != -1 {
		if exist {
			zlog.Error(msg)
			return msg, nil, -2
		}
	} else {
		zlog.Error(msg)
		return msg, nil, -1
	}

	// redis中获取验证码进行比对
	codeFromRedis, err := redis.GetKey("auth_code_" + req.Telephone)
	codeFromUser := req.SmsCode
	if err != nil {
		zlog.Error(err.Error())
		msg = "未发送验证码"
		return msg, nil, -1
	}
	if codeFromUser != codeFromRedis {
		msg = "验证码错误"
		zlog.Info(msg)
		return msg, nil, -2
	} else {
		// 删除redis中的验证码
		if err := redis.DelKeyIfExists("auth_code_" + req.Telephone); err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, nil, -1
		}
	}

	// 创建新用户
	var newUser model.UserInfo
	newUser.Telephone = req.Telephone
	newUser.Uuid = "U" + random.GetNowAndLenRandomString(11)
	newUser.Nickname = req.Nickname
	newUser.Avatar = ""
	hashedPassword, err := passwordutil.HashPassword(req.Password)
	if err != nil {
		zlog.Error(err.Error())
		return constants.SYSTEM_ERROR, nil, -1
	}
	newUser.Password = hashedPassword

	newUser.CreatedAt = time.Now()
	newUser.IsAdmin = u.checkUserIsAdminOrNot(newUser)
	newUser.Status = user_status_enum.NORMAL

	m, _ := userInfoDao.NewUser(&newUser)
	if m == constants.SYSTEM_ERROR {
		return constants.SYSTEM_ERROR, nil, -1
	}
	newUser.LastOnlineAt.Time = time.Now()
	rep := &respond.RegisterRespond{
		Uuid:      newUser.Uuid,
		Telephone: newUser.Telephone,
		Nickname:  newUser.Nickname,
		Email:     newUser.Email,
		Avatar:    newUser.Avatar,
		Gender:    newUser.Gender,
		Birthday:  newUser.Birthday,
		Signature: newUser.Signature,
		IsAdmin:   newUser.IsAdmin,
		Status:    newUser.Status,
	}
	year, month, day := newUser.CreatedAt.Date()
	rep.CreatedAt = fmt.Sprintf("%d.%d.%d", year, month, day)

	return m, rep, 0
}

func (u *UserInfoService) SmsLogin(req request.SmsLoginRequest) (msg string, rep *respond.LoginRespond, ret int) {
	// redis中获取验证码进行比对
	codeFromRedis, err := redis.GetKey("auth_code_" + req.Telephone)
	codeFromUser := req.SmsCode
	if err != nil {
		zlog.Error(err.Error())
		msg = "未发送验证码"
		return msg, nil, -1
	}
	if codeFromUser != codeFromRedis {
		msg = "验证码错误"
		zlog.Info(msg)
		return msg, nil, -2
	} else {
		// 删除redis中的验证码
		if err := redis.DelKeyIfExists("auth_code_" + req.Telephone); err != nil {
			zlog.Error(err.Error())
			return constants.SYSTEM_ERROR, nil, -1
		}
	}
	// 查询用户
	msg, user, ret := userInfoDao.GetUserInfoByTelephone(req.Telephone)
	if ret != 0 {
		zlog.Error(msg)
		return constants.SYSTEM_ERROR, nil, ret
	}

	// 登录成功后签发 access + refresh，前端后续通过 refresh 自动续期。
	tokenPair, err := jwtutil.GenerateTokenPair(user.Uuid)
	if err != nil {
		zlog.Error(err.Error())
		return "获取token失败", nil, -1
	}

	var data = &respond.LoginRespond{
		Uuid:             user.Uuid,
		Telephone:        user.Telephone,
		Nickname:         user.Nickname,
		Email:            user.Email,
		Avatar:           user.Avatar,
		Gender:           user.Gender,
		Birthday:         user.Birthday,
		Signature:        user.Signature,
		IsAdmin:          user.IsAdmin,
		Status:           user.Status,
		Token:            tokenPair.AccessToken,
		RefreshToken:     tokenPair.RefreshToken,
		AccessExpiresAt:  tokenPair.AccessExpiresAt.Unix(),
		RefreshExpiresAt: tokenPair.RefreshExpiresAt.Unix(),
	}

	return "login success", data, 0
}
