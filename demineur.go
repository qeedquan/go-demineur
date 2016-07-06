package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	_ "image/png"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/qeedquan/go-media/sdl"
	"github.com/qeedquan/go-media/sdl/sdlimage/sdlcolor"
	"github.com/qeedquan/go-media/sdl/sdlttf"
)

var (
	conf struct {
		assets     string
		width      int
		height     int
		mines      int
		difficulty int
		fullscreen bool
		cheat      bool
	}

	elapsed time.Duration
	ticker  *time.Ticker
	grid    *Grid

	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
	canvas   *image.RGBA

	master            *image.RGBA
	numbers           [10]*image.RGBA
	blue, black       *image.RGBA
	cross             *image.RGBA
	flip1, flip2      *image.RGBA
	mineBlue, mineRed *image.RGBA

	font *sdlttf.Font
	text struct {
		lose    *sdl.Surface
		win     *sdl.Surface
		restart *sdl.Surface
	}
)

var levels = [][3]int{
	{6, 6, 13},
	{10, 10, 20},
	{15, 15, 25},
	{20, 20, 40},
	{30, 30, 50},
	{40, 30, 50},
}

func main() {
	runtime.LockOSThread()
	parseFlags()
	initSDL()
	loadAssets()

	reset()
	for {
		play()
	}
}

func play() {
	defer func() {
		if r := recover(); r != nil && r != "reset" {
			panic(r)
		}
	}()

	event('p')
	update()
	blit()
}

func parseFlags() {
	conf.assets = filepath.Join(sdl.GetBasePath(), "assets")
	flag.StringVar(&conf.assets, "assets", conf.assets, "assets directory")
	flag.BoolVar(&conf.fullscreen, "fullscreen", false, "fullscreen mode")
	flag.BoolVar(&conf.cheat, "cheat", false, "invincible to mines")
	str := fmt.Sprintf("difficulty level [0-%v] (easiest to hardest)", len(levels)-1)
	flag.IntVar(&conf.difficulty, "difficulty", len(levels)/2, str)
	flag.Parse()

	if !(0 <= conf.difficulty && conf.difficulty < len(levels)) {
		ck(fmt.Errorf("Invalid difficulty: %v", conf.difficulty))
	}

	conf.width = levels[conf.difficulty][0]
	conf.height = levels[conf.difficulty][1]
	conf.mines = conf.width * conf.height / 8
}

func initSDL() {
	err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_TIMER)
	ck(err)

	err = sdlttf.Init()
	ck(err)

	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "best")
	w := conf.width * 30
	h := conf.height * 30
	wflag := sdl.WINDOW_RESIZABLE
	if conf.fullscreen {
		wflag |= sdl.WINDOW_FULLSCREEN_DESKTOP
	}
	window, renderer, err = sdl.CreateWindowAndRenderer(w, h, wflag)
	ck(err)

	renderer.SetLogicalSize(w, h)
	window.SetTitle("Demineur")

	texture, err = renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, w, h)
	ck(err)

	canvas = image.NewRGBA(image.Rect(0, 0, w, h))
}

func loadAssets() {
	master = loadImage("master.png")
	for i := range numbers {
		numbers[i] = subimage(i)
	}
	mineBlue = subimage(9)
	blue = subimage(10)
	cross = subimage(11)
	flip1 = subimage(12)
	flip2 = subimage(13)
	mineRed = subimage(14)
	black = subimage(15)

	font = loadFont("DejaVuSans.ttf", levels[conf.difficulty][2])
	text.lose = renderText("YOU LOSE")
	text.win = renderText("YOU WIN")
	text.restart = renderText("PRESS SPACE TO RESTART")
}

func loadImage(name string) *image.RGBA {
	name = filepath.Join(conf.assets, name)
	f, err := os.Open(name)
	ck(err)
	defer f.Close()

	m, _, err := image.Decode(f)
	ck(err)

	b := m.Bounds()
	p := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(p, p.Bounds(), m, image.ZP, draw.Over)

	return p
}

func loadFont(name string, ptSize int) *sdlttf.Font {
	name = filepath.Join(conf.assets, name)
	font, err := sdlttf.OpenFont(name, ptSize)
	ck(err)
	return font
}

func subimage(id int) *image.RGBA {
	return master.SubImage(image.Rect(id*30, 0, (id+1)*30, 30)).(*image.RGBA)
}

func renderText(text string) *sdl.Surface {
	fg := sdl.Color{0, 0, 0x55, 0xff}
	bg := sdl.Color{0xa0, 0xa0, 0xa0, 0xff}
	surface, err := font.RenderUTF8Shaded(text, fg, bg)
	ck(err)
	return surface
}

func reset() {
	rand.Seed(time.Now().UnixNano())
	elapsed = 0
	if ticker != nil {
		ticker.Stop()
	}
	ticker = time.NewTicker(1 * time.Second)

	grid = newGrid(conf.width, conf.height, conf.mines)
	grid.Draw()
}

func event(state int) {
	for {
		ev := sdl.PollEvent()
		if ev == nil {
			break
		}
		switch ev := ev.(type) {
		case sdl.QuitEvent:
			os.Exit(0)
		case sdl.KeyDownEvent:
			switch ev.Sym {
			case sdl.K_ESCAPE:
				os.Exit(0)
			case sdl.K_SPACE:
				reset()
				panic("reset")
			}
		case sdl.MouseButtonUpEvent:
			if state == 'p' {
				step(ev)
			}
		}
	}
}

