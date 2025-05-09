package config

type GRPCServer struct {
	Address string `yaml:"address" env-default:":8081"`
}
