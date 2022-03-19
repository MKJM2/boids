# Design document for boid simulations

### Introduction

In this project, I implement a brute flocking algorithm inspired by Craig
Reynolds [article](https://www.red3d.com/cwr/boids/), using the Go programming
language and
[Pixel](https://github.com/faiface/pixel)
as my 2D rendering library of choice.

In the process of creating this project, I have had to learn the basics of the Go language, the Pixel library, as well as understand the theory behind the flocking algorithm to try and implement it on my own.

Some of the biggest challenges I have come across while developing my project were:

* representing Boid objects in a non-object oriented language like Go

    To overcome this problem, I decided to store Boids as type structs and define object methods
    as functions acting on those type structs.

* lack of a proper GUI library for Go and Pixel

    Due to a lack of time and necessary knowledge, I opted out of implementing my own GUI and decided
    to go with command-line arguments instead.

One of my biggest personal goals for this project was to make the canvas "infinite", i.e. don't try to contain the boids inside of a fixed sized rectangle, but let them roam freely. I managed to do so, as
well as allow the user to move the camera around and follow the boids/flocks around the screen.

## The project in-depth

### main.go

We first import all of the necessary packages, define the `Flock` struct to
store all of the boids being simulated, define a custom fragment shader (used to
improve the graphics) as well as a helper function `loadPicture` which takes a
filename string as an argument and loads a picture into memory.

Inside of the `run()` function, we first handle all of the command-line arguments using methods
from the `flag` package. This approach seemed easier and faster than writing my own solution to parse command-line flags.

We then setup the configuration of the window we're going to use to display
the simulation, as well as create the window itself by using Pixel. Fortunately, Pixel takes care
of most of the dirty work in creating a window.

We then proceed to initialize all of the different variables and structs we are going to use in our
simulation:
* we enable the fragment shader
oject requires (for execution and testing) hardware or software other than that offered by VS Code, be sure that the the staff are aware of and have approved your project’s 
* we load in the sprites used to represent our boids

* we initialize the boids themselves using a simple for-loop

* we randomly assing positions and velocities to spread the boids out across the entire screen

* we initialize variables used to control the camera movement and measuring FPS

Inside of `for !win.Closed()` (the game loop), we handle the actual simulation: the camera movements, the controls, drawing the objects and the text to screen and calculating the FPS.

The implementation of the camera movement resembles the panning mechanic used in
Google Maps or Apple Maps. When the user presses his mouse button, dragging his cursor allows him to
move the map around. The actual camera movement code was heavily inspired by [this](https://github.com/faiface/pixel/wiki/Pressing-keys-and-clicking-mouse#moving-the-camera) Pixel tutorial, but the actual tutorial code needed major refactoring to make it work with a mouse.

The `Draw()` and `Run()` functions are simlpy helper functions to draw the boids
(and the background) to the screen and actually apply the three rules of flocking to
all of the boids respectively.

### boid/boid.go

We store the boid as a type struct, with fields like its name (redundant field for now; was meant to store a boid's name, like Bob or Alice - when the user hovered their cursor on top of a boid, a pop-up would display their name, which I though would be a fun touch to the simulation), the index of the sprite the boid is using, the boid's position, velocity and acceleration.

Then, we define some variables used to control the simulation, like the maximal speed a boid is allowed to have.

What follows is a number of helper function:

* the `New()` function serves as a "constructor" for our boids (analogous to constructors in Object Oriented Programming)
* the `Move()` function updates the position of the void by its current velocity, updates the velocity
by the boid's current acceleration and resets the acceleration 
* the rest of the functions are some simple getter functions to allow the main control code in main.go to access the boids' variables.

The most important part of this file is the `Rules()` function which applies the three rules which boids have to follow so that the flocks get created: cohesion, alignment and separation. For an in depth explanation of my implementation refer to the in-code comments. In short, we use a single
for-loop to iterate over all of the boids and calculate the desired acceleration of a boid based on the separation, alignment and cohesion. We scale the respective forces by their weights (the factors) and calculate the final acceleration based on them.

### Miscellaneous
The project directory contains a couple additional files, like the actual boid sprite stored inside of the boid.png, the README.md and DESIGN.md and Go files required to build and run the project.

## Room for improvement

* optimizing the algorithm w/ quadtrees
* writing my own GUI library for Go

    implementing sliders to allow the user to dynamically change the values of the simulation