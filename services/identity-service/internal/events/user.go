package events

type UserCreated struct {
	ID        int64  `json:"id,omitempty"`
	Email     string `json:"email,omitempty"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Phone     string `json:"phone,omitempty"`
}

func (u UserCreated) Name() string {
	return "UserCreated"
}

type UserUpdated struct {
	ID        int64  `json:"id,omitempty"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Phone     string `json:"phone,omitempty"`
}

func (u UserUpdated) Name() string {
	return "UserUpdated"
}
