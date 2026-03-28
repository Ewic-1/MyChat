package jwtutil

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	myconfig "mychat_server/internal/config"
	redisservice "mychat_server/internal/service/myredis"
	"mychat_server/pkg/utils/zlog"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// tokenTypeAccess / tokenTypeRefresh 用于区分 token 用途，防止混用。
	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"

	defaultActiveKID  = "v1"
	defaultAccessTTL  = 15 * time.Minute
	defaultRefreshTTL = 7 * 24 * time.Hour

	refreshTokenRedisKeyPrefix = "jwt_refresh_"
)

type jwtSettings struct {
	activeKID string

	accessSigningSecret  []byte
	refreshSigningSecret []byte

	accessVerifySecrets  [][]byte
	refreshVerifySecrets [][]byte

	accessTTL  time.Duration
	refreshTTL time.Duration
}

var (
	// 配置只在首次使用时加载一次，避免每次签发/验签都读取配置。
	settingsOnce sync.Once
	settings     jwtSettings
)

// TokenPair 统一返回 access/refresh 两类 token 及其过期时间。
type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

type CustomClaims struct {
	Uuid      string `json:"uuid"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// GenerateToken 兼容旧调用方，仅返回 access token。
func GenerateToken(uuid string) (string, error) {
	current := getSettings()
	accessExpiresAt := time.Now().Add(current.accessTTL)
	return generateSignedToken(uuid, tokenTypeAccess, accessExpiresAt, current.activeKID, current.accessSigningSecret)
}

// GenerateTokenPair 登录后签发 access + refresh，并记录 refresh token。
func GenerateTokenPair(uuid string) (*TokenPair, error) {
	current := getSettings()
	now := time.Now()
	accessExpiresAt := now.Add(current.accessTTL)
	refreshExpiresAt := now.Add(current.refreshTTL)

	accessToken, err := generateSignedToken(uuid, tokenTypeAccess, accessExpiresAt, current.activeKID, current.accessSigningSecret)
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateSignedToken(uuid, tokenTypeRefresh, refreshExpiresAt, current.activeKID, current.refreshSigningSecret)
	if err != nil {
		return nil, err
	}

	if err = storeRefreshToken(refreshToken, uuid, refreshExpiresAt); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessExpiresAt:  accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

// ParseToken 仅用于 access token 验签（给鉴权中间件使用）。
func ParseToken(tokenString string) (*CustomClaims, error) {
	current := getSettings()
	return parseBySecrets(tokenString, current.accessVerifySecrets, tokenTypeAccess)
}

// RotateTokenPair 使用 refresh token 轮换生成新 token 对（一次性消费旧 refresh token）。
func RotateTokenPair(refreshToken string) (*TokenPair, error) {
	current := getSettings()

	claims, err := parseBySecrets(refreshToken, current.refreshVerifySecrets, tokenTypeRefresh)
	if err != nil {
		return nil, err
	}

	if err = consumeRefreshToken(refreshToken, claims.Uuid); err != nil {
		return nil, err
	}

	return GenerateTokenPair(claims.Uuid)
}

// RevokeRefreshTokensByUUID 在用户退出时撤销该用户所有 refresh token。
func RevokeRefreshTokensByUUID(uuid string) {
	prefix := refreshTokenRedisKeyPrefix + uuid + "_"
	if err := redisservice.DelKeysWithPrefix(prefix); err != nil {
		zlog.Error(err.Error())
	}
}

func generateSignedToken(uuid, tokenType string, expiresAt time.Time, activeKID string, key []byte) (string, error) {
	claims := CustomClaims{
		Uuid:      uuid,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = activeKID
	return token.SignedString(key)
}

// parseBySecrets 会按密钥列表依次验签，用于密钥轮换期间兼容旧 token。
func parseBySecrets(tokenString string, verifyKeys [][]byte, expectedTokenType string) (*CustomClaims, error) {
	var lastErr error
	for _, secret := range verifyKeys {
		claims := &CustomClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if token.Method == nil || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, errors.New("unexpected signing method")
			}
			return secret, nil
		})
		if err != nil {
			lastErr = err
			continue
		}
		if !token.Valid {
			lastErr = errors.New("invalid token")
			continue
		}
		if claims.TokenType != expectedTokenType {
			return nil, errors.New("invalid token type")
		}
		return claims, nil
	}

	if lastErr == nil {
		lastErr = errors.New("invalid token")
	}
	return nil, lastErr
}

func getSettings() jwtSettings {
	settingsOnce.Do(loadSettings)
	return settings
}

func loadSettings() {
	cfg := myconfig.GetConfig()

	activeKID := firstNonEmpty(os.Getenv("JWT_ACTIVE_KID"), cfg.JWTConfig.ActiveKid, defaultActiveKID)

	accessSecret := firstNonEmpty(os.Getenv("JWT_ACCESS_SECRET"), cfg.JWTConfig.AccessSecret)
	refreshSecret := firstNonEmpty(os.Getenv("JWT_REFRESH_SECRET"), cfg.JWTConfig.RefreshSecret)
	// 开发环境若未配置密钥，生成进程级临时密钥，避免空密钥签名。
	if accessSecret == "" {
		accessSecret = createEphemeralSecret()
	}
	if refreshSecret == "" {
		refreshSecret = createEphemeralSecret()
	}

	accessPrevious := mergeSecretList(parseEnvSecretList("JWT_ACCESS_PREVIOUS_SECRETS"), cfg.JWTConfig.PreviousAccessSecrets)
	refreshPrevious := mergeSecretList(parseEnvSecretList("JWT_REFRESH_PREVIOUS_SECRETS"), cfg.JWTConfig.PreviousRefreshSecrets)

	accessVerifySecrets := []string{accessSecret}
	accessVerifySecrets = append(accessVerifySecrets, accessPrevious...)

	refreshVerifySecrets := []string{refreshSecret}
	refreshVerifySecrets = append(refreshVerifySecrets, refreshPrevious...)

	accessTTL := parseDurationSeconds(
		firstNonEmpty(os.Getenv("JWT_ACCESS_TTL_SECONDS"), strconv.FormatInt(cfg.JWTConfig.AccessTTLSeconds, 10)),
		defaultAccessTTL,
	)
	refreshTTL := parseDurationSeconds(
		firstNonEmpty(os.Getenv("JWT_REFRESH_TTL_SECONDS"), strconv.FormatInt(cfg.JWTConfig.RefreshTTLSeconds, 10)),
		defaultRefreshTTL,
	)

	settings = jwtSettings{
		activeKID: activeKID,

		accessSigningSecret:  []byte(accessSecret),
		refreshSigningSecret: []byte(refreshSecret),

		accessVerifySecrets:  secretStringsToBytes(accessVerifySecrets),
		refreshVerifySecrets: secretStringsToBytes(refreshVerifySecrets),

		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func createEphemeralSecret() string {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return hex.EncodeToString(raw)
}

func parseDurationSeconds(raw string, defaultValue time.Duration) time.Duration {
	seconds, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || seconds <= 0 {
		return defaultValue
	}
	return time.Duration(seconds) * time.Second
}

func parseEnvSecretList(envKey string) []string {
	raw := strings.TrimSpace(os.Getenv(envKey))
	if raw == "" {
		return nil
	}
	items := strings.Split(raw, ",")
	result := make([]string, 0, len(items))
	for _, item := range items {
		secret := strings.TrimSpace(item)
		if secret != "" {
			result = append(result, secret)
		}
	}
	return result
}

// mergeSecretList 会去重并过滤空值，保证验签密钥列表稳定。
func mergeSecretList(primary []string, secondary []string) []string {
	result := make([]string, 0, len(primary)+len(secondary))
	seen := make(map[string]struct{})
	for _, candidate := range append(primary, secondary...) {
		trimmed := strings.TrimSpace(candidate)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func secretStringsToBytes(secrets []string) [][]byte {
	result := make([][]byte, 0, len(secrets))
	for _, secret := range secrets {
		trimmed := strings.TrimSpace(secret)
		if trimmed == "" {
			continue
		}
		result = append(result, []byte(trimmed))
	}
	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func storeRefreshToken(token, uuid string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return errors.New("refresh token already expired")
	}

	// refresh token 只存哈希索引到 Redis，依赖 TTL 自动过期。
	key := refreshTokenRedisKey(uuid, token)
	return redisservice.SetKeyEx(key, "1", ttl)
}

// consumeRefreshToken 负责 refresh token 一次性消费语义。
func consumeRefreshToken(token, uuid string) error {
	// 通过 Lua 脚本在 Redis 内部原子执行“存在检查+删除”，避免并发竞态。
	key := refreshTokenRedisKey(uuid, token)
	consumed, err := redisservice.ConsumeKeyOnceAtomic(key)
	if err != nil {
		return err
	}
	if !consumed {
		return errors.New("refresh token not found")
	}

	return nil
}

func refreshTokenRedisKey(uuid, token string) string {
	return refreshTokenRedisKeyPrefix + uuid + "_" + tokenHash(token)
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
