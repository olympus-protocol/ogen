package primitives

import "testing"

func TestSimpleStack(t *testing.T) {
	s := NewStack(10)

	len := s.Len()
	if len != 0 {
		t.Errorf("Length of an empty stack should be 0, was %v", len)
	}

	s.Push(1)

	len = s.Len()
	if len != 1 {
		t.Errorf("Length should be 0, was %v", len)
	}

	pop := s.Pop().(int)
	if pop != 1 {
		t.Errorf("Top item should have been 1, was %v", pop)
	}

	len = s.Len()
	if len != 0 {
		t.Errorf("Stack should be empty, len was %v", len)
	}

	s.Push(1)
	s.Push(2)

	len = s.Len()
	if len != 2 {
		t.Errorf("Length should be 2, was %v", len)
	}

	pop = s.Pop().(int)
	if pop != 2 {
		t.Errorf("Top of the stack should be 2, was %v", pop)
	}
}

func TestUnlimitedStack(t *testing.T) {
	s := NewStack(0)
	num := 10000
	for i := 0; i < num; i++ {
		s.Push(i)
		len := s.Len()
		if len != i+1 {
			t.Errorf("Length should be %v, was %v", i+1, len)
		}
	}
	for i := num - 1; i >= 0; i-- {
		pop := s.Pop()
		if pop != i {
			t.Errorf("Top of the stack should be %v, was %v", i, pop)
		}
	}
}

func TestLimitedStack(t *testing.T) {
	s := NewStack(2)
	s.Push(1)
	s.Push(2)
	s.Push(3)

	len := s.Len()
	if len != 2 {
		t.Errorf("Length should be 2, was %v", len)
	}

	pop := s.Pop().(int)
	if pop != 3 {
		t.Errorf("Top of the stack should be 3, was %v", pop)
	}

	s.Push(4)
	s.Push(5)

	pop = s.Pop().(int)
	if pop != 5 {
		t.Errorf("Top of the stack should be 5, was %v", pop)
	}
	pop = s.Pop().(int)
	if pop != 4 {
		t.Errorf("Top of the stack should be 4, was %v", pop)
	}
	len = s.Len()
	if len != 0 {
		t.Errorf("Length should be 0, was %v", len)
	}
}

func TestClearStack(t *testing.T) {
	s := NewStack(2)
	s.Push(1)
	s.Push(2)
	s.Push(3)
	s.Clear()

	len := s.Len()
	if len != 0 {
		t.Errorf("Length should be 0, was %v", len)
	}

	pop := s.Pop()
	if pop != nil {
		t.Errorf("Stack should be empty, was %v", pop)
	}
}
