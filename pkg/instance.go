package pkg

// attributes
const (
	AttrInstanceType   = "instance_type"
	AttrInstanceState  = "instance_state"
	AttrKeyName        = "key_name"
	AttrTags           = "tags"
	AttrSecurityGroups = "security_groups"
	AttrPublicIP       = "public_ip"
)

type Instance struct {
	ID             string
	Type           string
	State          string
	KeyName        string
	Tags           map[string]string
	SecurityGroups []string
	PublicIP       string
}

type InstanceMap = map[string]Instance
