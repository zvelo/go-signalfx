package sfxproto

func NewDimension(key, value string) *Dimension {
	return &Dimension{
		Key:   key,
		Value: value,
	}
}
