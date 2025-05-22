package Router

import "strings"

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
	for pattern, handler := range r.handlers {
		if matchTopic(pattern, topic) {
			handler(payload)
			return
		}
	}
}
func matchTopic(pattern, topic string) bool {
	pp := strings.Split(pattern, "/")
	tp := strings.Split(topic, "/")

	if len(pp) != len(tp) {
		return false
	}
	for i := range pp {
		if pp[i] == "+" {
			continue
		}
		if pp[i] != tp[i] {
			return false
		}
	}
	return true
}
