package model

type scrollState struct {
	offset int
	max    int
}

func (s *scrollState) setMax(max int) {
	s.max = max
	if s.max < 0 {
		s.max = 0
	}
	s.clamp()
}

func (s *scrollState) down(n int) {
	s.offset += n
	s.clamp()
}

func (s *scrollState) up(n int) {
	s.offset -= n
	s.clamp()
}

func (s *scrollState) clamp() {
	if s.offset < 0 {
		s.offset = 0
	}
	if s.offset > s.max {
		s.offset = s.max
	}
}
