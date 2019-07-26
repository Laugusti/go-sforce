package credentials

// OAuth is the Salesforce OAuth credentials.
type OAuth struct {
	Username     string
	Password     string
	ClientID     string
	ClientSecret string
}

// New returns a pointer to a new OAuth credential.
func New(username, password, clientID, clientSecret string) *OAuth {
	return &OAuth{username, password, clientID, clientSecret}
}
