package pkg

type State map[string]any

// attributes
const (
	InstanceType   = "instance_type"
	InstanceState  = "instance_state"
	KeyName        = "key_name"
	Tags           = "tags"
	SecurityGroups = "security_groups"
	PublicIP       = "public_ip"
)

func SupportedAttributes() []string {
	return []string{InstanceType, InstanceState, KeyName, Tags, SecurityGroups, PublicIP}
}

type Instance struct {
	ID             string
	Type           string
	State          string
	KeyName        string
	Tags           map[string]string
	SecurityGroups []string
	PublicIP       string
}

func (i Instance) ToState() State {
	return State{
		InstanceType:   i.Type,
		InstanceState:  i.State,
		KeyName:        i.KeyName,
		Tags:           i.Tags,
		SecurityGroups: i.SecurityGroups,
		PublicIP:       i.PublicIP,
	}
}
