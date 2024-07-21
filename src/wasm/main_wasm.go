//go:build js && wasm
// +build js,wasm

package main

import (
	"strconv"
	"syscall/js"
	"time"

	"wasm-bugs/src/world"
)

const (
	WORLD_WIDTH  = 600
	WORLD_HEIGHT = 600

	FPS = 60
)

var (
	canvas           js.Value
	ctx              js.Value
	reportCanvas     js.Value
	reportCtx        js.Value
	startButton      js.Value
	pauseButton      js.Value
	resetButton      js.Value
	startingBacteria js.Value
	startingBugs     js.Value
	reseedRate       js.Value
	reportViewButton js.Value
	gameViewButton   js.Value
	reportView       js.Value
	gameView         js.Value

	started    bool
	paused     bool
	screenView world.ScreenView

	gameWorld *world.GameWorld
)

func main() {
	gameWorld = world.NewGameWorld(WORLD_WIDTH, WORLD_HEIGHT)

	doc := js.Global().Get("document")
	window := js.Global().Get("window")

	canvas = doc.Call("getElementById", "gameCanvas")
	ctx = canvas.Call("getContext", "2d")

	canvas.Set("width", window.Get("innerWidth").Int())
	canvas.Set("height", window.Get("innerHeight").Int())

	reportCanvas = doc.Call("getElementById", "reportCanvas")
	reportCtx = reportCanvas.Call("getContext", "2d")

	reportCanvas.Set("width", window.Get("innerWidth").Int())
	reportCanvas.Set("height", window.Get("innerHeight").Int())

	// Add event listener to the start button
	startButton = doc.Call("getElementById", "startButton")
	if startButton.IsNull() {
		println("Failed to get start button")
		return
	}
	startButton.Call("addEventListener", "click", js.FuncOf(startGame))

	pauseButton = doc.Call("getElementById", "pauseButton")
	if pauseButton.IsNull() {
		println("Failed to get pause button")
		return
	}
	pauseButton.Call("addEventListener", "click", js.FuncOf(pauseGame))

	resetButton = doc.Call("getElementById", "restartButton")
	if resetButton.IsNull() {
		println("Failed to get restart button")
		return
	}
	resetButton.Call("addEventListener", "click", js.FuncOf(resetGame))

	startingBacteria = doc.Call("getElementById", "starting_bacteria")
	if startingBacteria.IsNull() {
		println("Failed to get starting bacteria")
		return
	}
	startingBugs = doc.Call("getElementById", "starting_bugs")
	if startingBugs.IsNull() {
		println("Failed to get starting bugs")
		return
	}
	reseedRate = doc.Call("getElementById", "reseed_rate")
	if reseedRate.IsNull() {
		println("Failed to get reseed rate")
		return
	}
	gameViewButton = doc.Call("getElementById", "game-view-btn")
	if gameViewButton.IsNull() {
		println("Failed to get game-view-btn")
		return
	}
	gameViewButton.Call("addEventListener", "click", js.FuncOf(switchView))

	gameView = doc.Call("getElementById", "game-view")
	if gameView.IsNull() {
		println("Failed to get game-view")
		return
	}

	reportViewButton = doc.Call("getElementById", "report-view-btn")
	if reportViewButton.IsNull() {
		println("Failed to get report-view-btn")
		return
	}
	reportViewButton.Call("addEventListener", "click", js.FuncOf(switchView))

	reportView = doc.Call("getElementById", "report-view")
	if reportView.IsNull() {
		println("Failed to get report-view-")
		return
	}

	gameWorld.Initialize(canvas, ctx, reportCanvas, reportCtx)
	gameWorld.DrawBackground(canvas, ctx)
	paused = false
	screenView = world.GAME_VIEW

	// Prevent Go program from exiting
	select {}
}

func enableInputs() {
	startingBacteria.Set("disabled", false)
	startingBugs.Set("disabled", false)
	reseedRate.Set("disabled", false)
}

func disableInputs() {
	startingBacteria.Set("disabled", true)
	startingBugs.Set("disabled", true)
	reseedRate.Set("disabled", true)
}

func setParams() {
	v := startingBacteria.Get("value").String()
	n, err := strconv.Atoi(v)
	if err != nil {
		println("Invalid number for starting bacteria")
	} else {
		gameWorld.InitialBacteria = n
	}

	v = startingBugs.Get("value").String()
	n, err = strconv.Atoi(v)
	if err != nil {
		println("Invalid number for starting bugs")
	} else {
		gameWorld.InitialBugCount = n
	}

	v = reseedRate.Get("value").String()
	n, err = strconv.Atoi(v)
	if err != nil {
		println("Invalid number for reseed rate")
	} else {
		gameWorld.ReseedBacteria = n
	}
}

func resetGame(this js.Value, args []js.Value) interface{} {
	setParams()
	gameWorld.Initialize(canvas, ctx, reportCanvas, reportCtx)

	paused = false

	if !started {
		draw()
	}

	pauseButton.Set("disabled", true)
	startButton.Set("disabled", false)
	resetButton.Set("disabled", false)

	return nil
}

func switchView(this js.Value, args []js.Value) interface{} {
	if screenView == world.GAME_VIEW {
		screenView = world.REPORT_VIEW
		reportView.Set("hidden", false)
		gameView.Set("hidden", true)
	} else {
		screenView = world.GAME_VIEW
		reportView.Set("hidden", true)
		gameView.Set("hidden", false)
	}

	gameWorld.Draw(screenView)

	return nil
}

func pauseGame(this js.Value, args []js.Value) interface{} {
	enableInputs()
	started = false
	paused = true
	pauseButton.Set("disabled", true)
	startButton.Set("disabled", false)
	resetButton.Set("disabled", false)
	return nil
}

func startGame(this js.Value, args []js.Value) interface{} {
	disableInputs()
	started = true
	paused = false
	setParams()
	go gameLoop()

	resetButton.Set("disabled", true)
	pauseButton.Set("disabled", false)
	startButton.Set("disabled", true)

	return nil
}

func stopGame() {
	pauseButton.Set("disabled", true)
	startButton.Set("disabled", false)
	resetButton.Set("disabled", false)
	enableInputs()
}

func gameLoop() {
	const frameDuration = time.Second / FPS

	for {
		start := time.Now()

		err := update()
		draw()
		if err != nil && err == world.NoBugsError {
			stopGame()
			started = false
			break
		}

		elapsed := time.Since(start)
		sleepDuration := frameDuration - elapsed
		if sleepDuration > 0 {
			time.Sleep(sleepDuration)
		} else {
			time.Sleep(1000)
		}

		if !started {
			break
		}
	}
}

func update() error {
	return gameWorld.Next()
}

func draw() {
	gameWorld.Draw(screenView)
}
