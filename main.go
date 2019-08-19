package main

// TODO: Start splitting off code into packages for organiazation
import (
	"bufio"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"

	sdlmix "github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
)

// TODO: Rewrite drawing to be proportional to the window size
// TODO: Create a few window sizes to choose from (config file?)
const windowWidth, windowHeight int = 1080, 720
const maxScore int = 11

const (
	serveLeft = iota
	serveRight
)

// --

type gameState int

const (
	start gameState = iota
	play
)

var state = start

// --

var dir = serveRight

// TODO: Create enum of standard colors
type color struct {
	r, g, b byte
}

type position struct {
	x, y float32
}

type ball struct {
	position
	radius float32
	xv, yv float32
	color  color
	sfx    []*sdlmix.Chunk
}

func (b *ball) reset(dir int) {
	servingDirection := -1.0
	if dir == serveRight {
		servingDirection = 1.0
	}
	b.position = getCenter()
	b.xv = float32(math.Abs(float64(b.xv)) * servingDirection)
	b.yv = b.yv * float32(math.Pow(-1, float64(rand.Intn(24))))
}

func (b *ball) draw(pixels []byte) {
	for y := -b.radius; y < b.radius; y++ {
		for x := -b.radius; x < b.radius; x++ {
			if x*x+y*y < b.radius*b.radius {
				setPixel(int(b.x+x), int(b.y+y), b.color, pixels)
			}
		}
	}
}

// TODO: Figure out what to do when the paddle collides on top or bottom
// 	Right now, it looks a little funny when the ball pops out
func (b *ball) update(leftP, rightP *paddle, elapsedTime float32) {
	b.x += b.xv * elapsedTime
	b.y += b.yv * elapsedTime
	hitWall, hitPaddle, scored := false, false, false

	if b.y-b.radius < 0 {
		b.yv = -b.yv
		b.y = b.radius
		hitWall = true
	}

	if b.y+b.radius > float32(windowHeight) {
		b.yv = -b.yv
		b.y = float32(windowHeight) - b.radius
		hitWall = true
	}

	if b.x-b.radius < 0 || b.x+b.radius > float32(windowWidth) {
		if b.x-b.radius < 0 {
			rightP.score++
			dir = serveLeft
		} else {
			leftP.score++
			dir = serveRight
		}

		state = start
		scored = true
	}

	if b.x-b.radius < leftP.x+leftP.w/2 {
		if b.y > leftP.y-leftP.h/2 && b.y < leftP.y+leftP.h/2 {
			b.xv = -b.xv
			b.x = leftP.x + leftP.w/2.0 + b.radius
			hitPaddle = true
		}
	}

	if b.x+b.radius > rightP.x-rightP.w/2 {
		if b.y > rightP.y-leftP.h/2 && b.y < rightP.y+rightP.h/2 {
			b.xv = -b.xv
			b.x = rightP.x - rightP.w/2.0 - b.radius
			hitPaddle = true
		}
	}

	if hitWall {
		b.sfx[1].Play(1, 0)
	}
	if hitPaddle {
		b.sfx[2].Play(1, 0)
	}
	if scored {
		b.sfx[3].Play(1, 0)
	}
}

type paddle struct {
	position
	w, h  float32
	speed float32
	color color
	score int
}

func (p *paddle) draw(pixels []byte) {
	startX := p.x - p.w/2
	startY := p.y - p.h/2

	for y := float32(0); y < p.h; y++ {
		for x := float32(0); x < p.w; x++ {
			setPixel(int(startX+x), int(startY+y), p.color, pixels)
		}
	}
}

func (p *paddle) update(keyState []uint8, elapsedTime float32) {
	if keyState[sdl.SCANCODE_UP] != 0 {
		p.y -= p.speed * elapsedTime
	}
	if keyState[sdl.SCANCODE_DOWN] != 0 {
		p.y += p.speed * elapsedTime
	}
}

// TODO: Create a better AI that makes decisions based on arbitrary things
// TODO: Create multiple AIs that you could easily switch between
// TODO: Menu to select the opponent (AI algorithm) that you play against
func (p *paddle) aiUpdate(b *ball, elapsedTime float32) {
	if b.xv > 0 {
		if b.y <= p.y {
			p.y -= p.speed * elapsedTime
		} else {
			p.y += p.speed * elapsedTime
		}
	}
}

func (p *paddle) collision(b *ball, elapsedTime float32) {
	if p.y-p.h/2 < 0 {
		p.y = p.h / 2
	}

	if p.y+p.h/2 > float32(windowHeight) {
		p.y = float32(windowHeight) - p.h/2
	}
}

func drawScore(p1 *paddle, p2 *paddle, font map[rune][]byte, pixels []byte) {
	player1Score := strconv.Itoa(p1.score)
	player2Score := strconv.Itoa(p2.score)

	numX := lerp(0, getCenter().x, 0.3)
	drawString(position{numX, 35}, p1.color, 10, player1Score, font, pixels)

	numX = lerp(getCenter().x, float32(windowWidth), 0.7)
	drawString(position{numX, 35}, p2.color, 10, player2Score, font, pixels)
}

// TODO: Account for position based on size within the func
func drawString(pos position, c color, size int, msg string, font map[rune][]byte, pixels []byte) {
	startX := int(pos.x) - size*3/2
	startY := int(pos.y) - size*5/2

	for _, char := range msg {
		for i, v := range font[char] {
			if v == 49 {
				for y := startY; y < startY+size; y++ {
					for x := startX; x < startX+size; x++ {
						setPixel(x, y, c, pixels)
					}
				}
			}
			startX += size
			if (i+1)%3 == 0 {
				startY += size
				startX -= size * 3
			}
		}

		startY = int(pos.y) - size*5/2
		startX += 4 * size
	}
}

