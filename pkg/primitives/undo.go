package primitives

// Undo interface describes functionality of the undo.
type Undo interface {
	State() *State
	Save(newState *State)
	Undo()
	Redo()
	Clear()
}

// History type contains the history of the changing states.
type History struct {
	undos   *Stack
	current State
	redos   *Stack
	limit   int
}

// NewUndo creates new Undo with the undo/redo capacity of `limit`
func NewUndo(limit int) Undoer {
	undos := NewStack(limit)
	redos := NewStack(limit)
	h := &History{
		undos:   undos,
		current: nil,
		redos:   redos,
		limit:   limit,
	}
	return h
}

// State returns current state of the history.
// Notice that initial state is nil
func (h *History) State() State {
	return h.current
}

// Save method changes the current state and saves previous state for undo purposes.
func (h *History) Save(newState State) {
	h.undos.Push(h.current)
	h.current = newState
	h.redos.Clear()
}

// Undo method undos the state.
func (h *History) Undo() {
	state := h.undos.Pop()
	if state == nil {
		// no undo avaliable
		return
	}
	h.redos.Push(h.current)
	h.current = state
}

// Redo method redos the state.
func (h *History) Redo() {
	state := h.redos.Pop()
	if state == nil {
		return
	}
	h.undos.Push(h.current)
	h.current = state
}

// Clear method clears history
func (h *History) Clear() {
	h.undos.Clear()
	h.redos.Clear()
}
