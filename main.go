package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	_ "fmt"
	"image"
	_ "image"
	"image/color"
	_ "image/color"
	_ "image/png"
	"log"
	"math"
	"math/rand"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	_ "github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct {
	camPos       Vector
	camSpeed     float64
	camZoom      float64
	camZoomSpeed float64
	matrices     []ebiten.GeoM
}

const (
	screenWidth  = 1920
	screenHeight = 1080
)

//go:embed boidsGlow.png
var b_boidImage []byte

var (
	flock      *Flock
	sprites    []*ebiten.Image
	boidImage  *ebiten.Image
	backbuffer *ebiten.Image
	numBoids   = flag.Int("boids", 100, "number of boids")
)

// Flock struct to hold all the boids
type Flock struct {
	boids []*Boid
}

func loadSprites() {
	img_decoded, _, err := image.Decode(bytes.NewReader(b_boidImage))
	if err != nil {
		log.Fatal(err)
	}
	boidImage = ebiten.NewImageFromImage(img_decoded)
}

func loadBoidSprites() {
	const boidWidth = 32
	const boidHeight = 32
	for y := 0; y < boidImage.Bounds().Max.Y; y += boidHeight {
		for x := 0; x < boidImage.Bounds().Max.X; x += boidWidth {
			subImg := boidImage.SubImage(image.Rect(x, y, x+boidWidth, y+boidHeight))
			sprites = append(sprites, ebiten.NewImageFromImage(subImg))
		}
	}
}

func (g *Game) initialize() {
	flock = &Flock{boids: make([]*Boid, *numBoids)}

	backbuffer = ebiten.NewImage(screenWidth, screenHeight)

	for i := range flock.boids {
		flock.boids[i] = NewBoid(fmt.Sprintf("%d", i),
			rand.Intn(len(sprites)),
			Vector{X: -float64(screenWidth/2) + rand.Float64()*float64(screenWidth), Y: -float64(screenHeight/2) + rand.Float64()*float64(screenHeight)}, // position
			Vector{X: -100 + rand.Float64()*200, Y: -100 + rand.Float64()*200},                                                                           // velocity
			Vector{X: 0, Y: 0}, // acceleration
		)
		unproject := g.cam()
		unproject.Invert()
		mx, my := flock.boids[i].Position().X, flock.boids[i].Position().Y
		ux, uy := unproject.Apply(float64(mx), float64(my))
		mat := ebiten.GeoM{}
		mat.Translate(ux, uy)
		mat.Translate(-32/2, -32/2)
		g.matrices = append(g.matrices, ebiten.GeoM{})
	}
}

func main() {
	flag.Parse()
	loadSprites()
	loadBoidSprites()

	g := Game{
		camSpeed:     300,
		camZoom:      1,
		camZoomSpeed: 1.2,
		matrices:     []ebiten.GeoM{},
	}

	g.initialize()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Boids")

	if err := ebiten.RunGame(&g); err != nil {
		panic(err)
	}
}

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		os.Exit(0)
	}

	// Delta time
	dt := 1.0 / ebiten.ActualTPS()

	for i, boid := range flock.boids {
		boid.Rules(flock.boids)
		boid.Move(0.01)
		// unproject := g.cam()
		unproject := ebiten.GeoM{}
		unproject.Invert()
		mx, my := boid.PositionXY()
		ux, uy := unproject.Apply(mx, my)
		mat := ebiten.GeoM{}
		mat.Rotate(boid.Velocity().Angle() + 1.5707963267948966)
		mat.Translate(ux, uy)
		mat.Translate(-32/2, -32/2)
		g.matrices[i] = mat
	}

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.camPos.X -= g.camSpeed * dt * 1 / g.camZoom
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.camPos.X += g.camSpeed * dt * 1 / g.camZoom
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.camPos.Y += g.camSpeed * dt * 1 / g.camZoom
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.camPos.Y -= g.camSpeed * dt * 1 / g.camZoom
	}
	_, wy := ebiten.Wheel()
	g.camZoom *= math.Pow(g.camZoomSpeed, wy)

	return nil
}

/** cam returns the camera matrix */
func (g *Game) cam() ebiten.GeoM {
	cam := ebiten.GeoM{}
	cam.Translate(-g.camPos.X, -g.camPos.Y)
	cam.Scale(g.camZoom, g.camZoom)
	cam.Translate(g.camPos.X, g.camPos.Y)
	cam.Translate(float64(float64(screenWidth)/2-g.camPos.X), float64(screenHeight)/2-g.camPos.Y)
	return cam
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{7, 6, 25, 255})

	for i, boid := range flock.boids {
		op := &ebiten.DrawImageOptions{}
		op.GeoM = g.matrices[i]
		op.GeoM.Concat(g.cam())
		screen.DrawImage(sprites[boid.Sprite()], op)
	}

	// msg := fmt.Sprintf("FPS: %0.2f\nCam X: %0.2f\nCam Y: %0.2f\nZoom: %0.5f\nMode: %d\nLast Mouse X: %0.2f\n Last Mouse Y: %0.2f\n", ebiten.ActualFPS())
	// ebitenutil.DebugPrint(screen, msg)

	msg := fmt.Sprintf("TPS: %0.2f\nCam X: %0.2f\nCam Y: %0.2f\nZoom: %0.5f\nMode: %d\nLast Mouse X: %0.2f\n Last Mouse Y: %0.2f\n", ebiten.ActualTPS(), g.camPos.X, g.camPos.Y, g.camZoom, 0, 0, 0)
	ebitenutil.DebugPrint(screen, msg)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 1920, 1080
}
