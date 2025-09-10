module github.com/ivasnev/FinFlow/ff-files

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/google/uuid v1.6.0
	github.com/ivasnev/FinFlow/ff-tvm v0.0.0
	github.com/spf13/viper v1.18.2
	gorm.io/driver/postgres v1.5.6
	gorm.io/gorm v1.25.7
)

replace github.com/ivasnev/FinFlow/ff-tvm => ../ff-tvm 