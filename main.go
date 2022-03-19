package main

import (
	"fmt"
	"flag"
	"image"
	"math"
	"math/rand"
	"os"
	"time"
	_ "image/png"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/colornames"
	"cs50-project/boid"
	"image/color"
)

//Struct to hold all the boids
type Flock struct {
	boids []*boid.Boid
}

//Custom fragment shader for improved visuals
var fragmentShader = `
#version 330 core

in vec2  vTexCoords;

out vec4 fragColor;

uniform vec4 uTexBounds;
uniform sampler2D uTexture;

void main() {
	// Current screen coordinates
	vec2 t = (vTexCoords - uTexBounds.xy) / uTexBounds.zw;

	// We sum the three color channels
	float sum  = texture(uTexture, t).r;
	      sum += texture(uTexture, t).g;
	      sum += texture(uTexture, t).b;

	// Calculate the average and set the output to the result
	vec4 color = vec4( sum/3, sum/3, sum/3, 1.0);

	//If the pixel was transparent, we keep it the same as it was before
	if (texture(uTexture, t).a < 1) {
		fragColor = texture(uTexture, t);
	} else {
		fragColor = color;
	}
}
`

//Helper function to load pictures from filesystem
func loadPicture(path string) (pixel.Picture, error) {
	//Open the file
        file, err := os.Open(path)
	//Handle errors
        if err != nil {
                return nil, err
        }
        defer file.Close()
	//Decode file
        img, _, err := image.Decode(file)
        if err != nil {
                return nil, err
        }
        return pixel.PictureDataFromImage(img), nil
}

func run() {

	//Setup command-line arguments using the "flag" Go module
	cohPtr := flag.Float64("c", 1.0, "the cohesion factor")
	sepPtr := flag.Float64("s", 1.5, "the separation factor")
	aliPtr := flag.Float64("a", 1.0, "the alignment factor")
	boidNumPtr := flag.Int("n", 300, "the amount of boids to simulate")

	helpPtr := flag.Bool("help", false, "display the help message")

	//Parse the command line arguments
	flag.Parse()

	if *helpPtr {
		//Print the help message and exit the program
		flag.PrintDefaults()
		return
	}

	//Display debug messages to the user console
	fmt.Println("Cohesion factor: ", *cohPtr)
	fmt.Println("Separation factor: ", *sepPtr)
	fmt.Println("Alignment factor: ", *aliPtr)

	//Configure the window
	cfg := pixelgl.WindowConfig{
		Title: "CS50 Boids",
		Bounds: pixel.R(0, 0, 1920, 1080),
		VSync: false,
		Monitor: pixelgl.PrimaryMonitor(),
	}

	//Create new window
	win, err := pixelgl.NewWindow(cfg)

	//Handle errors
	if err != nil {
		panic(err)
	}

	//Enable the shader
	win.Canvas().SetFragmentShader(fragmentShader)

	//Configure font
	face := basicfont.Face7x13
	if err != nil {
		panic(err)
	}

	//Create new Text object
	atlas := text.NewAtlas(face, text.ASCII)
	txt := text.New(pixel.V(0,0), atlas)


	//Set the desired color of the text
	txt.Color = colornames.Whitesmoke

	//Load sprites from file
	spritesheet, err := loadPicture("boidSprite.png")
	if err != nil {
		panic(err)
	}

	//Use the Pixel Batch for efficient drawing to screen
	batch := pixel.NewBatch(&pixel.TrianglesData{}, spritesheet)

	//Handle the spritesheet (in case of multiple 64x64 sprites being passed in)
	var boidFrames []pixel.Rect
        for x := spritesheet.Bounds().Min.X; x < spritesheet.Bounds().Max.X; x += 64 {

                for y := spritesheet.Bounds().Min.Y; y < spritesheet.Bounds().Max.Y; y += 64 {

                        boidFrames = append(boidFrames, pixel.R(x, y, x+64, y+64))
                }
        }

	//Create sprites and store them in the 'sprites' array
	sprites := make([]*pixel.Sprite, len(boidFrames))
	for i := 0; i < len(sprites); i++ {
		sprites[i] = pixel.NewSprite(spritesheet, boidFrames[i])
	}


	//Variables for controlling the camera movement
        var (
                camPos       = pixel.ZV
                camZoom      = 0.5
                camZoomSpeed = 1.2
		camMode	     = 0
        )

	//Monitoring the frames per second (fps)
        var (
                frames = 0
                second = time.Tick(time.Second)
        )

	//Number of boids (1000 by default)
	numBoids := *boidNumPtr


	//Create boids
	flock := Flock { make([]*boid.Boid, numBoids) }

	/*To set a random velocity for each boid, we get a random number in range [0..1) using rand.Float64()
	  and then map it to the range [min, max) using the formula:
	
	  randomNumber = min + rand.Float64() * (max - min)

	Credit to https://stackoverflow.com/questions/49746992/generate-random-float64-numbers-in-specific-range-using-golang
	*/


	var (
		winX = win.Bounds().Size().X
		winY = win.Bounds().Size().Y
	)

	for i := 0; i < numBoids; i++ {
		flock.boids[i] = boid.New("test",
		rand.Intn(len(boidFrames)),
		pixel.V(-winX/2 + rand.Float64()*winX, -winY/2 + rand.Float64()*winY), //position
		pixel.V(-100 + rand.Float64() * 200, -100 + rand.Float64() * 200), //velocity 
		pixel.V(0,0), //acceleration
		)
	}

	//Used for displaying the factors to the screen
	factors := [3]float64{*aliPtr, *cohPtr, *sepPtr}

	boid.SetFactors(factors)


	//Mouse anchor camera controls
	lastMousePos := pixel.ZV
	lastCamPos := pixel.ZV

        last := time.Now()
        for !win.Closed() {
		//Delta time (time it took to render the last frame)
                dt := time.Since(last).Seconds()
                last = time.Now()

		//Move the camera to it's current position inside of the world view (camPos)
                cam := pixel.IM.Scaled(camPos, camZoom).Moved(win.Bounds().Center().Sub(camPos))

		//Display only the cam view to the user
                win.SetMatrix(cam)

		//If user wants to exit the simulation
		if win.JustPressed(pixelgl.KeyEscape) {
			return
		}

		//Handle camera controls
		if win.JustPressed(pixelgl.MouseButtonLeft) {
			camMode = 1
			lastMousePos = win.MousePosition()
			lastCamPos = camPos
		}

		if win.JustReleased(pixelgl.MouseButtonLeft) {
			camMode = 0
		}

		//If user is holding his left mouse button down, we use the
		//mouse's position as an anchor point to move the camera around
		if camMode == 1 {

			//camPos = cam.Unproject(win.MousePosition())
			camPos = lastCamPos.Sub(win.MousePosition().Sub(lastMousePos).Scaled(1/camZoom))
		}


		//Zoom control with scroll wheel
                camZoom *= math.Pow(camZoomSpeed, win.MouseScroll().Y)

		//Draw the sprites and run the simulation
		Draw(batch, win, &flock, sprites, dt)
		Run(&flock, dt)


		//Draw the text to the window
		s := boid.InfoString()
		txt.WriteString(s)

		//Calculate the position of the text in the world view
		txtPos := pixel.V(-1000, 900)


		txt.Draw(win, pixel.IM.Scaled(txt.Orig, 6).Moved(txtPos))
		txt.Clear()

		//Update the window (Pixel built-in method)
		win.Update()

		//Calculate the fps and display it inside of the window title
		frames++
		select {
		case <-second:
			win.SetTitle(fmt.Sprintf("%s | FPS: %d", cfg.Title, frames))
			frames = 0
		default:
		}
	}
}

