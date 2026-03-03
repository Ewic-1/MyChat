package gorm

import (
	"mychat_server/internal/dao"
	"mychat_server/internal/dto/request"
	"mychat_server/internal/dto/respond"
	"mychat_server/internal/model"
	"mychat_server/internal/service/sms"
	"mychat_server/pkg/utils/jwtutil"
	"mychat_server/pkg/utils/passwordutil"
	"mychat_server/pkg/utils/zlog"
	"regexp"
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
		zlog.Error("wrong password")
		return "wrong password", nil, -2
	}

	// 登录成功后签发 access + refresh，前端后续通过 refresh 自动续期。
	tokenPair, err := jwtutil.GenerateTokenPair(user.Uuid)
	if err != nil {
		zlog.Error(err.Error())
		return "generate token failed", nil, -1
	}

	data = &respond.LoginRespond{
		Uuid:      user.Uuid,
		Telephone: user.Telephone,
		Nickname:  user.Nickname,
		Email:     user.Email,
		Avatar:    user.Avatar,
		Gender:    user.Gender,
		Birthday:  user.Birthday,
		Signature: user.Signature,
		IsAdmin:   user.IsAdmin,
		Status:    user.Status,
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
