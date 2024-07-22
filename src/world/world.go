package world

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"syscall/js"
)

type ScreenView int

const (
	GAME_VIEW ScreenView = iota
	REPORT_VIEW
)

type NoBugsErrorType struct{}

func (n *NoBugsErrorType) Error() string {
	return "all bugs are dead"
}

var NoBugsError *NoBugsErrorType = &NoBugsErrorType{}

type HistoryEntry struct {
	Cycle           int
	BacteriaCount   int
	BacteriaPercent float64
	BugCount        int
	YellowBugs      int
	CyanBugs        int
	MagentaBugs     int
	RedBugs         int
}

type GameWorld struct {
	Height int
	Width  int

	InitialBacteria int // percentage expressed as a whole number, i.e., 5 == 5%
	ReseedBacteria  int
	InitialBugCount int

	cycle         int
	reseedTotal   int
	cells         []byte
	bugs          []*Bug
	history       []HistoryEntry
	bacteriaCount int

	gameCanvas   js.Value
	gameCtx      js.Value
	reportCanvas js.Value
	reportCtx    js.Value

	bugsBottomLine        int
	redBugsBottomLine     int
	magentaBugsBottomLine int
	cyanBugsBottomLine    int
	yellowBugsBottomLine  int
}

func NewGameWorld(width int, height int) *GameWorld {
	result := &GameWorld{
		Width:                 width,
		Height:                height,
		InitialBacteria:       3,
		ReseedBacteria:        10,
		InitialBugCount:       20,
		reseedTotal:           0,
		cycle:                 0,
		bacteriaCount:         0,
		bugsBottomLine:        height,
		redBugsBottomLine:     height,
		magentaBugsBottomLine: height,
		cyanBugsBottomLine:    height,
		yellowBugsBottomLine:  height,
		bugs:                  []*Bug{},
		history:               make([]HistoryEntry, 0),
	}
	result.cells = make([]byte, width*height)

	return result
}

func (w *GameWorld) Initialize(gameCanvas, gameCtx, reportCanvas, reportCtx js.Value) {
	w.bacteriaCount = 0
	w.cycle = 0
	w.reseedTotal = 0
	w.bugs = []*Bug{}
	w.history = []HistoryEntry{}
	w.gameCanvas = gameCanvas
	w.gameCtx = gameCtx
	w.reportCanvas = reportCanvas
	w.reportCtx = reportCtx

	for i := range len(w.cells) {
		if rand.IntN(100) < w.InitialBacteria {
			w.cells[i] = 1
			w.bacteriaCount++
		} else {
			w.cells[i] = 0
		}
	}

	for range w.InitialBugCount {
		x := rand.IntN(w.Width)
		y := rand.IntN(w.Height)
		w.bugs = append(w.bugs, NewBug(x, y))
	}

}

func (w *GameWorld) HasRun() bool {
	return w.cycle != 0
}

func CalculatePosition(x, y, width int) (int, error) {
	if x >= 0 && y >= 0 {
		return (y * width) + x, nil
	}
	return 0, fmt.Errorf("position %d, %d not valid - negative value", x, y)
}

func (w *GameWorld) SetCell(x, y int, value byte) error {
	pos, err := CalculatePosition(x, y, w.Width)
	if err != nil {
		return err
	} else if pos >= len(w.cells) {
		return fmt.Errorf("position %d, %d exceeds the size of the GameWorld", x, y)
	}

	w.cells[pos] = value
	return nil
}

func (w *GameWorld) GetCell(x, y int) (byte, error) {
	pos, err := CalculatePosition(x, y, w.Width)
	if err != nil {
		return 0, err
	} else if pos >= len(w.cells) {
		return 0, fmt.Errorf("position %d, %d exceeds the size of the GameWorld", x, y)
	}

	return w.cells[pos], nil
}

func (w *GameWorld) addHistoryEntry() {
	entry := HistoryEntry{
		Cycle:           w.cycle,
		BacteriaCount:   w.bacteriaCount,
		BacteriaPercent: float64(w.bacteriaCount) / float64(w.Height*w.Width),
		BugCount:        len(w.bugs),
		YellowBugs:      0,
		CyanBugs:        0,
		MagentaBugs:     0,
		RedBugs:         0,
	}

	for _, bug := range w.bugs {
		switch bug.Classification {
		case YELLOW:
			entry.YellowBugs++
		case CYAN:
			entry.CyanBugs++
		case MAGENTA:
			entry.MagentaBugs++
		default:
			entry.RedBugs++
		}
	}

	w.history = append(w.history, entry)
	if len(w.history) > w.Width {
		w.history = w.history[len(w.history)-w.Width:]
	}
}

