package music

type Jukebox struct {
}

type Source interface {
}

type Track interface {
}

func NewJukebox() *Jukebox {
	return &Jukebox{}
}
