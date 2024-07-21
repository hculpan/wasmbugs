package world

import (
	"log/slog"
	"math/rand/v2"
	"syscall/js"
)

const (
	YELLOW  = "Yellow"
	CYAN    = "Cyan"
	MAGENTA = "Magenta"
	RED     = "Red"
)

type Bug struct {
	X int
	Y int

	Age            int
	Energy         int
	Classification string

	direction      int
	geneValue      [6]int
	geneWeight     [6]int
	totalOfWeights int
}

func NewBug(x, y int) *Bug {
	result := &Bug{
		X:         x,
		Y:         y,
		Energy:    400,
		Age:       0,
		direction: rand.IntN(6),
	}

	result.totalOfWeights = 0
	for i := range 6 {
		result.geneValue[i] = rand.IntN(4) - 2
		result.geneWeight[i] = result.geneValue[i] * result.geneValue[i]
		result.totalOfWeights += result.geneWeight[i]
	}
	result.SetClassification()

	return result
}

func (b *Bug) NewBugFrom() *Bug {
	result := &Bug{
		X:         b.X,
		Y:         b.Y,
		direction: b.direction,
		Energy:    b.Energy / 2,
		Age:       0,
	}

	result.geneValue = [6]int{}
	result.geneWeight = [6]int{}
	result.totalOfWeights = 0
	for i := range 6 {
		result.geneValue[i] = b.geneValue[i]
		result.geneWeight[i] = result.geneValue[i] * result.geneValue[i]
		result.totalOfWeights += result.geneWeight[i]
	}
	result.SetClassification()

	return result
}

func (b *Bug) Mutate(delta int) {
	b.totalOfWeights = 0
	n := rand.IntN(6)
	b.geneValue[n] += delta
	for i := range 6 {
		b.geneWeight[i] = b.geneValue[i] * b.geneValue[i]
		b.totalOfWeights += b.geneWeight[i]
	}
	b.SetClassification()
}

func (b *Bug) selectTurn() int {
	n := rand.IntN(b.totalOfWeights)
	for i := range 6 {
		if n < b.geneWeight[i] {
			return i
		}
	}

	return 5
}

func (b *Bug) move(width, height int) (int, int) {
	turn := b.selectTurn()
	b.direction = (b.direction + turn) % 6

	x := b.X
	y := b.Y
	switch b.direction {
	case 0:
		y += 2
	case 1:
		x += 2
		y += 1
	case 2:
		x += 2
		y += -1
	case 3:
		y += -2
	case 4:
		x += -2
		y += -1
	case 5:
		x += -2
		y += 1
	default:
		slog.Error("invalid direction", "direction", b.direction)
	}

	if x < 0 {
		x += width
	} else if x >= width {
		x -= width
	}
	if y < 0 {
		y += height
	} else if y >= height {
		y -= height
	}

	return x, y
}

func (b *Bug) SetClassification() {
	forwardMove := (float64(b.geneWeight[0]) / float64(b.totalOfWeights)) * 100
	b.Classification = RED
	if forwardMove > 80 {
		b.Classification = YELLOW
	} else if forwardMove > 50 {
		b.Classification = CYAN
	} else if forwardMove > 25 {
		b.Classification = MAGENTA
	}
}

func (b *Bug) Update(width, height int) {
	b.X, b.Y = b.move(width, height)

	b.Age++
	b.Energy--
}

func (b *Bug) Draw(ctx js.Value) {
	if b.Classification == YELLOW {
		ctx.Set("fillStyle", "yellow")
	} else if b.Classification == CYAN {
		ctx.Set("fillStyle", "cyan")
	} else if b.Classification == MAGENTA {
		ctx.Set("fillStyle", "magenta")
	} else {
		ctx.Set("fillStyle", "red")
	}
	ctx.Call("fillRect", b.X-1, b.Y-1, 3, 3)
}
