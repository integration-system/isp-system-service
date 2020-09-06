package service

import (
	"encoding/hex"
	"math/rand"
	"time"

	"isp-system-service/conf"

	"github.com/dgrijalva/jwt-go"
	"github.com/integration-system/isp-lib/v2/config"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var Jwt jwtService

type jwtService struct{}

func (s jwtService) CreateApplication(appId int32, expTime int64) (string, error) {
	var (
		claims  = jwt.MapClaims{}
		created = time.Now()
		secret  = config.GetRemote().(*conf.RemoteConfig).ApplicationSecret
	)

	claims["appId"] = appId
	claims["iat"] = created.Unix()
	claims["salt"] = s.getSalt()
	if expTime > 0 {
		claims["exp"] = created.Add(time.Millisecond * time.Duration(expTime)).Unix()
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString([]byte(secret))
}

func (jwtService) getSalt() string {
	salt := make([]byte, rand.Intn(30)+10)
	rand.Read(salt)

	return hex.EncodeToString(salt)
}
