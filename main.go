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
	"math/rand"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	_ "github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct {
	camPos       Vector
	camSpeed     float64
	camZoom      float64
	camZoomSpeed float64
	camMode      int
	anchorPos    Vector
	lastCamPos   Vector
	matrices     []ebiten.GeoM
	cursor       ebiten.CursorShapeType
}

const (
	screenWidth  = 1920
	screenHeight = 1080
	boidWidth    = 64
	boidHeight   = 64
)

//go:embed boidsGlow64.png
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
			Vector{X: -float64(screenWidth/2) + rand.Float64()*float64(screenWidth),
				Y: -float64(screenHeight/2) + rand.Float64()*float64(screenHeight)}, // position
			Vector{X: -100 + rand.Float64()*200, Y: -100 + rand.Float64()*200}, // velocity
			Vector{X: 0, Y: 0}, // acceleration
		)
		unproject := g.cam()
		unproject.Invert()
		mx, my := flock.boids[i].Position().X, flock.boids[i].Position().Y
		ux, uy := unproject.Apply(float64(mx), float64(my))
		mat := ebiten.GeoM{}
		mat.Translate(-boidWidth/2, -boidHeight/2)
		mat.Translate(ux, uy)
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
		camZoomSpeed: 1.05,
		camMode:      0,
		anchorPos:    Vector{X: 0, Y: 0},
		lastCamPos:   Vector{X: 0, Y: 0},
		matrices:     []ebiten.GeoM{},
		cursor:       ebiten.CursorShapeDefault,
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

	// In ebitengine, the delta time is always 1/60
	// dt := 1.0 / 60

	for i, boid := range flock.boids {
		boid.Rules(flock.boids)
		boid.Move(0.01)
		// unproject := g.cam()
		unproject := ebiten.GeoM{}
		unproject.Invert()
		mx, my := boid.PositionXY()
		ux, uy := unproject.Apply(mx, my)
		mat := ebiten.GeoM{}
		// Stretch horizontally the faster the boid goes
		mat.Translate(-boidWidth/2, -boidHeight/2)
		mat.Scale(1-boid.Velocity().Len()/750, 1+boid.Velocity().Len()/550)
		mat.Rotate(boid.Velocity().Angle() + 1.5707963267948966)
		mat.Translate(ux, uy)
		g.matrices[i] = mat
	}

	/* Handle pan to move the camera, using mouse as anchor point */
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.camMode = 1
		mx, my := ebiten.CursorPosition()
		g.anchorPos = NewVector(float64(mx), float64(my))
		g.cursor = ebiten.CursorShapeMove
		ebiten.SetCursorShape(g.cursor)
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		g.camMode = 0
		g.cursor = ebiten.CursorShapeDefault
		ebiten.SetCursorShape(g.cursor)
		g.lastCamPos = g.camPos
	}

	if g.camMode == 1 {
		mx, my := ebiten.CursorPosition()
		dx, dy := float64(mx)-g.anchorPos.X, float64(my)-g.anchorPos.Y

		// g.camPos.X = -(float64(mx) - g.anchorPos.X) * 1 / g.camZoom
		// g.camPos.Y = -(float64(my) - g.anchorPos.Y) * 1 / g.camZoom
		g.camPos.X = g.lastCamPos.X - dx*1/g.camZoom
		g.camPos.Y = g.lastCamPos.Y - dy*1/g.camZoom
	}

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		*separationFactor -= 0.1
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		*separationFactor += 0.1
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		*alignmentFactor -= 0.1
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		*alignmentFactor += 0.1
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		unproject := g.cam()
		unproject.Invert()
		mx, my := ebiten.CursorPosition()
		ux, uy := unproject.Apply(float64(mx), float64(my))
		mat := ebiten.GeoM{}
		mat.Translate(-boidWidth/2, -boidHeight/2)
		mat.Translate(ux, uy)
		g.matrices = append(g.matrices, mat)
		boid := NewBoid(fmt.Sprintf("%d", len(flock.boids)),
			rand.Intn(len(sprites)),
			Vector{X: ux, Y: uy}, // position
			Vector{X: -100 + rand.Float64()*200, Y: -100 + rand.Float64()*200}, // velocity
			Vector{X: 0, Y: 0}, // acceleration
		)
		flock.boids = append(flock.boids, boid)
		*numBoids++
	}

	_, wheel_y := ebiten.Wheel()
	// g.camZoom *= math.Pow(g.camZoomSpeed, wheel_y)

	mx, my := ebiten.CursorPosition()

	// Get the camera transformation
	cam := g.cam()

	// Get the inverse of the camera transformation
	inverseCam := cam
	inverseCam.Invert()

	// Convert the screen coordinates to world coordinates
	wx, wy := inverseCam.Apply(float64(mx), float64(my))

	if ebiten.IsKeyPressed(ebiten.KeyQ) || wheel_y < 0 {
		g.camZoom *= (1 / g.camZoomSpeed)
		g.camPos.X -= (wx - g.camPos.X) * g.camZoomSpeed / 20
		g.camPos.Y -= (wy - g.camPos.Y) * g.camZoomSpeed / 20

		g.lastCamPos = g.camPos
	}

	if ebiten.IsKeyPressed(ebiten.KeyE) || wheel_y > 0 {
		g.camZoom *= g.camZoomSpeed
		g.camPos.X += (wx - g.camPos.X) * (1 / g.camZoomSpeed) / 20
		g.camPos.Y += (wy - g.camPos.Y) * (1 / g.camZoomSpeed) / 20

		g.lastCamPos = g.camPos
	}

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

	msg := fmt.Sprintf(`
			FPS: %0.2f
			TPS: %0.2f
			Cam X: %0.2f
			Cam Y: %0.2f
			Zoom: %0.5f
			Mode: %d
			A: %0.2f
			S: %0.2f
			C: %0.2f`, ebiten.ActualFPS(), ebiten.ActualTPS(), g.camPos.X, g.camPos.Y, g.camZoom, g.camMode, *alignmentFactor, *separationFactor, *cohesionFactor)
	ebitenutil.DebugPrint(screen, msg)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 1920, 1080
}
