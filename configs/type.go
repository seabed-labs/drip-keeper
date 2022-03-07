package configs

const (
	PROJECT_DIR = "keeper-bot"
)

type Environment string

const (
	NilEnv   = Environment("")
	LocalEnv = Environment("LOCAL")
	TestEnv  = Environment("TEST")
	DevEnv   = Environment("DEV")
	ProdEnv  = Environment("PROD")
)

func IsLocal(env Environment) bool {
	return env == LocalEnv || env == NilEnv
}

func IsProd(env Environment) bool {
	return env == ProdEnv
}
