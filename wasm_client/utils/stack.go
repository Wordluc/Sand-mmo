package utils

import (
	"errors"
	"sync"
)

type Stack struct {
	buf [][]byte
	*sync.Mutex
}

func NewStack() Stack {
	return Stack{Mutex: &sync.Mutex{}}
}

func (s *Stack) Push(n []byte) {
	s.Lock()
	s.buf = append(s.buf, n)
	s.Unlock()
}
func (s *Stack) Len() int {
	return len(s.buf)
}
func (s *Stack) Pop() ([]byte, error) {
	s.Lock()

	defer s.Unlock()
	if len(s.buf) == 0 {
		return nil, errors.New("Stack empty")
	}

	res := s.buf[len(s.buf)-1]
	s.buf = s.buf[:len(s.buf)-1]
	return res, nil
}

func (s *Stack) PopAndClean() ([]byte, error) {
	s.Lock()
	defer s.Unlock()
	if len(s.buf) == 0 {
		return nil, errors.New("Stack empty")
	}

	res := s.buf[len(s.buf)-1]
	s.buf = make([][]byte, 0)
	return res, nil
}