func (w *GameWorld) Next() error {
	w.cycle++

	if w.cycle%20 == 0 {
		w.addHistoryEntry()
	}

	for w.reseedTotal >= 0 {
		w.reseedTotal -= 100
		for {
			x := rand.IntN(w.Width)
			y := rand.IntN(w.Height)
			v, _ := w.GetCell(x, y)
			if v == 0 {
				w.SetCell(x, y, 1)
				w.bacteriaCount++
				break
			}
		}
	}

	w.reseedTotal += w.ReseedBacteria

	w.updateBugs()

	if len(w.bugs) == 0 {
		return NoBugsError
	}

	return nil
}

func (w *GameWorld) updateBugs() {
	nextGneBugs := []*Bug{}

	for _, b := range w.bugs {
		if b.Age > 800 && b.Energy > 1000 {
			b1 := b.NewBugFrom()
			b1.Mutate(1)
			nextGneBugs = append(nextGneBugs, b1)
			b2 := b.NewBugFrom()
			b2.Mutate(-1)
			nextGneBugs = append(nextGneBugs, b2)
		} else if b.Energy > 0 {
			nextGneBugs = append(nextGneBugs, b)
		}
	}

	for _, b := range nextGneBugs {
		b.Update(w.Width, w.Height)
		b.Energy += w.bacteriaUnderBug(b)
		if b.Energy > 1500 {
			b.Energy = 1500
		}
	}

	w.bugs = nextGneBugs
}

func (w *GameWorld) bacteriaUnderBug(bug *Bug) int {
	result := 0
	yd := bug.Y
	for i := range 3 {
		yd += i - 1
		xd := bug.X
		for j := range 3 {
			xd += j - 1
			v, _ := w.GetCell(xd, yd)
			if v > 0 {
				w.SetCell(xd, yd, 0)
				w.bacteriaCount--
				result++
			}
		}
	}

	return result * 40
}

func (w *GameWorld) drawHUD(ctx js.Value) error {
	ctx.Set("font", "20px Arial")
	ctx.Set("fillStyle", "black")
	ctx.Call("fillText", fmt.Sprintf("Cycle : %d", w.cycle), 30, w.Height+25)

	ratio := float64(w.bacteriaCount) / float64(w.Width*w.Height) * 100
	ctx.Call("fillText", fmt.Sprintf("Bacteria : %d (%2.1f%%)", w.bacteriaCount, ratio), 180, w.Height+25)

	ctx.Call("fillText", fmt.Sprintf("Bugs : %d", len(w.bugs)), 450, w.Height+25)
	return nil
}

func (w *GameWorld) drawBugs(ctx js.Value) {
	for _, b := range w.bugs {
		b.Draw(ctx)
	}
}

func (w *GameWorld) drawGameView() error {
	w.DrawBackground(w.gameCanvas, w.gameCtx)

	w.gameCtx.Set("fillStyle", "green")
	for x := range w.Width {
		for y := range w.Height {
			v, err := w.GetCell(x, y)
			if err != nil {
				return err
			}

			if v != 0 {
				w.gameCtx.Call("fillRect", x, y, 1, 1)
			}
		}
	}

	w.drawBugs(w.gameCtx)

	w.drawHUD(w.gameCtx)

	return nil
}

func (w *GameWorld) drawBugHistory() {
	w.reportCtx.Set("strokeStyle", "lightgray")
	w.reportCtx.Call("beginPath")
	w.reportCtx.Call("moveTo", 0, w.bugsBottomLine)
	w.reportCtx.Call("lineTo", w.Width, w.bugsBottomLine)
	w.reportCtx.Call("stroke")

	w.reportCtx.Set("strokeStyle", "white")
	w.reportCtx.Call("beginPath")

	startIndex := 0
	if len(w.history) > w.Width {
		startIndex = len(w.history) - w.Width
	}

	for i := startIndex; i < len(w.history); i++ {
		h := w.history[i]
		x := i - startIndex
		y := w.bugsBottomLine - h.BugCount - 2
		if y < w.redBugsBottomLine+20 {
			w.redBugsBottomLine = y - 30
		}
		if i == 0 {
			w.reportCtx.Call("moveTo", x, y)
		} else {
			w.reportCtx.Call("lineTo", x, y)
		}
	}
	w.reportCtx.Call("stroke")
}

