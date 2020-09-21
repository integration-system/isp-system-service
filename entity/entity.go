package entity

import "time"

const (
	RootAdminApplicationId = 1
)

const (
	AppTypeSystem = "SYSTEM"
	AppTypeMobile = "MOBILE"
)

type System struct {
	tableName   string `pg:"system_service.system" json:"-"` //nolint
	Id          int32
	Name        string `valid:"required~Required"`
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Domain struct {
	tableName   string `pg:"system_service.domain" json:"-"` //nolint
	Name        string `valid:"required~Required"`
	Id          int32
	SystemId    int32
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Service struct {
	tableName   string `pg:"system_service.service" json:"-"` //nolint
	Id          int32
	DomainId    int32  `valid:"required~Required"`
	Name        string `valid:"required~Required"`
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Application struct {
	tableName   string `pg:"system_service.application" json:"-"` //nolint
	Name        string `valid:"required~Required"`
	Description string
	Type        string `valid:"required~Required,in(SYSTEM|MOBILE)"`
	ServiceId   int32  `valid:"required~Required"`
	Id          int32
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Token struct {
	tableName  string `pg:"system_service.token" json:"-"` //nolint
	Token      string `valid:"required~Required" pg:"pk:token"`
	AppId      int32  `valid:"required~Required"`
	ExpireTime int64
	CreatedAt  time.Time
}

type AccessList struct {
	tableName string `pg:"system_service.access_list" json:"-"` //nolint
	Method    string `pg:",pk"`
	AppId     int32  `pg:",pk"`
	Value     bool
}
