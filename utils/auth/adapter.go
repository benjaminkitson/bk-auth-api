package auth

type AdapterHandler func(map[string]string) (map[string]string, error)

type EmailVerifier interface {
	VerifyEmail(map[string]string) (map[string]string, error)
}
