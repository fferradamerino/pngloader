package pngloader

type PixelRGB struct {
	Red byte
	Green byte
	Blue byte
}

type PNGData struct {
	IsValid bool
	Header [8]byte

	Width uint
	Height uint
	BitDepth byte
	ColorType byte
	CompressionMethod byte
	FilterMethod byte
	InterlaceMethod byte

	Palette []PixelRGB
	Background PixelRGB

	Data []byte
}

type RawData struct {
	Width uint32
	Height uint32
	Format byte

	Data []byte
}
