package options

type RedisOptions struct {
	Address  string
	Password string
}

var (
	DefaultCacheOptions RedisOptions = RedisOptions{
		Address:  "10.8.9.100",
		Password: "6379",
	}
)
