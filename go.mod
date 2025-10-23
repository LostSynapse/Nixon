// go.mod
module nixon

go 1.21

require (
	github.com/go-gst/go-glib v1.4.0
	github.com/go-gst/go-gst v1.4.0
	github.com/gorilla/websocket v1.5.0
	github.com/natefinch/atomic v1.0.1 // Replaced hashicorp/go-atomicfile
	// github.com/jinzhu/gorm v1.9.16 // Should be removed by 'go mod tidy' if unused
	github.com/mattn/go-sqlite3 v1.14.22
	golang.org/x/sys v0.24.0 // Version might be updated by tidy
	gorm.io/driver/sqlite v1.5.6
	gorm.io/gorm v1.25.11
)

require (
	// Corrected timestamp verified
	github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.15.5 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.5 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-pointer v0.0.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	// github.com/jinzhu/gorm v1.9.16 // indirect dependency might still exist
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	golang.org/x/arch v0.5.0 // indirect
	golang.org/x/crypto v0.25.0 // indirect
	golang.org/x/exp v0.0.0-20240719175910-8a7402abbf56 // indirect
	golang.org/x/net v0.27.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Gin dependency removed as net/http is used