func (w *GameWorld) drawRedBugsHistory() {
	w.reportCtx.Set("strokeStyle", "lightgray")
	w.reportCtx.Call("beginPath")
	w.reportCtx.Call("moveTo", 0, w.redBugsBottomLine)
	w.reportCtx.Call("lineTo", w.Width, w.redBugsBottomLine)
	w.reportCtx.Call("stroke")

	w.reportCtx.Set("strokeStyle", "red")
	w.reportCtx.Call("beginPath")

	startIndex := 0
	if len(w.history) > w.Width {
		startIndex = len(w.history) - w.Width
	}

	var x int
	for i := startIndex; i < len(w.history); i++ {
		h := w.history[i]
		x = i - startIndex
		y := w.redBugsBottomLine - h.RedBugs - 2
		if y < w.magentaBugsBottomLine+20 {
			w.magentaBugsBottomLine = y - 30
		}
		if i == 0 {
			w.reportCtx.Call("moveTo", x, y)
		} else {
			w.reportCtx.Call("lineTo", x, y)
		}
	}
	w.reportCtx.Call("stroke")

	w.reportCtx.Set("font", "12px Arial")
	w.reportCtx.Set("fillStyle", "red")
	text := strconv.Itoa(w.history[len(w.history)-1].RedBugs)
	textMetrics := w.reportCtx.Call("measureText", text)
	x = x - int(textMetrics.Get("width").Float()) - 5
	if x < 1 {
		x = 1
	}
	w.reportCtx.Call("fillText", text, x, w.redBugsBottomLine-5)
}

func (w *GameWorld) drawMagentaBugsHistory() {
	w.reportCtx.Set("strokeStyle", "lightgray")
	w.reportCtx.Call("beginPath")
	w.reportCtx.Call("moveTo", 0, w.magentaBugsBottomLine)
	w.reportCtx.Call("lineTo", w.Width, w.magentaBugsBottomLine)
	w.reportCtx.Call("stroke")

	w.reportCtx.Set("strokeStyle", "magenta")
	w.reportCtx.Call("beginPath")

	startIndex := 0
	if len(w.history) > w.Width {
		startIndex = len(w.history) - w.Width
	}

	var x int
	for i := startIndex; i < len(w.history); i++ {
		h := w.history[i]
		x = i - startIndex
		y := w.magentaBugsBottomLine - h.MagentaBugs - 2
		if y < w.cyanBugsBottomLine+20 {
			w.cyanBugsBottomLine = y - 30
		}

		if i == 0 {
			w.reportCtx.Call("moveTo", x, y)
		} else {
			w.reportCtx.Call("lineTo", x, y)
		}
	}
	w.reportCtx.Call("stroke")

	w.reportCtx.Set("font", "12px Arial")
	w.reportCtx.Set("fillStyle", "magenta")
	text := strconv.Itoa(w.history[len(w.history)-1].MagentaBugs)
	textMetrics := w.reportCtx.Call("measureText", text)
	x = x - int(textMetrics.Get("width").Float()) - 5
	if x < 1 {
		x = 1
	}
	w.reportCtx.Call("fillText", text, x, w.magentaBugsBottomLine-5)

}

