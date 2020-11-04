package github

// ExchangeCodeResponse ...
type ExchangeCodeResponse struct {
	AccessToken string `json:"access_token"`
	Score       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

// UserProfile ...
type UserProfile struct {
	Login     string `json:"login"`
	ID        int    `json:"id"`
	AvatarURL string `json:"avatar_url"`
	URL       string `json:"url"`
	HTMLURL   string `json:"html_url"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	Company   string `json:"company"`
	Location  string `json:"location"`
	Email     string `json:"email"`
	Bio       string `json:"bio"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
