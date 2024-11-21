package config

import "os"

type ContextKey string

const (
	// Context keys
	UserIDKey         ContextKey = "userID"
	CartIDKey         ContextKey = "cartID"
	SessionIDKey      ContextKey = "sessionID"
	TempSessionIDKey  ContextKey = "temporarySessionID"
	CartContentKey    ContextKey = "cartContent"
	ProductDetailsKey ContextKey = "productDetails"

	// Staff discount
	StaffDiscountRate = 0.25 // 25% discount for staff members
	StaffMonthlyQuota = 3    // Maximum number of items staff can buy at a discount per month

	// Member discount tiers
	MemberDiscountTier1Min = 3000  // £30.00
	MemberDiscountTier2Min = 6000  // £60.00
	MemberDiscountTier3Min = 9000  // £90.00
	MemberDiscountTier4Min = 12000 // £120.00

	MemberDiscountTier1Rate = 0.05 // 5% discount
	MemberDiscountTier2Rate = 0.10 // 10% discount
	MemberDiscountTier3Rate = 0.15 // 15% discount
	MemberDiscountTier4Rate = 0.20 // 20% discount
)

var (
	RedisAddress  = os.Getenv("FM_REDIS_ADDRESS")
	RedisPassword = os.Getenv("FM_REDIS_PASSWORD")
	DBHost        = os.Getenv("FM_DB_HOST")
	DBUsername    = os.Getenv("FM_DB_USERNAME")
	DBName        = os.Getenv("FM_DB_NAME")
	DBPort        = os.Getenv("FM_DB_PORT")
	DBSslMode     = os.Getenv("FM_DB_SSL_MODE") // "require" or "disable"
	APIPort       = os.Getenv("FM_ADMIN_API_PORT")
	Pepper        = os.Getenv("FM_USER_PASSWORD_PEPPER")

	SessionIDSecretKey = os.Getenv("FM_SESSION_ID_SECRET_KEY")

	RedisCachePrefixCart    = os.Getenv("FM_REDIS_CACHE_CART_PREFIX")
	RedisCachePrefixProduct = os.Getenv("FM_REDIS_CACHE_PRODUCT_PREFIX")

	HighPriorityEmailQueue = "high_priority_email_queue"
	LowPriorityEmailQueue  = "low_priority_email_queue"

	ImageDestProtocol = "file://"
	ImageDestDir      = "/Users/cindyho/fairymade/system_v2/test_images"
)

// GetMemberDiscountRate returns the appropriate discount rate based on the accumulated value
func GetMemberDiscountRate(accumulatedValue int) float64 {
	switch {
	case accumulatedValue >= MemberDiscountTier4Min:
		return MemberDiscountTier4Rate
	case accumulatedValue >= MemberDiscountTier3Min:
		return MemberDiscountTier3Rate
	case accumulatedValue >= MemberDiscountTier2Min:
		return MemberDiscountTier2Rate
	case accumulatedValue >= MemberDiscountTier1Min:
		return MemberDiscountTier1Rate
	default:
		return 0
	}
}
