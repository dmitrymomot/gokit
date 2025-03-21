package google

type Config struct {
	ClientID     string   `env:"GOOGLE_OAUTH_CLIENT_ID,required"`
	ClientSecret string   `env:"GOOGLE_OAUTH_CLIENT_SECRET,required"`
	RedirectURL  string   `env:"GOOGLE_OAUTH_REDIRECT_URL,required"`
	Scopes       []string `env:"GOOGLE_OAUTH_SCOPES" envDefault:"openid,profile,email"`
	StateKey     string   `env:"GOOGLE_OAUTH_STATE_KEY" envDefault:"google_oauth_state"`
	VerifiedOnly bool     `env:"GOOGLE_OAUTH_VERIFIED_ONLY" envDefault:"true"`
}
