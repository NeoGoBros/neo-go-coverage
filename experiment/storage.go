package experiment

type Storage int

type Context int

var realStorage map[string]interface{} = make(map[string]interface{})

func MakeStorage() Storage {
	return 0
}

func (s *Storage) GetContext() Context {
	return 0
}

func (s *Storage) Get(c Context, key []byte) interface{} {
	key1 := string(key[:])
	result, ok := realStorage[key1]
	if ok {
		return result
	} else {
		return nil
	}
}

func (s *Storage) Put(c Context, key []byte, value interface{}) {
	key1 := string(key[:])
	realStorage[key1] = value
}

var storage = MakeStorage()
