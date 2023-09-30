package main

type Board struct {
	cards []bool
}

func NewBoard(size int) *Board {
	return &Board{
		cards: make([]bool, size),
	}
}

func (board *Board) Flip(index int, down bool) error {
	if board.InBounds(index) {
		board.cards[index] = down
	}
	return nil
}

func (board *Board) InBounds(index int) bool {
	return index > 0 && index < len(board.cards)
}
