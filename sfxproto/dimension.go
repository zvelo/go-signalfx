package sfxproto

// NewDimension creates a new Dimension with the given key and value
func NewDimension(key, value string) *Dimension {
	return &Dimension{
		Key:   key,
		Value: value,
	}
}
