package primitives_test

import "testing"

func TestUndo(t *testing.T) {
	undo := NewUndo(10)
	state := undo.State()
	if state != nil {
		t.Errorf("state not initialized, should be nil, was %v", state)
	}

	undo.Save(1)
	undo.Save(2)
	state = undo.State().(int)
	if state != 2 {
		t.Errorf("state should be 2, was %v", state)
	}

	undo.Undo()
	state = undo.State().(int)
	if state != 1 {
		t.Errorf("state should be 1, was %v", state)
	}

	undo.Undo()
	state = undo.State()
	if state != 1 {
		t.Errorf("state should be 1, was %v", state)
	}

	undo.Redo()
	state = undo.State()
	if state != 2 {
		t.Errorf("state should be 2, was %v", state)
	}

	undo.Redo()
	state = undo.State()
	if state != 2 {
		t.Errorf("state should be 2, was %v", state)
	}
}

func TestUnlimitedUndo(t *testing.T) {
	undo := Newundo(0)
	num := 100
	for i := 0; i < num; i++ {
		undo.Save(i)
		state := undo.State()
		if state != i {
			t.Errorf("state should be %v, was %v", i, state)
		}
		for j := 0; j < i; j++ {
			undo.Undo()
			state := undo.State()
			if state != i-j-1 {
				t.Errorf("state should be %v, was %v", i-j-1, state)
			}
		}
		for j := 0; j < i; j++ {
			undo.Redo()
			state := undo.State()
			if state != j+1 {
				t.Errorf("state should be %v, was %v", i-j-1, state)
			}
		}
	}
}

func TestLimitedUndo(t *testing.T) {
	undo := Newundo(2)
	undo.Save(1)
	undo.Save(2)
	undo.Save(3)
	undo.Save(4)

	undo.Undo()
	undo.Undo()

	state := undo.State()
	if state != 2 {
		t.Errorf("state should be 2, was %v", state)
	}
	undo.Undo()
	state = undo.State()
	if state != 2 {
		t.Errorf("state should be 2, was %v", state)
	}

	undo.Redo()
	state = undo.State()
	if state != 3 {
		t.Errorf("state should be 3, was %v", state)
	}

	undo.Save(5)
	undo.Redo()
	state = undo.State()
	if state != 5 {
		t.Errorf("state should be 5, was %v", state)
	}
}

func TestClearUndo(t *testing.T) {
	undo := Newundo(2)
	undo.Save(1)
	undo.Save(2)
	undo.Save(3)
	undo.Save(4)
	undo.Clear()

	state := undo.State()
	if state != 4 {
		t.Errorf("state should be 4, was %v", state)
	}

	// Does nothing
	undo.Undo()
	state = undo.State()
	if state != 4 {
		t.Errorf("state should be 4, was %v", state)
	}

	//does nothing
	undo.Redo()
	state = undo.State()
	if state != 4 {
		t.Errorf("state should be 4, was %v", state)
	}
}
