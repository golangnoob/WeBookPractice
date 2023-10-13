package domain

type SMSRetry struct {
	Id           int64
	Biz          string
	Args         []string
	PhoneNumbers []string
}