func getCenter() position {
	return position{float32(windowWidth / 2), float32(windowHeight / 2)}
}

func lerp(a, b, pct float32) float32 {
	return a + pct*(b-a)
}

// TODO: Create different fonts you can load in
// TODO: Include even normal fonts like .ttf, etc
func loadFont(filename string) (map[rune][]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	font := make(map[rune][]byte, 36)

	scanner := bufio.NewScanner(f)
	hash := rune('0')
	for scanner.Scan() {
		row := make([]byte, 15)
		for i, ch := range scanner.Text() {
			if i == 0 {
				hash = rune(ch)
			} else {
				row[i-1] = byte(ch)
			}
		}

		font[hash] = row
	}

	return font, nil
}

func setPixel(x, y int, c color, pixels []byte) {
	index := (y*windowWidth + x) * 4

	if index < len(pixels)-4 && index >= 0 {
		pixels[index] = c.r
		pixels[index+1] = c.g
		pixels[index+2] = c.b
	}
}

func clear(pixels []byte) {
	for i := range pixels {
		pixels[i] = 0
	}
}

func drawCenterLine(pixels []byte) {
	startX := windowWidth / 2
	for y := 0; y < windowHeight; y++ {
		if int(y/12)%2 == 0 {
			for x := -1; x < 2; x++ {
				setPixel(startX+x, y, color{255, 255, 255}, pixels)
			}
		}
	}
}

func main() {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow(
		"Pong",
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		int32(windowWidth),
		int32(windowHeight),
		sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	err = sdlmix.Init(0)
	if err != nil {
		panic(err)
	}
	defer sdlmix.Quit()

	sndfmt := uint16(sdlmix.DEFAULT_FORMAT)
	err = sdlmix.OpenAudio(44100, sndfmt, 2, 1024)
	if err != nil {
		panic(err)
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	defer renderer.Destroy()

	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(windowWidth), int32(windowHeight))
	if err != nil {
		panic(err)
	}
	defer texture.Destroy()

	pixels := make([]byte, windowWidth*windowHeight*4)

	font, err := loadFont("assets/font")
	if err != nil {
		panic(err)
	}

	sfxWall, err := sdlmix.LoadWAV("assets/wall.wav")
	if err != nil {
		panic(err)
	}
	sfxPaddle, err := sdlmix.LoadWAV("assets/paddle.wav")
	if err != nil {
		panic(err)
	}
	sfxScore, err := sdlmix.LoadWAV("assets/score.wav")
	if err != nil {
		panic(err)
	}
	sfxScore.Volume(34)

	player1 := paddle{
		position: position{50, 100},
		w:        12,
		h:        50,
		speed:    800,
		score:    0,
		color:    color{255, 255, 255}}
	player2 := paddle{
		position: position{float32(windowWidth) - 50, 100},
		w:        12,
		h:        50,
		speed:    560,
		score:    0,
		color:    color{255, 255, 255}}
	ball := ball{
		position: getCenter(),
		radius:   8,
		xv:       590,
		yv:       590,
		color:    color{255, 255, 255},
		sfx:      []*sdlmix.Chunk{nil, sfxWall, sfxPaddle, sfxScore}}
	msg := ""
	// TODO: This is where I want to move the position code to account for size
	centeredPos := position{getCenter().x - 240.0, getCenter().y}

	// TODO: Handle a controller
	// TODO: Handle more than one player, instead of AI
	keyState := sdl.GetKeyboardState()

	var frameStart time.Time
	var elapsedTime float32

	running := true
	for running {
		frameStart = time.Now()

		// TODO: Create menu states
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				running = false
				break
			}
		}

		if state == play {
			player1.update(keyState, elapsedTime)
			player2.aiUpdate(&ball, elapsedTime)
			ball.update(&player1, &player2, elapsedTime)

			player1.collision(&ball, elapsedTime)
			player2.collision(&ball, elapsedTime)
		} else if state == start {
			if player1.score == maxScore || player2.score == maxScore {
				msg = "PLAYER 1 WINS"
				if player2.score == maxScore {
					msg = "PLAYER 2 WINS"
				}
			}
			ball.reset(dir)
			if keyState[sdl.SCANCODE_SPACE] != 0 {
				if player1.score == maxScore || player2.score == maxScore {
					player1.score = 0
					player2.score = 0
					msg = ""
				}
				state = play
			}
		}

		clear(pixels)
		drawCenterLine(pixels)
		drawScore(&player1, &player2, font, pixels)
		player1.draw(pixels)
		player2.draw(pixels)
		ball.draw(pixels)
		if len(msg) > 0 {
			drawString(centeredPos, color{255, 255, 255}, 10, msg, font, pixels)
		}

		err = texture.Update(nil, pixels, windowWidth*4)
		if err != nil {
			panic(err)
		}
		err = renderer.Copy(texture, nil, nil)
		if err != nil {
			panic(err)
		}
		renderer.Present()

		elapsedTime = float32(time.Since(frameStart).Seconds())
		if elapsedTime < .005 {
			sdl.Delay(5 - uint32(elapsedTime/1000.0))
			elapsedTime = float32(time.Since(frameStart).Seconds())
		}
	}
}
