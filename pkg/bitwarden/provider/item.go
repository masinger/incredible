package provider

type Item struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Login *Login `json:"login,omitempty"`
}

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
