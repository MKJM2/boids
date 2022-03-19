package boid

import (
	"fmt"
	"github.com/faiface/pixel"
)

type Boid struct {
	name		string
	sprite		int
	position	pixel.Vec
	velocity	pixel.Vec
	acceleration	pixel.Vec
}

var (
	separationFactor = 3.0
	separationPerception = 50.0
	alignmentFactor = 5.0
	alignmentPerception = 150.0
	cohesionPerception = 150.0
	cohesionFactor = 2.0
	maxVelocity = 200.0
)

func New(name string, sprite int, position pixel.Vec, velocity pixel.Vec, acceleration pixel.Vec) *Boid {
	b := Boid {name, sprite, position, velocity, acceleration}
	return &b
}

//Since everything is passed by value in Go, we use a pointer as the function argument
func (b *Boid) Move(dt float64) {
	b.position = b.position.Add(b.velocity.Scaled(dt))
	b.velocity = b.velocity.Add(b.acceleration)
	limit(&b.velocity, maxVelocity)
	//reset the acceleration (that is, set it to the Zero Vector)
	b.acceleration = pixel.ZV
}

/* Getter functions */

func (b *Boid) Position() pixel.Vec {
	return b.position
}

func (b *Boid) Sprite() int {
	return b.sprite
}

func (b *Boid) Velocity() pixel.Vec {
	return b.velocity
}


func (b *Boid) Acceleration() pixel.Vec {
	return b.acceleration
}

func InfoString() string {
	return fmt.Sprintf("A: %.2f\nC: %.2f\nS: %.2f", alignmentFactor, cohesionFactor, separationFactor)
}

/* --- */


/* Sets the simulation factors for a boid */
func SetFactors(factors [3]float64) {
	alignmentFactor = factors[0]
	cohesionFactor = factors[1]
	separationFactor = factors[2]
}



//Limit vector's magnitude to the specified max
func limit(u *pixel.Vec, max float64) {
	if u.Len() > max {
		*u = u.Unit().Scaled(max)
	}
}


//Apply separation, cohesion and alignment
func (b *Boid) Rules(boids []*Boid) {


	//Vars
	sepBoids := 0
	aliBoids := 0
	cohBoids := 0
	var d float64 = 0.0 //distance between two boids

	separationSteer := pixel.ZV //vector used for steering
	alignmentSteer := pixel.ZV
	cohesionSteer := pixel.ZV

        for _, other := range boids {
		if b != other {
			//Calculate the distance between boid b and the other boid
			diff := b.Position().Sub(other.Position())
			d = diff.Len()


			//Check if in range of boid b's perception and if yes, act accordingly
			if (d < separationPerception) {
				sepBoids++

				/*Since the separation force should get weaker with distance:
				we want its value to get smaller and smaller 
				when "d" gets larger and larger

				Dividing by "d" once simply resets the magnitude of vector "diff"
				to 1 (since d is the length of that vector), so we divide
				the vector diff by d^2 
				*/

				diff = diff.Scaled(1/(d*d))
				separationSteer = separationSteer.Add(diff)
			}

			if (d < alignmentPerception) {
				aliBoids++
				alignmentSteer = alignmentSteer.Add(other.Velocity())
			}

			if (d < cohesionPerception) {
				cohBoids++
				cohesionSteer = cohesionSteer.Add(other.position)
				//cohesionSteer.Add(other.position)
			}
		}
        }
	if sepBoids > 0 {
		//Set magnitude to max speed
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
		cohesionSteer = cohesionSteer.Scaled(1/float64(cohBoids))
		cohesionSteer = cohesionSteer.Sub(b.position)
		cohesionSteer = cohesionSteer.Unit().Scaled(maxVelocity)
		cohesionSteer = cohesionSteer.Sub(b.velocity)
		limit(&cohesionSteer, 0.85)
	}


	b.acceleration = b.acceleration.Add(separationSteer.Scaled(separationFactor))
	b.acceleration = b.acceleration.Add(alignmentSteer.Scaled(alignmentFactor))
	b.acceleration = b.acceleration.Add(cohesionSteer.Scaled(cohesionFactor))
	
	//We are taking the average of the three accelerations acting on our object
	b.acceleration = b.acceleration.Scaled(0.3)

}

