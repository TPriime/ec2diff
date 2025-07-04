package pkg

// attributes
const (
	InstanceType   = "instance_type"
	InstanceState  = "instance_state"
	KeyName        = "key_name"
	Tags           = "tags"
	SecurityGroups = "security_groups"
	PublicIP       = "public_ip"
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

type State map[string]any

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

func SupportedAttributes() []string {
	return []string{InstanceType, InstanceState, KeyName, Tags, SecurityGroups, PublicIP}
}
