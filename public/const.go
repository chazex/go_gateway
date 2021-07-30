package public

const (
	ValidatorKey        = "ValidatorKey"
	TranslatorKey       = "TranslatorKey"
	AdminSessionInfoKey = "AdminSessionInfoKey"

	LoadTypeHTTP = 0 // 负载类型http
	LoadTypeTCP  = 1 // 负载类型tcp
	LoadTypeGRPC = 2 // 负载类型grpc

	HTTPRuleTypePrefixURL = 0 // HTTP匹配类型， URL Path前缀匹配
	HTTPRuleTypeDomain    = 1 // HTTP匹配类型， 域名匹配

	RedisFlowDayKey  = "flow_day_count"
	RedisFlowHourKey = "flow_hour_count"

	FlowTotal          = "flow_total"
	FlowServicePrefix  = "flow_service_"
	FlowAppPrefix = "flow_app_"

	JwtSignKey = "my_sign_key"
	JwtExpires = 60*60
)

var (
	LoadTypeMap = map[int]string{
		LoadTypeHTTP: "HTTP",
		LoadTypeTCP:  "TCP",
		LoadTypeGRPC: "GRPC",
	}
)
