package types

type IdGenerator struct {
	id int
}

func NewIdGenerator() IdGenerator {
	return IdGenerator {
		id: 0,
	}
}

func (self *IdGenerator) Next() int {
	self.id += 1
	return self.id
}
