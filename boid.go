package main

import (
	"flag"
	"fmt"
)

var (
	separationFactor     = flag.Float64("separationFactor", 7.0, "separation factor") // Boid represents a boid in the simulation.
	separationPerception = 70.0
	alignmentFactor      = flag.Float64("alignmentFactor", 4.0, "alignment factor")
	alignmentPerception  = 250.0
	cohesionPerception   = 200.0
	cohesionFactor       = flag.Float64("cohesionFactor", 3.0, "cohesion factor")
	minVelocity          = 20.0
	maxVelocity          = 300.0
	centerAttraction     = flag.Float64("centerAttraction", 0.2, "center attraction")
)

type Boid struct {
	name         string
	sprite       int
	position     Vector
	velocity     Vector
	acceleration Vector
}

// New creates a new Boid instance.
func NewBoid(name string, sprite int, position, velocity, acceleration Vector) *Boid {
	return &Boid{
		name:         name,
		sprite:       sprite,
		position:     position,
		velocity:     velocity,
		acceleration: acceleration,
	}
}

// Move updates the boid's position based on its velocity and acceleration.
func (b *Boid) Move(dt float64) error {

	b.position = b.position.Add(b.velocity.Scaled(dt))
	b.velocity = b.velocity.Add(b.acceleration)
	// limit(&b.velocity, maxVelocity)
	b.MaintainVelocity()
	b.acceleration = Vector{}
	return nil
}

func (b *Boid) MaintainVelocity() {
	if b.velocity.Len() < minVelocity {
		b.velocity = b.velocity.Unit().Scaled(minVelocity)
	} else if b.velocity.Len() > maxVelocity {
		b.velocity = b.velocity.Unit().Scaled(maxVelocity)
	}
}

// Position returns the boid's position.
func (b *Boid) Position() Vector {
	return b.position
}

func (b *Boid) PositionXY() (float64, float64) {
	return b.position.X, b.position.Y
}

// Sprite returns the boid's sprite index.
func (b *Boid) Sprite() int {
	return b.sprite
}

// Velocity returns the boid's velocity.
func (b *Boid) Velocity() Vector {
	return b.velocity
}

// Acceleration returns the boid's acceleration.
func (b *Boid) Acceleration() Vector {
	return b.acceleration
}

// InfoString returns a formatted string containing simulation factors.
func InfoString() string {
	return fmt.Sprintf("A: %.2f\nC: %.2f\nS: %.2f", *alignmentFactor, *cohesionFactor, *separationFactor)
}

// SetFactors sets the simulation factors for a boid.
func SetFactors(factors [3]float64) {
	*alignmentFactor = factors[0]
	*cohesionFactor = factors[1]
	*separationFactor = factors[2]
}

// Limit vector's magnitude to the specified max.
func limit(u *Vector, max float64) {
	if u.Len() > max {
		*u = u.Unit().Scaled(max)
	}
}

// Rules apply separation, cohesion, and alignment rules to the boid.
func (b *Boid) Rules(boids []*Boid) {

	sepBoids := 0
	aliBoids := 0
	cohBoids := 0
	var d float64

	separationSteer := Vector{}
	alignmentSteer := Vector{}
	cohesionSteer := Vector{}

	for _, other := range boids {
		if b != other {
			diff := b.Position().Sub(other.Position())
			d = diff.Len()

			if d < separationPerception {
				sepBoids++
				diff = diff.Scaled(1 / (d * d))
				separationSteer = separationSteer.Add(diff)
			}

			if d < alignmentPerception {
				aliBoids++
				alignmentSteer = alignmentSteer.Add(other.Velocity())
			}

			if d < cohesionPerception {
				cohBoids++
				cohesionSteer = cohesionSteer.Add(other.position)
			}
		}
	}

	if sepBoids > 0 {
		separationSteer = separationSteer.Unit().Scaled(maxVelocity)
		separationSteer = separationSteer.Sub(b.Velocity())
		limit(&separationSteer, 1.25)
	}

	if aliBoids > 0 {
		alignmentSteer = alignmentSteer.Unit().Scaled(maxVelocity)
		alignmentSteer = alignmentSteer.Sub(b.Velocity())
		limit(&alignmentSteer, 1.0)
	}

	if cohBoids > 0 {
		cohesionSteer = cohesionSteer.Scaled(1 / float64(cohBoids))
		cohesionSteer = cohesionSteer.Sub(b.position)
		cohesionSteer = cohesionSteer.Unit().Scaled(maxVelocity)
		cohesionSteer = cohesionSteer.Sub(b.velocity)
		limit(&cohesionSteer, 0.85)
	}

	// Center attraction for boid b scaled with distance from center
	center := Vector{X: 0, Y: 0}
	b.acceleration = b.acceleration.Add(center.Sub(b.position).Scaled(*centerAttraction / 100))

	b.acceleration = b.acceleration.Add(separationSteer.Scaled(*separationFactor))
	b.acceleration = b.acceleration.Add(alignmentSteer.Scaled(*alignmentFactor))
	b.acceleration = b.acceleration.Add(cohesionSteer.Scaled(*cohesionFactor))
	b.acceleration = b.acceleration.Scaled(0.3)
}
