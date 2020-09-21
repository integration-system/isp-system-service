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
	CreatedAt   time.Time
	UpdatedAt   time.Time
	TableName   string `sql:"system_service.system" json:"-"`
	Id          int32
	Name        string `valid:"required~Required"`
	Description string
}

type Domain struct {
	CreatedAt   time.Time
	UpdatedAt   time.Time
	TableName   string `sql:"system_service.domain" json:"-"`
	Id          int32
	SystemId    int32
	Name        string `valid:"required~Required"`
	Description string
}

type Service struct {
	CreatedAt   time.Time
	UpdatedAt   time.Time
	TableName   string `sql:"system_service.service" json:"-"`
	Id          int32
	DomainId    int32  `valid:"required~Required"`
	Name        string `valid:"required~Required"`
	Description string
}

type Application struct {
	CreatedAt   time.Time
	UpdatedAt   time.Time
	TableName   string `sql:"system_service.application" json:"-"`
	Name        string `valid:"required~Required"`
	Description string
	Type        string `valid:"required~Required,in(SYSTEM|MOBILE)"`
	ServiceId   int32  `valid:"required~Required"`
	Id          int32
}

type Token struct {
	TableName  string `sql:"system_service.token" json:"-"`
	Token      string `valid:"required~Required" sql:"pk:token"`
	AppId      int32  `valid:"required~Required"`
	ExpireTime int64
	CreatedAt  time.Time
}

type AccessList struct {
	TableName string `sql:"system_service.access_list" json:"-"`
	Method    string `sql:",pk"`
	AppId     int32  `sql:",pk"`
	Value     bool
}