func (w *GameWorld) drawCyanBugsHistory() {
	w.reportCtx.Set("strokeStyle", "lightgray")
	w.reportCtx.Call("beginPath")
	w.reportCtx.Call("moveTo", 0, w.cyanBugsBottomLine)
	w.reportCtx.Call("lineTo", w.Width, w.cyanBugsBottomLine)
	w.reportCtx.Call("stroke")

	w.reportCtx.Set("strokeStyle", "cyan")
	w.reportCtx.Call("beginPath")

	startIndex := 0
	if len(w.history) > w.Width {
		startIndex = len(w.history) - w.Width
	}

	var x int
	for i := startIndex; i < len(w.history); i++ {
		h := w.history[i]
		x = i - startIndex
		y := w.cyanBugsBottomLine - h.CyanBugs - 2
		if y < w.yellowBugsBottomLine+20 {
			w.yellowBugsBottomLine = y - 30
		}
		if i == 0 {
			w.reportCtx.Call("moveTo", x, y)
		} else {
			w.reportCtx.Call("lineTo", x, y)
		}
	}
	w.reportCtx.Call("stroke")

	w.reportCtx.Set("font", "12px Arial")
	w.reportCtx.Set("fillStyle", "cyan")
	text := strconv.Itoa(w.history[len(w.history)-1].CyanBugs)
	textMetrics := w.reportCtx.Call("measureText", text)
	x = x - int(textMetrics.Get("width").Float()) - 5
	if x < 1 {
		x = 1
	}
	w.reportCtx.Call("fillText", text, x, w.cyanBugsBottomLine-5)
}

func (w *GameWorld) drawYellowBugsHistory() {
	w.reportCtx.Set("strokeStyle", "lightgray")
	w.reportCtx.Call("beginPath")
	w.reportCtx.Call("moveTo", 0, w.yellowBugsBottomLine)
	w.reportCtx.Call("lineTo", w.Width, w.yellowBugsBottomLine)
	w.reportCtx.Call("stroke")

	w.reportCtx.Set("strokeStyle", "yellow")
	w.reportCtx.Call("beginPath")

	startIndex := 0
	if len(w.history) > w.Width {
		startIndex = len(w.history) - w.Width
	}

	var x int
	for i := startIndex; i < len(w.history); i++ {
		h := w.history[i]
		x = i - startIndex
		y := w.yellowBugsBottomLine - h.YellowBugs - 2
		if i == 0 {
			w.reportCtx.Call("moveTo", x, y)
		} else {
			w.reportCtx.Call("lineTo", x, y)
		}
	}
	w.reportCtx.Call("stroke")

	w.reportCtx.Set("font", "12px Arial")
	w.reportCtx.Set("fillStyle", "yellow")
	text := strconv.Itoa(w.history[len(w.history)-1].YellowBugs)
	textMetrics := w.reportCtx.Call("measureText", text)
	x = x - int(textMetrics.Get("width").Float()) - 5
	if x < 1 {
		x = 1
	}
	w.reportCtx.Call("fillText", text, x, w.yellowBugsBottomLine-5)
}

func (w *GameWorld) drawBacteriaHistory() {
	w.reportCtx.Set("strokeStyle", "green")
	w.reportCtx.Call("beginPath")

	startIndex := 0
	if len(w.history) > w.Width {
		startIndex = len(w.history) - w.Width
	}

	for i := startIndex; i < len(w.history); i++ {
		x := i - startIndex
		h := w.history[i]
		gap := float64(w.Height) / 50
		y := w.Height - int((h.BacteriaPercent*100)*gap)
		if y < w.bugsBottomLine {
			w.bugsBottomLine = y - 25
		}
		if i == 0 {
			w.reportCtx.Call("moveTo", x, y)
		} else {
			w.reportCtx.Call("lineTo", x, y)
		}
	}
	w.reportCtx.Call("stroke")
}

func (w *GameWorld) drawReportView() error {
	w.DrawBackground(w.reportCanvas, w.reportCtx)

	w.drawHUD(w.reportCtx)

	w.drawBugHistory()
	w.drawBacteriaHistory()
	w.drawRedBugsHistory()
	w.drawMagentaBugsHistory()
	w.drawCyanBugsHistory()
	w.drawYellowBugsHistory()

	return nil
}

func (w *GameWorld) Draw(screenView ScreenView) error {
	if screenView == GAME_VIEW {
		return w.drawGameView()
	} else {
		return w.drawReportView()
	}
}

func (w *GameWorld) DrawBackground(canvas, ctx js.Value) {
	ctx.Call("clearRect", 0, 0, canvas.Get("width").Int(), canvas.Get("height").Int())

	ctx.Set("fillStyle", "black")
	ctx.Call("fillRect", 0, 0, w.Width, w.Height)

	ctx.Set("fillStyle", "gray")
	ctx.Call("fillRect", 0, w.Height, w.Width, 40)
}
