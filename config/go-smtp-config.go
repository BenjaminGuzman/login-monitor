package config

type GoSMTPConfig struct {
	Identity string `json:"identity"`
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
}
