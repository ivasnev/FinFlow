package uuid

const ConstantUUID = "482ed3b8-6da2-4775-b745-7054eaf0cae6"

type ConstantProvider struct{}

func NewConstantProvider() *ConstantProvider {
	return &ConstantProvider{}
}

func (p ConstantProvider) NewString() string {
	return ConstantUUID
}

