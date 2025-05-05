package Router

type HandlerFunc func([]byte)

type Router struct {
	handlers map[string]HandlerFunc
}

func NewRouter() *Router {
	return &Router{handlers: make(map[string]HandlerFunc)}
}

func (r *Router) Register(topic string, handler HandlerFunc) {
	r.handlers[topic] = handler
}

func (r *Router) Handle(topic string, payload []byte) {
	if handler, ok := r.handlers[topic]; ok {
		handler(payload)
	}
}