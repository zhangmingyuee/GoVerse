package jwt

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
	"time"
)

//const TokenExpireDuration = time.Hour * 2

// CustomSecret 用于加盐的字符串
var CustomSecret = []byte("夏天夏天悄悄过去")

// CustomClaims 自定义声明类型 并内嵌jwt.RegisteredClaims
// jwt包自带的jwt.RegisteredClaims只包含了官方字段
// 假设我们这里需要额外记录一个username字段，所以要自定义结构体
// 如果想要保存更多信息，都可以添加到这个结构体中
type CustomClaims struct {
	// 可根据需要自行添加字段
	Userid               int64  `json:"uid"`
	Username             string `json:"uname"`
	jwt.RegisteredClaims        // 内嵌标准的声明
}

// GenToken 生成JWT
func GenToken(userid int64, username string) (string, error) {
	// 创建一个我们自己的声明
	claims := CustomClaims{
		userid,
		username, // 自定义字段
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(viper.GetInt("auth.jwt_expire")) * time.Hour)), // 设置过期时间
			Issuer:    "bluebell",                                                                                     // 签发人
		},
	}
	// 使用 jwt.NewWithClaims 创建 JWT 令牌，指定 签名算法 为 HS256（HMAC-SHA256），并携带 claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 使用指定的secret签名并获得完整的编码后的字符串token
	return token.SignedString(CustomSecret)
}

// ParseToken 解析JWT
func ParseToken(tokenString string) (*CustomClaims, error) {
	claims := new(CustomClaims)
	// 解析token
	// 如果是自定义Claim结构体则需要使用 ParseWithClaims 方法
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (i interface{}, err error) {
		// 直接使用标准的Claim则可以直接使用Parse方法
		//token, err := jwt.Parse(tokenString, func(token *jwt.Token) (i interface{}, err error) {
		return CustomSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if token.Valid { // 校验token
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
