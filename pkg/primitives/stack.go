package primitives

// Stack struct
type Stack struct {
	states   []State
	capacity int
}

// NewStack creates new Stack object and returns its pointer.
// NewStack receives capacity as an argument, which means what is the maximum
// capacity of the stack. If capacity is zero stack has no limit.
func NewStack(capacity int) *Stack {
	s := make([]State, 0, 1)
	stack := &Stack{
		states:   s,
		capacity: capacity,
	}
	return stack
}

// Pop method pops last added element out of the stack and returns it.
// If stack is empty Pop() returns nil.
func (s *Stack) Pop() State {
	n := len(s.states)
	if n == 0 {
		return nil
	}
	v := s.states[n-1]
	s.states = s.states[:n-1]
	return v
}

// Push method pushes element into the stack
func (s *Stack) Push(elem State) {
	n := len(s.states)
	if s.capacity > 0 && n == s.capacity {
		s.states = s.states[1:n]
	}
	s.states = append(s.states, elem)
}

// Len method returns number of elements in stack
func (s *Stack) Len() int {
	return len(s.states)
}

// Clear method empties the stack
func (s *Stack) Clear() {
	s.states = make([]State, 0, 1)
}
