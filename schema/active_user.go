package schema

import "encoding/json"

// ActiveUser struct
type ActiveUser struct { //Payload
	Id       string `bson:"_id,omitempty" json:"_id,omitempty" redis:"_id"`
	Role     int    `bson:"role,omitempty" json:"role,omitempty" redis:"role"`
	Activity string `bson:"activity,omitempty" json:"activity,omitempty" redis:"activity"`
}

func (user ActiveUser) Database() string {
	return ""
}

func (user ActiveUser) Collection() string {
	return ""
}

func (user ActiveUser) Key() string {
	return user.Id
}

func (user ActiveUser) MarshalBinary() ([]byte, error) {
	return json.Marshal(user)
}
