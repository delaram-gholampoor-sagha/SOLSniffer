package utils

type Environment string

const (
	Testing       Environment = "testing"
	Production    Environment = "production"
	Canary        Environment = "canary"
	Staging       Environment = "staging"
	StagingCanary Environment = "staging-canary"
	Local         Environment = "local"
)

func (e Environment) IsTesting() bool {
	return e == Testing
}

func (e Environment) IsProduction() bool {
	return e == Production
}

func (e Environment) IsStaging() bool {
	return e == Staging
}

func (e Environment) IsLocal() bool {
	return e == Local
}

func (e Environment) IsCanary() bool {
	return e == Canary
}

func (e Environment) IsStagingCanary() bool {
	return e == Canary
}
