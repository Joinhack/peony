package peony

type Session struct {
	Attribute map[string]interface{}
	Id        string
}

type SessionManager interface {
	GenerateId() string
	Store(c *Controller, s *Session) //save session
	Get(c *Controller) *Session      //get session
}

//Set attribute
func (s *Session) Set(key string, value interface{}) {
	s.Attribute[key] = value
}

//Get attribute
func (s *Session) Get(key string) (val interface{}, ok bool) {
	val, ok = s.Attribute[key]
	return
}

//Get session id
func (s *Session) GetId() string {
	return s.Id
}

//Set session id
func (s *Session) SetId(id string) {
	s.Id = id
}

func GetSessionFilter(s *Server) Filter {
	return func(c *Controller, filter []Filter) {
		if s.SessionManager == nil {
			filter[0](c, filter[1:])
			return
		}
		session := s.SessionManager.Get(c)
		c.Session = session
		filter[0](c, filter[1:])
		s.SessionManager.Store(c, session)
	}
}
