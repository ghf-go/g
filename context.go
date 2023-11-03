package g

type GContext struct {
	App GEngine
}
type WebHandlerFunc func(*GContext)
type SockHandlerFunc func(*GContext)
