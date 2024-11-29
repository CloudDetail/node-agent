package config

type MiddlewarePortConfig struct {
	RabbitMQPort []uint16 `yaml:"rabbitmq"`
	KafkaPort    []uint16 `yaml:"kafka"`
	ActiveMQPort []uint16 `yaml:"activemq"`
	RocketMQPort []uint16 `yaml:"rocketmq"`

	MySQLPort      []uint16 `yaml:"mysql"`
	PostgreSQLPort []uint16 `yaml:"postgresql"`
	MongoDBPort    []uint16 `yaml:"mongodb"`
}

func (m *MiddlewarePortConfig) setDefault() {
	if len(m.RabbitMQPort) == 0 {
		m.RabbitMQPort = []uint16{5672}
	}
	if len(m.KafkaPort) == 0 {
		m.KafkaPort = []uint16{9092}
	}
	if len(m.ActiveMQPort) == 0 {
		m.ActiveMQPort = []uint16{61616}
	}
	if len(m.RocketMQPort) == 0 {
		m.RocketMQPort = []uint16{10911}
	}
	if len(m.MySQLPort) == 0 {
		m.MySQLPort = []uint16{3306}
	}
	if len(m.PostgreSQLPort) == 0 {
		m.PostgreSQLPort = []uint16{5432}
	}
	if len(m.MongoDBPort) == 0 {
		m.MongoDBPort = []uint16{27017}
	}
}
