package viewer

type keyMap struct {
	Down     keyBinding
	Up       keyBinding
	PageDown keyBinding
	PageUp   keyBinding
	NextFile keyBinding
	PrevFile keyBinding
	NextHunk keyBinding
	PrevHunk keyBinding
	Toggle   keyBinding
	Filter   keyBinding
	Apply    keyBinding
	Clear    keyBinding
	Palette  keyBinding
	Quit     keyBinding
}

type keyBinding struct {
	keys []string
}

func (k keyBinding) Keys() []string {
	return k.keys
}

func defaultKeyMap() keyMap {
	return keyMap{
		Down:     keyBinding{keys: []string{"j", "down"}},
		Up:       keyBinding{keys: []string{"k", "up"}},
		PageDown: keyBinding{keys: []string{"ctrl+d"}},
		PageUp:   keyBinding{keys: []string{"ctrl+u"}},
		NextFile: keyBinding{keys: []string{"]"}},
		PrevFile: keyBinding{keys: []string{"["}},
		NextHunk: keyBinding{keys: []string{"n"}},
		PrevHunk: keyBinding{keys: []string{"N"}},
		Toggle:   keyBinding{keys: []string{"tab"}},
		Filter:   keyBinding{keys: []string{"/"}},
		Apply:    keyBinding{keys: []string{"enter"}},
		Clear:    keyBinding{keys: []string{"esc"}},
		Palette:  keyBinding{keys: []string{"ctrl+k", "p"}},
		Quit:     keyBinding{keys: []string{"q", "ctrl+c"}},
	}
}
