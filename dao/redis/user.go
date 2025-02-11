package redis

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

const otpGenerate = "otp"

// StoreOTP 存储用户的验证码到redis
func SetOTP(c *gin.Context, userID int64, otp string) (err error) {
	key := otpGenerate + ":" + strconv.FormatInt(userID, 10)
	fmt.Println("OPTkey: ", key)
	err = rdb.Set(c, key, otp, 5*time.Minute).Err()
	return
}

// GetOTP 从redis中拿到验证码
func GetOTP(c *gin.Context, userID int64) (otp string, err error) {
	key := otpGenerate + ":" + strconv.FormatInt(userID, 10)
	return rdb.Get(c, key).Result()
}
