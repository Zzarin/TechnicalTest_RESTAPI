package config

import "time"

type DBConfig struct {
	ConnectString   string        // строка подключения к БД
	Host            string        // host БД
	Port            string        // порт листенера БД
	Net             string        // протокол
	Dbname          string        // имя БД
	SslMode         string        // режим SSL
	User            string        // пользователь для подключения к БД
	Password        string        // пароль пользователя
	ConnMaxLifetime time.Duration // время жизни подключения в миллисекундах
	MaxOpenConns    int           // максимальное количество открытых подключений
	MaxIdleConns    int           // максимальное количество простаивающих подключений
	DriverName      string        // имя драйвера "mysqlDB"
}

/*
cfg := mysqlDB.Config{
        User:   os.Getenv("DBUSER"),
        Passwd: os.Getenv("DBPASS"),
        Net:    "tcp",
        Addr:   "127.0.0.1:3306",
        DBName: "recordings",
}
db, err = sql.Open("mysqlDB", cfg.FormatDSN())
*/

func GetDBConfig() *DBConfig {
	return &DBConfig{
		ConnectString:   "",
		Host:            "mysqlDB",
		Port:            "3306",
		Net:             "tcp",
		Dbname:          "exchange_rate",
		User:            "root",
		Password:        "pw",
		ConnMaxLifetime: time.Duration(3) * time.Minute,
		MaxOpenConns:    10,
		MaxIdleConns:    10,
		DriverName:      "mysqlDB",
	}
}

//root:mysqlpw@tcp(localhost:49168)/exchange_rate
func GetDBConfigLocal() *DBConfig {
	return &DBConfig{
		ConnectString:   "",
		Host:            "localhost",
		Port:            "3306",
		Net:             "tcp",
		Dbname:          "exchange_rate",
		User:            "root",
		Password:        "pw",
		ConnMaxLifetime: time.Duration(3) * time.Minute,
		MaxOpenConns:    10,
		MaxIdleConns:    10,
		DriverName:      "mysqlDB",
	}
}
