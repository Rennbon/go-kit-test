package config

import "time"

func DecodeConfig(path string) (c *Config, err error) {
	provider, err := FromConfigString(path, "toml")
	if err != nil {
		return nil, err
	}
	c = new(Config)
	err = provider.Unmarshal(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

type Config struct {
	Server *ServerConfig `json:"server"`
	Consul *ConsulConfig `json:"consul"`
}
type ServerConfig struct {
	Local          bool          `json:"local"`          //是否本机
	Host           string        `json:"host"`           // 127.0.0.1
	Port           string        `json:"port"`           // grpc 服务端口
	MetricsPort    string        `json:"metricsport"`    // http metrics端口，供prometheus拉取用
	Name           string        `json:"name"`           //service name
	HealthInterval time.Duration `json:"healthInterval"` //consul 健康检查周期
	Deregister     time.Duration `json:"deregister"`     //consul 注销时间
	ZipkinReporter string        `json:"zipkinReporter"` //http://XXX:9411/api/v2/spans
	CertFile       string        `json:"certFile"`
	KeyFile        string        `json:"keyFile"`
}

type ConsulConfig struct {
	Address    string        `json:"address"`    //default:127.0.0.1:8500
	Scheme     string        `json:"scheme"`     //default:http
	Datacenter string        `json:"datacenter"` //数据中心
	User       string        `json:"user"`       //用户名
	Password   string        `json:"password"`   //密码  for Auth
	WaitTime   time.Duration `json:"waittime"`   //请求等待时间
	TLSconfig  *TLSConfig    `json:"tls"`        //tls配置
}

type TLSConfig struct {
	// Address is the optional address of the Consul server. The port, if any
	// will be removed from here and this will be set to the ServerName of the
	// resulting config.
	Address string `json:"address"` //consul.test

	// CAFile is the optional path to the CA certificate used for Consul
	// communication, defaults to the system bundle if not specified.
	CAFile string `json:"cafile"` //ca.pem

	// CAPath is the optional path to a directory of CA certificates to use for
	// Consul communication, defaults to the system bundle if not specified.
	CAPath string `json:"capath"` //cert/

	// CertFile is the optional path to the certificate for Consul
	// communication. If this is set then you need to also set KeyFile.
	CertFile string `json:"certfile"` //client.crt

	// KeyFile is the optional path to the private key for Consul communication.
	// If this is set then you need to also set CertFile.
	KeyFile string `json:"keyfile"` //client.key

	//是否启用
	Enable bool `json:"enable"`
}