func Draw(batch *pixel.Batch, win *pixelgl.Window, flock *Flock, sprites []*pixel.Sprite, dt float64) {

	color := color.RGBA {7, 6, 25, 255}
	win.Clear(color)
	//win.Clear(colornames.Black)
	batch.Clear()

	for _, boid := range flock.boids {
		//boid.Draw(batch, pixel.IM.Scaled(pixel.ZV, 4).Moved(boid.Position()))
		//boid.Move()
		//fmt.Printf("Boid %d: X=%f, Y=%f\n", i, position.X, position.Y)	
		sprites[boid.Sprite()].Draw(batch, pixel.IM.Scaled(pixel.ZV, 1).Rotated(pixel.ZV, boid.Velocity().Angle() - math.Pi/2).Moved(boid.Position()))
	}


	batch.Draw(win)

}


func Run(flock *Flock, dt float64) {

	for _, boid := range flock.boids {

		//Check if on edge (optional since the camera movement is implemented already)

		//Check all the rules for this boid
		boid.Rules(flock.boids)


		//Move the boid accordingly
		boid.Move(dt)

	}

	//fmt.Printf("Boid 0's velocity: (%f, %f)\n", flock.boids[0].Velocity().X, flock.boids[0].Velocity().Y)
	//fmt.Printf("Boid 0's acceleration: (%f, %f)\n", flock.boids[0].Acceleration().X, flock.boids[0].Acceleration().Y)
}

func main() {
	//Seed the random number generator 	
	rand.Seed(time.Now().UnixNano())

	//Run Pixel
	pixelgl.Run(run)
}

