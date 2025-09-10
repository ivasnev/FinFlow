module github.com/ivasnev/FinFlow/ff-id

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/ivasnev/FinFlow/ff-tvm v0.0.0
	github.com/redis/go-redis/v9 v9.4.0
	github.com/spf13/viper v1.18.2
	golang.org/x/crypto v0.19.0
	gorm.io/driver/postgres v1.5.6
	gorm.io/gorm v1.25.7
)

replace github.com/ivasnev/FinFlow/ff-tvm => ../ff-tvm 