func step(ev sdl.MouseButtonUpEvent) {
	g := grid
	x := int(ev.X/30) % conf.width
	y := int(ev.Y/30) % conf.height
	s := g.Get(x, y)
	if s == nil || s.Seen {
		return
	}

	switch ev.Button {
	case 1:
		if s.Warn {
			return
		}

		flip(fill(x, y))
		if !conf.cheat && s.Value == 9 {
			over('l', x, y)
		}

	case 3:
		s.Warn = !s.Warn
		if s.Warn {
			idraw(s, cross)
		} else {
			idraw(s, blue)
		}
	}
}

func fill(x, y int) []image.Point {
	g := grid
	s := g.Get(x, y)
	if s.Seen || s.Warn {
		return nil
	}
	s.Seen = true
	if s.Value != 9 {
		g.Count--
	}

	p := []image.Point{{x, y}}
	for i := 0; i < len(p); i++ {
		s := g.Get(p[i].X, p[i].Y)
		if s == nil || s.Value != 0 {
			continue
		}

		for j := -1; j <= 1; j++ {
			for k := -1; k <= 1; k++ {
				if j == 0 && k == 0 {
					continue
				}
				x := p[i].X + k
				y := p[i].Y + j
				s = g.Get(x, y)
				if s != nil && !(s.Seen || s.Warn) {
					s.Seen = true
					g.Count--
					p = append(p, image.Pt(x, y))
				}
			}
		}
	}
	return p
}

func flip(p []image.Point) {
	g := grid
	for _, p := range p {
		idraw(g.Get(p.X, p.Y), flip1)
	}
	delay(50)

	for _, p := range p {
		idraw(g.Get(p.X, p.Y), black)
	}
	delay(50)

	for _, p := range p {
		idraw(g.Get(p.X, p.Y), flip2)
	}
	delay(100)

	for _, p := range p {
		s := g.Get(p.X, p.Y)
		n := s.Value
		idraw(s, numbers[n])
	}
}

func update() {
	select {
	case <-ticker.C:
		elapsed += 1 * time.Second
	default:
	}
	title := fmt.Sprintf("Demineur: Elapsed: %v Squares Left: %v", elapsed, grid.Count)
	window.SetTitle(title)
	if grid.Count == 0 {
		over('w', -1, -1)
	}
}

func blit() {
	renderer.SetDrawColor(sdlcolor.Black)
	renderer.Clear()
	texture.Update(nil, canvas.Pix, canvas.Stride)
	renderer.Copy(texture, nil, nil)
	renderer.Present()
}

func over(state, x, y int) {
	var result *sdl.Surface
	if state == 'l' {
		s := grid.Get(x, y)
		for i := 0; i < 5; i++ {
			if i&1 != 0 {
				idraw(s, mineBlue)
			} else {
				idraw(s, mineRed)
			}
			delay(500)
		}
		result = text.lose
	} else {
		flip(grid.Mines)
		delay(500)
		result = text.win
	}

	r := result.Bounds()
	rw, rh := r.Dx(), r.Dy()
	rx := (conf.width*30 - rw) / 2
	ry := (conf.height*30 - rh - 20) / 2
	draw.Draw(canvas, image.Rect(rx, ry, rx+rw, ry+rh), result, image.ZP, draw.Over)

	h := rh + 60
	r = text.restart.Bounds()
	rw, rh = r.Dx(), r.Dy()
	rx = (conf.width*30 - rw) / 2
	ry = (conf.height*30 - rh - 20 + h) / 2
	draw.Draw(canvas, image.Rect(rx, ry, rx+rw, ry+rh), text.restart, image.ZP, draw.Over)

	for {
		delay(1e5)
	}
}

func idraw(s *Square, m *image.RGBA) {
	draw.Draw(canvas, s.Rectangle, m, m.Bounds().Min, draw.Over)
}

func delay(ms time.Duration) {
	t := time.NewTicker(ms * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			return
		default:
			event('s')
			blit()
		}
	}
}

func ck(err error) {
	if err != nil {
		sdl.LogCritical(sdl.LOG_CATEGORY_APPLICATION, "%v", err)
		sdl.ShowSimpleMessageBox(sdl.MESSAGEBOX_ERROR, "Error", err.Error(), window)
		os.Exit(1)
	}
}

type Square struct {
	image.Rectangle
	Value int
	Warn  bool
	Seen  bool
}

type Grid struct {
	Squares [][]Square
	Mines   []image.Point
	Count   int
}

func newGrid(w, h, n int) *Grid {
	s := make([][]Square, h)
	for i := range s {
		s[i] = make([]Square, w)
		for j := range s[i] {
			s[i][j] = Square{
				Rectangle: image.Rect(j*30, i*30, (j+1)*30, (i+1)*30),
			}
		}
	}
	m := newMines(w, h, n)
	g := &Grid{
		Squares: s,
		Mines:   m,
		Count:   w*h - n,
	}

	for _, p := range m {
		c := g.Get(p.X, p.Y)
		c.Value = 9

		for i := -1; i <= 1; i++ {
			for j := -1; j <= 1; j++ {
				if i == 0 && j == 0 {
					continue
				}

				c := g.Get(p.X+j, p.Y+i)
				if c != nil && c.Value < 9 {
					c.Value++
				}
			}
		}
	}

	return g
}

func (g *Grid) Get(x, y int) *Square {
	if !(0 <= y && y < len(g.Squares)) || !(0 <= x && x < len(g.Squares[y])) {
		return nil
	}
	return &g.Squares[y][x]
}

func (g *Grid) Draw() {
	for y := range g.Squares {
		for _, s := range g.Squares[y] {
			draw.Draw(canvas, s.Rectangle, blue, blue.Bounds().Min, draw.Src)
		}
	}
}

func newMines(w, h, n int) []image.Point {
	m := make([]image.Point, n)
	p := rand.Perm(w * h)
	for i := range m {
		m[i] = image.Pt(p[i]%w, p[i]/w)
	}
	return m
}
