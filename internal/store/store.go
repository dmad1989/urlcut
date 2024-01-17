package store

type urlMap map[string]string

func New() *urlMap {
	res := make(urlMap, 2)
	return &res
}

func (u urlMap) Get(key string) string {
	return u[key]
}

func (u urlMap) Add(key, value string) {
	u[key] = value
}

func (u urlMap) Has(key string) bool {
	_, ok := u[key]
	return ok
}

func (u urlMap) GetKey(value string) (res string) {
	if len(u) == 0 {
		return
	}
	for key, val := range u {
		if val == value {
			return key
		}
	}
	return
}
