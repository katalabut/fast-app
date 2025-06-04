package configloader

func New[T any](opts ...Option) (*T, error) {
	var cfg T
	ops := []Option{
		WithEnv(""),
	}
	ops = append(ops, opts...)

	parser, err := NewParser(ops...)
	if err != nil {
		return nil, err
	}

	err = parser.Parse(&cfg)

	return &cfg, err
}
