package idGen

type IdGenerator struct {
	id int
}

func New() IdGenerator {
	return IdGenerator {
		id: 0,
	}
}

func (self *IdGenerator) Next() int {
	self.id += 1
	return self.id
}
