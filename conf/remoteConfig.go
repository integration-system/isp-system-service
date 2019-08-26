package conf

import (
	"github.com/integration-system/isp-lib/structure"
)

type RemoteConfig struct {
	DB                     structure.DBConfiguration      `schema:"Настройка базы данных"`
	RedisAddress           structure.AddressConfiguration `schema:"Настройка Redis"`
	DefaultTokenExpireTime int64                          `schema:"Время жизни токена по умолчанию,время жизни токена в миллисекундах с момента его создания. если = -1 - время жизни неограниченно"`
	Metrics                structure.MetricConfiguration  `schema:"Настройка метрик"`
}
