package middleware

import (
	"github.com/CloudDetail/node-agent/config"
	"github.com/CloudDetail/node-agent/utils"
)

type MiddlewareType int

const (
	RABBIT_MQ MiddlewareType = iota
	KAFKA
	ACTIVE_MQ
	ROCKET_MQ
	MYSQL
	POSTGRESQL
	MONGODB
	UNKNOWN
)

func GetMiddlewareType(port uint16) MiddlewareType {
	cfg := config.GlobalCfg.Middleware
	if utils.Contains(cfg.RabbitMQPort, port) {
		return RABBIT_MQ
	} else if utils.Contains(cfg.KafkaPort, port) {
		return KAFKA
	} else if utils.Contains(cfg.ActiveMQPort, port) {
		return ACTIVE_MQ
	} else if utils.Contains(cfg.RocketMQPort, port) {
		return ROCKET_MQ
	} else if utils.Contains(cfg.MySQLPort, port) {
		return MYSQL
	} else if utils.Contains(cfg.PostgreSQLPort, port) {
		return POSTGRESQL
	} else if utils.Contains(cfg.MongoDBPort, port) {
		return MONGODB
	}
	return UNKNOWN
}

func (m MiddlewareType) String() string {
	switch m {
	case RABBIT_MQ:
		return "rabbitmq"
	case KAFKA:
		return "kafka"
	case ACTIVE_MQ:
		return "activemq"
	case ROCKET_MQ:
		return "rocketmq"
	case MYSQL:
		return "mysql"
	case POSTGRESQL:
		return "postgresql"
	case MONGODB:
		return "mongodb"
	default:
		return "unknown"
	}
}
